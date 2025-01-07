package main

import (
	"crypto/sha1"
	"sync"
	"time"
)

var CacheInstance *Cache
var once sync.Once

type trainData struct {
	value        any
	insertedTime time.Time
}
type Cache struct {
	value map[[20]byte]trainData
	ttl   time.Duration
}

func NewCache() *Cache {
	once.Do(func() {
		CacheInstance = &Cache{value: make(map[[20]byte]trainData), ttl: time.Duration(time.Minute) * CACHE_TTL}
	})
	return CacheInstance
}

func (c *Cache) Set(username, from, to string, value any) {
	key := sha1.Sum([]byte(username + from + to))
	c.value[key] = trainData{value: value, insertedTime: time.Now()}
	for v, td := range c.value {
		if time.Now().Sub(td.insertedTime) > c.ttl {
			delete(c.value, v)
		}
	}
}

func (c *Cache) Get(username, from, to string) any {
	key := sha1.Sum([]byte(username + from + to))
	if val, exists := c.value[key]; exists {
		return val.value
	}
	return nil
}
