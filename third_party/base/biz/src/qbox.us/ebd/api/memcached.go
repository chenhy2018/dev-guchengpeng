package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	master "qbox.us/ebdmaster/api"
	"qbox.us/memcache/mcg"
)

type memcachedClient struct {
	*memcache.Client
	McExpires int32
}

func newMemcachedClient(conns []mcg.Node, mcExpires int32) (*memcachedClient, error) {

	c, err := mcg.New(conns)
	if err != nil {
		return nil, err
	}
	return &memcachedClient{
		Client:    c,
		McExpires: mcExpires,
	}, nil
}

// ----------------------------------------------------------

var errMc404 = errors.New("404")
var errMc500 = errors.New("500")
var errMcBad = errors.New("bad")

func (self *memcachedClient) Set(l rpc.Logger, fid uint64, fi *master.FileInfo) error {

	xl := xlog.NewWith(l)

	var xerr error
	defer xl.Xtrack("mc.s", time.Now(), &xerr)

	key := strconv.FormatUint(fid, 36)
	b, err := json.Marshal(fi)
	if err != nil {
		xerr = errMcBad
		return err
	}
	item := &memcache.Item{
		Key:        key,
		Value:      b,
		Expiration: self.McExpires,
	}
	err = self.Client.Set(item)
	xl.Debugf("memcached.Set, key; %v, value: %v, err: %v\n", item.Key, string(item.Value), err)
	if err != nil {
		xerr = errMc500
		return err
	}
	return nil
}

func (self *memcachedClient) Get(l rpc.Logger, fid uint64) (*master.FileInfo, error) {

	xl := xlog.NewWith(l)

	var xerr error
	defer xl.Xtrack("mc.g", time.Now(), &xerr)

	key := strconv.FormatUint(fid, 36)
	item, err := self.Client.Get(key)
	if err != nil {
		xl.Debugf("memcached.Get, key; %v, err: %v\n", key, err)
		if err == memcache.ErrCacheMiss {
			xerr = errMc404
			return nil, err
		}
		xerr = errMc500
		return nil, err
	}
	xl.Debugf("memcached.Get, key; %v, value: %v, err: %v\n", key, string(item.Value), err)

	ret := new(master.FileInfo)
	err = json.NewDecoder(bytes.NewReader(item.Value)).Decode(ret)
	if err != nil {
		xerr = errMcBad
		return nil, err
	}
	return ret, nil
}

type nilMemcached struct {
}

func (self nilMemcached) Set(l rpc.Logger, fid uint64, fi *master.FileInfo) error {
	return nil
}

func (self nilMemcached) Get(l rpc.Logger, fid uint64) (*master.FileInfo, error) {
	return nil, memcache.ErrCacheMiss
}
