package localcache

import (
	"errors"
	"sync"
	"time"

	"github.com/qiniu/log.v1"
)

var ErrNoSuchEntry = errors.New("no such entry")

type cacheItem struct {
	lastModified int64 // unix nano time
	data         []byte
}

type MemCache struct {
	items map[string]*cacheItem

	expiredAfter time.Duration
	lock         sync.RWMutex
}

func NewMemCache(expiredAfter time.Duration) (*MemCache, error) {
	if expiredAfter < 0 {
		return nil, errors.New("negative expiredAfter")
	}
	c := MemCache{expiredAfter: expiredAfter, items: make(map[string]*cacheItem)}
	go c.cleanUp()
	return &c, nil
}

func (c *MemCache) Set(key string, data []byte) {
	item := cacheItem{
		lastModified: time.Now().UnixNano(),
		data:         data,
	}
	c.lock.Lock()
	// TODO: change to tmpfs
	c.items[key] = &item
	c.lock.Unlock()
	return
}

func (c *MemCache) Get(key string) ([]byte, error) {
	c.lock.RLock()
	item, ok := c.items[key]
	c.lock.RUnlock()
	if !ok {
		return nil, ErrNoSuchEntry
	}
	return item.data, nil
}

func (c *MemCache) Remove(key string) {
	c.lock.Lock()
	delete(c.items, key)
	c.lock.Unlock()
	return
}

func (c *MemCache) cleanUp() {
	if c.expiredAfter <= 0 {
		return
	}

	tick := time.NewTicker(c.expiredAfter / 10)
	for {
		now := <-tick.C
		expired := now.UnixNano() - int64(c.expiredAfter)
		c.lock.Lock()
		for key, item := range c.items {
			if item.lastModified < expired {
				log.Println("mc:", key, "expired")
				delete(c.items, key)
			}
		}
		c.lock.Unlock()
	}
}
