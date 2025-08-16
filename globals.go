package main

import "database/sql"

const (
	//база данных
	host_db     = "localhost"
	port_db     = 5432
	user_db     = "orders_user"
	password_db = "1"
	dbname_db   = "orders_db"

	//кэш
	cache_size = 100

	//кафка
	kafka_bootstrap_servers = "localhost:9092"
	kafka_group_id          = "orders-group"
	kafka_topic             = "orders-topic"
	kafka_auto_offset_reset = "earliest"
	kafka_fetch_min_bytes   = 1
	kafka_fetch_wait_max_ms = 500
)

var (
	GlobalCache *LRUCache

	db *sql.DB
)
