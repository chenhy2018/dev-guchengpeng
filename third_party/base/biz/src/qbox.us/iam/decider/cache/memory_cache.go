package cache

import (
	"sync"
)

var _ CacheStore = &MemoryCache{}

type MemoryCache struct {
	data map[uint32]*Item
	lock *sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[uint32]*Item),
		lock: &sync.RWMutex{},
	}
}

func (c *MemoryCache) Get(id uint32) (*Item, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, ok := c.data[id]
	return item, ok
}

func (c *MemoryCache) Set(id uint32, item *Item) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.data[id] = item
}
