package configuration

import (
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// ------------------------------------------------------------------------------------------

type Refresher interface {
	Refresh(l rpc.Logger, id string) (err error)
}

// ------------------------------------------------------------------------------------------

type GroupItem struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type entry struct {
	Grp string `bson:"grp"`
	Key string `bson:"key"`
	Val string `bson:"val"`
}

// ------------------------------------------------------------------------------------------

type M bson.M

type Instance struct {
	*mgo.Collection
	cache     *memInstance
	Refresher Refresher
}

func New(c *mgo.Collection, refresher Refresher) (p *Instance, err error) {

	index := mgo.Index{
		Key:    []string{"grp", "key"},
		Unique: true,
	}
	err = c.EnsureIndex(index)
	if err != nil {
		return
	}

	cache := newMemInstance()
	p = &Instance{c, cache, refresher}

	iter := c.Find(M{}).Iter()
	var e entry
	for iter.Next(&e) {
		g := cache.dirtyRequire(e.Grp)
		g.data[e.Key] = e.Val
	}
	err = iter.Err()
	return
}

func (p *Instance) Put(grp, key, val string) (err error) {

	g := p.cache.Lock(grp)
	_, err = p.Upsert(M{"grp": grp, "key": key}, M{"grp": grp, "key": key, "val": val})
	if err == nil {
		g.data[key] = val
	}
	g.Unlock()

	if err == nil {
		p.refresh(grp, key)
	}
	return
}

func (p *Instance) Get(grp, key string) (val string, err error) {

	g := p.cache.Lock(grp)
	val, ok := g.data[key]
	g.Unlock()

	if ok {
		return
	}
	err = mgo.ErrNotFound
	return
}

func (p *Instance) Group(grp string) (items []GroupItem, err error) {

	g := p.cache.Lock(grp)
	items, err = g.Group()
	g.Unlock()

	return
}

func (p *Instance) Delete(grp, key string) (err error) {

	g := p.cache.Lock(grp)
	err = p.Remove(M{"grp": grp, "key": key})
	if err == nil {
		delete(g.data, key)
	}
	g.Unlock()

	if err == nil {
		p.refresh(grp, key)
	}
	return
}

func (p *Instance) refresh(grp, key string) {

	if refresher := p.Refresher; refresher != nil {
		if err := refresher.Refresh(nil, "grp:"+grp); err != nil {
			log.Warn("refresh failed id:", "grp:"+grp)
		}
		if err := refresher.Refresh(nil, "uc:"+grp+":"+key); err != nil {
			log.Warn("refresh failed id:", "uc:"+grp+":"+key)
		}
		if err := refresher.Refresh(nil, "bucketInfo:"+key); err != nil {
			log.Warn("refresh failed id:", "bucketInfo:"+key)
		}
	}
}

// ------------------------------------------------------------------------------------------
