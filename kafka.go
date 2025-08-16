package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func handleMessageKafka(message string) {
	var order OrderResponse
	err := json.Unmarshal([]byte(message), &order)
	if err != nil {
		log.Printf("ошибка парсинга json: %v", err)
		log.Printf("не распарсен заказ: %s", message)
		return
	}

	log.Printf("успешно распарсен заказ: %s", order.OrderUID)

	err = saveOrderToDB(order)
	if err != nil {
		log.Printf("ошибка сохранения в базу данных: %v", err)
		return
	}

	log.Printf("заказ %s успешно сохранен в базу данных", order.OrderUID)

}

func consumeKafka() {
	config := &kafka.ConfigMap{
		"bootstrap.servers": kafka_bootstrap_servers,
		"group.id":          kafka_group_id,
		"auto.offset.reset": kafka_auto_offset_reset,
		"fetch.min.bytes":   kafka_fetch_min_bytes,
		"fetch.wait.max.ms": kafka_fetch_wait_max_ms,
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		panic(fmt.Sprintf("не удалось создать потребителя: %v", err))
	}

	err = consumer.SubscribeTopics([]string{kafka_topic}, nil)
	if err != nil {
		panic(fmt.Sprintf("не удалось подписаться: %v", err))
	}

	fmt.Println("консьюмер kafka инициализирован")

	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			handleMessageKafka(string(msg.Value))
			fmt.Printf("получено сообщение: %s\n", string(msg.Value))
		} else {
			fmt.Printf("ошибка потребителя: %v (%v)\n", err, msg)
		}
	}
}
