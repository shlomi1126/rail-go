package cache

import (
	"crypto/sha1"
	"sync"
	"time"
)

const CACHE_TTL = 30 // Cache TTL in minutes

var CacheInstance *Cache
var once sync.Once

type trainData struct {
	value        any
	insertedTime time.Time
}

type Cache struct {
	value map[[20]byte]trainData
	ttl   time.Duration
	mutex sync.RWMutex
}

func NewCache() *Cache {
	once.Do(func() {
		CacheInstance = &Cache{
			value: make(map[[20]byte]trainData),
			ttl:   time.Duration(CACHE_TTL) * time.Second,
			mutex: sync.RWMutex{},
		}
	})
	return CacheInstance
}

func (c *Cache) Set(username, from, to string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	key := sha1.Sum([]byte(username + from + to))
	c.value[key] = trainData{value: value, insertedTime: time.Now()}
	
	// Clean up expired entries
	for v, td := range c.value {
		if time.Since(td.insertedTime) > c.ttl {
			delete(c.value, v)
		}
	}
}

func (c *Cache) Get(username, from, to string) any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	key := sha1.Sum([]byte(username + from + to))
	if val, exists := c.value[key]; exists {
		// Check if entry is still valid
		if time.Since(val.insertedTime) <= c.ttl {
			return val.value
		}
	}
	return nil
}
