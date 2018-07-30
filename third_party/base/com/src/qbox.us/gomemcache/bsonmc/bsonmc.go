package bsonmc

import (
	"github.com/bradfitz/gomemcache/memcache"
	"labix.org/v2/mgo/bson"
)

var (
	ErrCacheMiss = memcache.ErrCacheMiss
)

type Client struct {
	Conn *memcache.Client
}

func (r Client) Set(key string, val interface{}) (err error) {

	b, err := bson.Marshal(val)
	if err != nil {
		return
	}

	item := &memcache.Item{Key: key, Value: b}
	return r.Conn.Set(item)
}

func (r Client) Get(key string, val interface{}) (err error) {

	item, err := r.Conn.Get(key)
	if err != nil {
		return
	}

	return bson.Unmarshal(item.Value, val)
}

func (r Client) Delete(key string) (err error) {

	return r.Conn.Delete(key)
}

