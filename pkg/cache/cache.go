package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type MemoryCache struct {
	c *cache.Cache
}

func NewMemoryCache(defaultExpiration, cleanupInterval time.Duration) *MemoryCache {
	return &MemoryCache{
		c: cache.New(defaultExpiration, cleanupInterval),
	}
}

func (m *MemoryCache) Set(key string, value interface{}, duration time.Duration) {
	m.c.Set(key, value, duration)
}

func (m *MemoryCache) Get(key string) (interface{}, bool) {
	return m.c.Get(key)
}

func (m *MemoryCache) Delete(key string) {
	m.c.Delete(key)
}
