package tmdb

import (
	"encoding/json"
	"sync"
	"time"
)

type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

type Cache struct {
	items    map[string]CacheItem
	mu       sync.RWMutex
	duration time.Duration
}

func NewCache(duration time.Duration) *Cache {
	return &Cache{
		items:    make(map[string]CacheItem),
		duration: duration,
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		delete(c.items, key)
		return nil, false
	}

	return item.Data, true
}

func (c *Cache) Set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CacheItem{
		Data:      data,
		ExpiresAt: time.Now().Add(c.duration),
	}
}

func (c *Cache) MarshalAndSet(key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.Set(key, jsonData)
	return nil
}

func (c *Cache) GetAndUnmarshal(key string, target interface{}) bool {
	data, found := c.Get(key)
	if !found {
		return false
	}

	jsonData, ok := data.([]byte)
	if !ok {
		return false
	}

	err := json.Unmarshal(jsonData, target)
	return err == nil
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]CacheItem)
}
