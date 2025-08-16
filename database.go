package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host_db     = "localhost"
	port_db     = 5432
	user_db     = "orders_user"
	password_db = "1"
	dbname_db   = "orders_db"
)

var db *sql.DB

func connectToDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host_db, port_db, user_db, password_db, dbname_db)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func saveOrderToDB(order OrderResponse) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
        INSERT INTO orders (
            order_uid, track_number, entry, locale, internal_signature, 
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO UPDATE SET
            track_number = EXCLUDED.track_number,
            entry = EXCLUDED.entry,
            locale = EXCLUDED.locale,
            internal_signature = EXCLUDED.internal_signature,
            customer_id = EXCLUDED.customer_id,
            delivery_service = EXCLUDED.delivery_service,
            shardkey = EXCLUDED.shardkey,
            sm_id = EXCLUDED.sm_id,
            date_created = EXCLUDED.date_created,
            oof_shard = EXCLUDED.oof_shard`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard)

	if err != nil {
		return fmt.Errorf("не удалось вставить заказ: %v", err)
	}
	log.Printf("данные заказа сохранены для %s", order.OrderUID)

	_, err = tx.Exec(`
        INSERT INTO delivery (
            order_uid, name, phone, zip, city, address, region, email
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (order_uid) DO UPDATE SET
            name = EXCLUDED.name,
            phone = EXCLUDED.phone,
            zip = EXCLUDED.zip,
            city = EXCLUDED.city,
            address = EXCLUDED.address,
            region = EXCLUDED.region,
            email = EXCLUDED.email`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email)

	if err != nil {
		return fmt.Errorf("не удалось вставить доставку: %v", err)
	}
	log.Printf("данные доставки сохранены для %s", order.OrderUID)

	_, err = tx.Exec(`
        INSERT INTO payment (
            order_uid, transaction, request_id, currency, provider, amount, 
            payment_dt, bank, delivery_cost, goods_total, custom_fee
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO UPDATE SET
            transaction = EXCLUDED.transaction,
            request_id = EXCLUDED.request_id,
            currency = EXCLUDED.currency,
            provider = EXCLUDED.provider,
            amount = EXCLUDED.amount,
            payment_dt = EXCLUDED.payment_dt,
            bank = EXCLUDED.bank,
            delivery_cost = EXCLUDED.delivery_cost,
            goods_total = EXCLUDED.goods_total,
            custom_fee = EXCLUDED.custom_fee`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)

	if err != nil {
		return fmt.Errorf("не удалось вставить платеж: %v", err)
	}
	log.Printf("данные платежа сохранены для %s", order.OrderUID)

	_, err = tx.Exec("DELETE FROM items WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return fmt.Errorf("не удалось удалить старые товары: %v", err)
	}

	for i, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO items (
                order_uid, chrt_id, track_number, price, rid, name, 
                sale, size, total_price, nm_id, brand, status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price,
			item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status)

		if err != nil {
			return fmt.Errorf("не удалось вставить товар %d: %v", i, err)
		}
	}
	log.Printf("%d товаров сохранено для %s", len(order.Items), order.OrderUID)

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("не удалось зафиксировать транзакцию: %v", err)
	}

	return nil
}

func loadCacheFromDB() error {
	log.Println("загрузка кэша из базы данных...")

	rows, err := db.Query("SELECT order_uid FROM orders ORDER BY date_created DESC LIMIT 100")
	if err != nil {
		return fmt.Errorf("не удалось запросить заказы: %v", err)
	}
	defer rows.Close()

	var orderIDs []string
	for rows.Next() {
		var orderID string
		if err := rows.Scan(&orderID); err != nil {
			log.Printf("ошибка сканирования order_id: %v", err)
			continue
		}
		orderIDs = append(orderIDs, orderID)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("ошибка итерации заказов: %v", err)
	}

	log.Printf("найдено %d заказов для загрузки в кэш", len(orderIDs))

	for _, orderID := range orderIDs {
		order, err := getOrderFromDB(orderID)
		if err != nil {
			log.Printf("ошибка загрузки заказа %s: %v", orderID, err)
			continue
		}

		GlobalCache.Add(orderID, order)
	}

	log.Printf("кэш загружен с %d заказами", GlobalCache.Size())
	return nil
}

func getOrderFromDB(orderID string) (OrderResponse, error) {
	var order Order
	err := db.QueryRow("SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid = $1", orderID).
		Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		return OrderResponse{}, fmt.Errorf("не удалось запросить заказ: %v", err)
	}

	var delivery Delivery
	err = db.QueryRow("SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1", orderID).
		Scan(&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email)
	if err != nil {
		return OrderResponse{}, fmt.Errorf("не удалось запросить доставку: %v", err)
	}

	var payment Payment
	err = db.QueryRow("SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE order_uid = $1", orderID).
		Scan(&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee)
	if err != nil {
		return OrderResponse{}, fmt.Errorf("не удалось запросить платеж: %v", err)
	}

	rows, err := db.Query("SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1", orderID)
	if err != nil {
		return OrderResponse{}, fmt.Errorf("не удалось запросить товары: %v", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
		if err != nil {
			log.Printf("ошибка сканирования товара: %v", err)
			continue
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return OrderResponse{}, fmt.Errorf("ошибка итерации товаров: %v", err)
	}

	return OrderResponse{
		OrderUID:          order.OrderUID,
		TrackNumber:       order.TrackNumber,
		Entry:             order.Entry,
		Delivery:          delivery,
		Payment:           payment,
		Items:             items,
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerID:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		Shardkey:          order.Shardkey,
		SmID:              order.SmID,
		DateCreated:       order.DateCreated,
		OofShard:          order.OofShard,
	}, nil
}
