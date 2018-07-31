package cc

import (
	"container/list"
	"github.com/qiniu/log.v1"
	"sync"
)

// ----------------------------------------------------------
// Cache

type Cache interface {
	Get(key string) interface{}
	Set(key string, val interface{}) interface{}
	Check(key string) interface{}
	Stat() (missing int64, total int64)
}

// ----------------------------------------------------------
// SimpleCache

type simpleCacheItem struct {
	key string      // Key
	v   interface{} // Val
}

type SimpleCache struct {
	cache   map[string]*list.Element
	chunks  *list.List
	mutex   *sync.Mutex
	missing int64
	hit     int64
	limit   int
}

func NewSimpleCache(limit int) *SimpleCache {
	cache := make(map[string]*list.Element)
	return &SimpleCache{cache, list.New(), new(sync.Mutex), 0, 0, limit}
}

func (p *SimpleCache) Stat() (missing int64, total int64) {
	return p.missing, p.missing + p.hit
}

func (p *SimpleCache) Get(key string) interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if e, ok := p.cache[key]; ok {
		item := e.Value.(*simpleCacheItem)
		p.chunks.MoveToBack(e)
		p.hit++
		return item.v
	}
	p.missing++
	return nil
}

func (p *SimpleCache) Check(key string) interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if e, ok := p.cache[key]; ok {
		return e.Value.(*simpleCacheItem).v
	}
	return nil
}

func (p *SimpleCache) Set(key string, v interface{}) interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if _, ok := p.cache[key]; ok { // check exist first
		return v
	}
	var e *list.Element
	var item *simpleCacheItem
	var oldv interface{}
	log.Debug("SimpleCache.Set:", p.chunks.Len(), p.limit)
	if p.chunks.Len() >= p.limit {
		e = p.chunks.Front()
		item = e.Value.(*simpleCacheItem)
		oldv = item.v
		delete(p.cache, item.key)
		p.chunks.MoveToBack(e)
	} else {
		item = new(simpleCacheItem)
		oldv = nil
		e = p.chunks.PushBack(item)
	}
	item.key = key
	item.v = v
	p.cache[key] = e
	return oldv
}

// ----------------------------------------------------------
