package cc

import (
	"container/list"
	"sync"
)

// ----------------------------------------------------------
// IntCache

type IntCache interface {
	Get(key string) int32
	Set(key string, val int32) int32
	Check(key string) int32
	Stat() (missing int64, total, wtotal int64)
}

// ----------------------------------------------------------
// SimpleIntCache

type simpleIntCacheItem struct {
	key string // Key
	v   int32  // Val
}

type SimpleIntCache struct {
	cache   map[string]*list.Element
	chunks  *list.List
	mutex   *sync.Mutex
	missing int64
	hit     int64
	wtotal  int64
	limit   int
}

func NewSimpleIntCache(limit int) *SimpleIntCache {
	cache := make(map[string]*list.Element)
	return &SimpleIntCache{cache, list.New(), new(sync.Mutex), 0, 0, 0, limit}
}

func (p *SimpleIntCache) Stat() (missing int64, total, wtotal int64) {
	return p.missing, p.missing + p.hit, p.wtotal
}

func (p *SimpleIntCache) Get(key string) int32 {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if e, ok := p.cache[key]; ok {
		item := e.Value.(*simpleIntCacheItem)
		p.chunks.MoveToBack(e)
		p.hit++
		return item.v
	}
	p.missing++
	return -1
}

func (p *SimpleIntCache) Check(key string) int32 {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if e, ok := p.cache[key]; ok {
		return e.Value.(*simpleIntCacheItem).v
	}
	return -1
}

func (p *SimpleIntCache) Set(key string, v int32) int32 {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.wtotal++

	if _, ok := p.cache[key]; ok { // check exist first
		return v
	}
	var e *list.Element
	var item *simpleIntCacheItem
	var oldv int32
	if p.chunks.Len() >= p.limit {
		e = p.chunks.Front()
		item = e.Value.(*simpleIntCacheItem)
		oldv = item.v
		delete(p.cache, item.key)
		p.chunks.MoveToBack(e)
	} else {
		item = new(simpleIntCacheItem)
		oldv = -1
		e = p.chunks.PushBack(item)
	}
	item.key = key
	item.v = v
	p.cache[key] = e
	return oldv
}

// ----------------------------------------------------------
