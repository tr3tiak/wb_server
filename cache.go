package main

import (
	"container/list"
	"log"
	"sync"
)

type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mutex    sync.RWMutex
}

type CacheItem struct {
	key   string
	value OrderResponse
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

func (lru *LRUCache) Get(key string) (OrderResponse, bool) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element, exists := lru.cache[key]; exists {
		lru.list.MoveToFront(element)

		cacheItem := element.Value.(*CacheItem)
		log.Printf("кэш hit для заказа: %s", key)
		return cacheItem.value, true
	}

	log.Printf("кэш miss для заказа: %s", key)
	return OrderResponse{}, false
}

func (lru *LRUCache) Add(key string, value OrderResponse) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element, exists := lru.cache[key]; exists {
		cacheItem := element.Value.(*CacheItem)
		cacheItem.value = value
		lru.list.MoveToFront(element)
		log.Printf("кэш обновлен для заказа: %s", key)
		return
	}

	cacheItem := &CacheItem{
		key:   key,
		value: value,
	}

	element := lru.list.PushFront(cacheItem)
	lru.cache[key] = element

	log.Printf("кэш добавлен для заказа: %s", key)

	if lru.list.Len() > lru.capacity {
		lru.removeOldest()
	}

	log.Printf("размер кэша: %d/%d", lru.list.Len(), lru.capacity)
}

func (lru *LRUCache) removeOldest() {
	oldest := lru.list.Back()
	if oldest != nil {
		lru.removeElement(oldest)
	}
}

func (lru *LRUCache) removeElement(element *list.Element) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()
	lru.list.Remove(element)
	delete(lru.cache, element.Value.(*CacheItem).key)
}

func (lru *LRUCache) Size() int {
	lru.mutex.RLock()
	defer lru.mutex.RUnlock()
	return lru.list.Len()
}

func (lru *LRUCache) Clear() {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	lru.list.Init()
	lru.cache = make(map[string]*list.Element)
	log.Println("кэш очищен")
}

func (lru *LRUCache) Keys() []string {
	lru.mutex.RLock()
	defer lru.mutex.RUnlock()

	keys := make([]string, 0, len(lru.cache))
	for element := lru.list.Front(); element != nil; element = element.Next() {
		cacheItem := element.Value.(*CacheItem)
		keys = append(keys, cacheItem.key)
	}
	return keys
}
