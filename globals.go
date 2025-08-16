package main

import "database/sql"

const (
	host_db     = "localhost"
	port_db     = 5432
	user_db     = "orders_user"
	password_db = "1"
	dbname_db   = "orders_db"
	cache_size  = 100
)

var (
	GlobalCache *LRUCache

	db *sql.DB
)
