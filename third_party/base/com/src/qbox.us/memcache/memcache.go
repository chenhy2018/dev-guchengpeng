package memcache

import (
	"container/list"
	"fmt"
	"qbox.us/cc/time"
	"sync"
)

// ----------------------------------------------------------

type Service struct {
	cache   map[string]*list.Element
	data    *list.List
	mutex   sync.Mutex
	missing int64
	hit     int64
	wtotal  int64
	limit   int
}

func New(limit int) *Service {
	cache := make(map[string]*list.Element)
	data := list.New()
	return &Service{
		cache: cache,
		data:  data,
		limit: limit,
	}
}

// ----------------------------------------------------------

func (p *Service) Stat() (missing int64, total, wtotal int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.missing, p.missing + p.hit, p.wtotal
}

func (p *Service) StatTxt() string {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	format := "missing:%v total:%v wtotal:%v\n"
	return fmt.Sprintf(format, p.missing, p.missing+p.hit, p.wtotal)
}

func (p *Service) ClearStat() {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.hit = int64(0)
	p.wtotal = int64(0)
	p.missing = int64(0)
}

// ----------------------------------------------------------

type cacheItem struct {
	key      string      // Key
	v        interface{} // Val
	deadline int64
	putTime  int64
}

func (p *Service) Check(key string) interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if e, ok := p.cache[key]; ok {
		item := e.Value.(*cacheItem)
		if time.Nanoseconds() <= item.deadline {
			return item.v
		}
	}
	return nil
}

func (p *Service) Get(key string) (v interface{}, ok bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if e, ok1 := p.cache[key]; ok1 {
		p.hit++
		item := e.Value.(*cacheItem)
		v = item.v
		if ok = (time.Nanoseconds() <= item.deadline); ok {
			p.data.MoveToBack(e)
		}
		return
	}
	p.missing++
	return
}

func (p *Service) Get2(key string, expireTime int64) (v interface{}, ok bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if e, ok1 := p.cache[key]; ok1 {
		p.hit++
		item := e.Value.(*cacheItem)
		v = item.v
		if ok = (time.Nanoseconds() <= item.deadline && item.putTime > expireTime); ok {
			p.data.MoveToBack(e)
		}
		return
	}
	p.missing++
	return
}

func (p *Service) Set(key string, v interface{}, expries int64) (oldv interface{}) {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.wtotal++

	if e, ok := p.cache[key]; ok { // check exist first
		item := e.Value.(*cacheItem)
		oldv = item.v
		item.v = v
		item.putTime = time.Nanoseconds()
		item.deadline = item.putTime + expries
		p.data.MoveToBack(e)
		return
	}

	var e *list.Element
	var item *cacheItem
	if p.data.Len() >= p.limit {
		e = p.data.Front()
		item = e.Value.(*cacheItem)
		oldv = item.v
		delete(p.cache, item.key)
		p.data.MoveToBack(e)
	} else {
		item = new(cacheItem)
		oldv = nil
		e = p.data.PushBack(item)
	}
	item.key = key
	item.v = v
	item.putTime = time.Nanoseconds()
	item.deadline = item.putTime + expries
	p.cache[key] = e
	return
}

func (p *Service) Del(key string) (oldv interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if item, ok := p.cache[key]; ok {
		oldv = p.data.Remove(item)
		delete(p.cache, key)
	}
	return
}

// ----------------------------------------------------------
