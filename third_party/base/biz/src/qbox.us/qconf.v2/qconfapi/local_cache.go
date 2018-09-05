package qconfapi

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/xlog.v1"
	"labix.org/v2/mgo/bson"
	"qbox.us/limit"
	"qbox.us/limit/keycount"
)

// ------------------------------------------------------------------------

const defaultChanBufSize = 1024 * 16

type cacheItem struct {
	ret interface{}
	val []byte
	t   int64
}

type localCache struct {
	cache         map[string]*cacheItem
	expires       int64
	duration      int64
	hit           int64
	missing       int64
	mutex         sync.RWMutex
	updateChan    chan UpdateItem
	updateFn      func(log *xlog.Logger, id string, cacheFlags int) (v []byte, exp bool, err error)
	updateBlocker limit.Limit
}

func newLocalCache(
	expires int64, duration int64, chanBufSize int,
	updateFn func(log *xlog.Logger, id string, cacheFlags int) (v []byte, exp bool, err error)) *localCache {

	if chanBufSize == 0 {
		chanBufSize = defaultChanBufSize
	}
	lc := &localCache{
		cache:         make(map[string]*cacheItem),
		expires:       expires,
		duration:      duration,
		updateChan:    make(chan UpdateItem, chanBufSize),
		updateFn:      updateFn,
		updateBlocker: keycount.NewBlockingKeyCountLimit(1),
	}
	go lc.routine()
	return lc
}

func (p *localCache) routine() {

	duration := p.duration
	cache := p.cache
	updateFn := p.updateFn
	log := xlog.NewDummy()

	for req := range p.updateChan {
		p.mutex.RLock()
		item, ok := cache[req.Id]
		p.mutex.RUnlock()
		now := time.Now().UnixNano()
		if ok && now-item.t <= duration {
			continue
		}
		v, exp, err := updateFn(log, req.Id, req.CacheFlags)
		if err != nil {
			log.Warn("qconf.localCache: update failed -", req.Id, req.CacheFlags, err)
			continue
		}
		if !exp {
			p.mutex.Lock()
			cache[req.Id] = &cacheItem{ret: nil, val: v, t: now}
			p.mutex.Unlock()
		}

	}
}

func (p *localCache) deleteItemSafe(id string) {
	p.mutex.Lock()
	delete(p.cache, id)
	p.mutex.Unlock()
}

var unitTestFunc func()

func (p *localCache) getFromLc(log *xlog.Logger,
	ret interface{}, id string, cacheFlags int, now int64) (ok, exp bool, err error) {

	p.mutex.RLock()
	item, ok := p.cache[id]

	if ok {
		if len(item.val) > 0 {
			p.mutex.RUnlock()
			if unitTestFunc != nil {
				unitTestFunc()
			}
			p.mutex.Lock()
			if len(item.val) > 0 {
				err = bson.Unmarshal(item.val, ret)
				if err != nil {
					p.mutex.Unlock()
					log.Error("qconf.Get: bson.Unmarshal failed -", err)
					return
				}
				if item2, ok2 := p.cache[id]; ok2 && item2 == item {
					item.ret = ret
					item.val = nil
				}
			} else {
				v := reflect.Indirect(reflect.ValueOf(ret))
				cache := reflect.Indirect(reflect.ValueOf(item.ret))
				v.Set(cache)
			}
			p.mutex.Unlock()
		} else {
			v := reflect.Indirect(reflect.ValueOf(ret))
			cache := reflect.Indirect(reflect.ValueOf(item.ret))
			v.Set(cache)
			p.mutex.RUnlock()
		}
		d := now - item.t
		if d <= p.expires {
			atomic.AddInt64(&p.hit, 1)
			if d > p.duration {
				p.updateItem(log, id, cacheFlags)
			}
			return
		}
		exp = true
	} else {
		p.mutex.RUnlock()
	}
	return
}

func (p *localCache) updateItem(log *xlog.Logger, id string, cacheFlags int) {

	select {
	case p.updateChan <- UpdateItem{id, cacheFlags}:
	default:
		log.Warn("qconf.localCache: updateChan full, skipped -", id, cacheFlags)
	}
}

func (p *localCache) get(log *xlog.Logger, ret interface{}, id string, cacheFlags int, now int64) (err error) {

	ok, exp, err := p.getFromLc(log, ret, id, cacheFlags, now)
	if err != nil {
		return err
	}
	if ok && !exp {
		return nil
	}

	key := []byte(id)
	p.updateBlocker.Acquire(key)
	defer p.updateBlocker.Release(key)

	ok, exp, err = p.getFromLc(log, ret, id, cacheFlags, now)
	if err != nil {
		return err
	}
	if ok && !exp {
		return nil
	}

	atomic.AddInt64(&p.missing, 1)

	v, exp, err := p.updateFn(log, id, cacheFlags)
	if err != nil {
		if ok && httputil.DetectCode(err)/100 == 5 { // 服务故障?
			log.Error("qconf.localCache: update failed -", id, cacheFlags, err)
			return nil
		}
		log.Warn("qconf.localCache: update failed -", id, cacheFlags, err)
		return
	}
	if !exp {
		err = bson.Unmarshal(v, ret)
		if err != nil {
			log.Error("qconf.Get: bson.Unmarshal failed -", err)
			return
		}

		p.mutex.Lock()
		p.cache[id] = &cacheItem{ret: ret, t: now}
		p.mutex.Unlock()
	}

	return
}

func (p *localCache) Stat() (hit int64, missing int64) {

	hit = atomic.LoadInt64(&p.hit)
	missing = atomic.LoadInt64(&p.missing)
	return
}

// ------------------------------------------------------------------------
