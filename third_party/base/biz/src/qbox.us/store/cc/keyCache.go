package cc

import (
	"container/list"
	"sync"
)

// ----------------------------------------------------------
// KeyCache

type KeyCache interface {
	Get(key string) string
	Set(key string, size int64) string
	Check(key string) string
	Stat() (missing int64, total, wtotal int64)
}

// ----------------------------------------------------------
// SimpleKeyCache

type simpleKeyCacheItem struct {
	key  string // Key
	size int64  // Size
}

type SimpleKeyCache struct {
	cache      map[string]*list.Element
	chunks     *list.List
	mutex      sync.Mutex
	missing    int64
	hit        int64
	wtotal     int64
	space      int64
	count      int
	outOfLimit func(space int64, count int) bool
}

func NewSimpleKeyCache(outOfLimit func(space int64, count int) bool) *SimpleKeyCache {
	cache := make(map[string]*list.Element)
	return &SimpleKeyCache{cache, list.New(), sync.Mutex{}, 0, 0, 0, 0, 0, outOfLimit}
}

func (p *SimpleKeyCache) Stat() (missing, total, wtotal int64) {
	return p.missing, p.missing + p.hit, p.wtotal
}

func (p *SimpleKeyCache) DirtyGet(key string) string {
	if e, ok := p.cache[key]; ok {
		item := e.Value.(*simpleKeyCacheItem)
		p.chunks.MoveToBack(e)
		p.hit++
		return item.key
	}
	p.missing++
	return ""
}

func (p *SimpleKeyCache) Get(key string) string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.DirtyGet(key)
}

func (p *SimpleKeyCache) Check(key string) string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if e, ok := p.cache[key]; ok {
		return e.Value.(*simpleKeyCacheItem).key
	}
	return ""
}

func (p *SimpleKeyCache) DirtySet(key string, size int64) string {

	p.wtotal++

	if item, ok := p.cache[key]; ok { // check exist first
		p.space -= item.Value.(*simpleKeyCacheItem).size
		p.space += size
		item.Value.(*simpleKeyCacheItem).size = size
		return ""
	}
	var e *list.Element
	var item *simpleKeyCacheItem
	var oldKey string
	if p.outOfLimit(p.space, p.count) {
		e = p.chunks.Front()
		if e != nil {
			item = e.Value.(*simpleKeyCacheItem)
			oldKey = item.key
			delete(p.cache, item.key)
			p.space -= item.size
			p.chunks.MoveToBack(e)
		}
	}
	if item == nil {
		item = new(simpleKeyCacheItem)
		oldKey = ""
		e = p.chunks.PushBack(item)
		p.count += 1
	}
	item.key = key
	item.size = size
	p.space += size
	p.cache[key] = e
	return oldKey
}

func (p *SimpleKeyCache) Set(key string, size int64) string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.DirtySet(key, size)
}

func (p *SimpleKeyCache) DirtyDelete(key string) string {

	if item, ok := p.cache[key]; ok {
		p.space -= item.Value.(*simpleKeyCacheItem).size
		delete(p.cache, key)
		p.chunks.Remove(item)
		p.count -= 1
		return key
	}
	return ""
}

// ----------------------------------------------------------
