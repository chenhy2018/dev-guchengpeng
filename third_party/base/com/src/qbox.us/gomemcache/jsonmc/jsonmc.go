package jsonmc

import (
	"encoding/json"
	"github.com/bradfitz/gomemcache/memcache"
)

var (
	ErrCacheMiss = memcache.ErrCacheMiss
)

type Client struct {
	Conn *memcache.Client
}

func (r Client) Set(key string, val interface{}) (err error) {

	b, err := json.Marshal(val)
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

	return json.Unmarshal(item.Value, val)
}

func (r Client) Delete(key string) (err error) {

	return r.Conn.Delete(key)
}

