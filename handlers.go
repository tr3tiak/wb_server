package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func orderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		http.Error(w, "Отсутствует параметр id", http.StatusBadRequest)
		return
	}

	if cachedOrder, found := GlobalCache.Get(orderID); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cachedOrder)
		return
	}

	orderFromDB, err := getOrderFromDB(orderID)
	if err != nil {
		log.Printf("ошибка получения заказа из бд: %v", err)
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	GlobalCache.Add(orderID, orderFromDB)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orderFromDB)
}
