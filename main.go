package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	GlobalCache = NewLRUCache(cache_size)

	var err error
	db, err = connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := loadCacheFromDB(); err != nil {
		log.Printf("предупреждение: не удалось загрузить кэш из базы данных: %v", err)
	}

	go consumeKafka()
	http.HandleFunc("/order", orderHandler)
	fmt.Println("сервер запущен на порту :8080")
	http.ListenAndServe(":8080", nil)
}
