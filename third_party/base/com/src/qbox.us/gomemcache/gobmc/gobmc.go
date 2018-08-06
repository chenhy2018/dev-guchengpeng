package gobmc

import (
	"bytes"
	"encoding/gob"
	"github.com/bradfitz/gomemcache/memcache"
)

var (
	ErrCacheMiss = memcache.ErrCacheMiss
)

type Client struct {
	Conn *memcache.Client
}

func (r Client) Set(key string, val interface{}) (err error) {

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err = encoder.Encode(val)
	if err != nil {
		return
	}

	item := &memcache.Item{Key: key, Value: buffer.Bytes()}
	return r.Conn.Set(item)
}

func (r Client) Get(key string, val interface{}) (err error) {

	item, err := r.Conn.Get(key)
	if err != nil {
		return
	}

	reader := bytes.NewReader(item.Value)
	decoder := gob.NewDecoder(reader)
	return decoder.Decode(val)
}

func (r Client) Delete(key string) (err error) {

	return r.Conn.Delete(key)
}

