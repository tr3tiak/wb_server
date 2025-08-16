package main

var (
	GlobalCache *LRUCache

	GlobalConfig = struct {
		CacheSize int
		Debug     bool
	}{
		CacheSize: 100,
		Debug:     true,
	}

	GlobalCounter int
)
