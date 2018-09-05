package api

import (
	"net"
	"testing"

	master "qbox.us/ebdmaster/api"
	"qbox.us/memcache/mcg"

	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func setupMcd(xl *xlog.Logger, server string) bool {

	c, err := net.Dial("tcp", server)
	if err != nil {
		xl.Warnf("skipping test; no server running at %s", server)
		return false
	}
	c.Write([]byte("flush_all\r\n"))
	c.Close()
	return true
}

type keyValue struct {
	key uint64
	val master.FileInfo
}

func TestMcService(t *testing.T) {

	xl := xlog.NewDummy()
	if !setupMcd(xl, "localhost:11211") {
		return
	}
	if !setupMcd(xl, "localhost:11212") {
		return
	}

	conns := []mcg.Node{
		{[]string{"key1", "key2"}, "localhost:11211"},
		{[]string{"key3", "key4"}, "localhost:11212"},
	}
	client, err := newMemcachedClient(conns, 0)
	if err != nil {
		t.Fatal(err)
	}

	kvs := []keyValue{
		{1, master.FileInfo{1, 2, []uint64{3}, []uint64{55}}},
		{4, master.FileInfo{2, 4, []uint64{4}, []uint64{56}}},
		{1<<64 - 1, master.FileInfo{3, 6, []uint64{5}, []uint64{57}}},
	}

	for _, kv := range kvs {
		var ret *master.FileInfo
		ret, err = client.Get(xl, kv.key)
		assert.Equal(t, err, memcache.ErrCacheMiss, "Get %v", kv.key)
		err = client.Set(xl, kv.key, &kv.val)
		assert.NoError(t, err, "Set %v", kv.key)
		ret, err = client.Get(xl, kv.key)
		assert.NoError(t, err, "Get %v", kv.key)
		assert.Equal(t, *ret, kv.val, "Get %v", kv.key)
	}

	client, err = newMemcachedClient(conns, 2)
	if err != nil {
		t.Fatal(err)
	}

	for _, kv := range kvs {
		var ret *master.FileInfo
		err = client.Set(xl, kv.key, &kv.val)
		assert.NoError(t, err, "Set %v", kv.key)
		ret, err = client.Get(xl, kv.key)
		assert.NoError(t, err, "Get %v", kv.key)
		assert.Equal(t, *ret, kv.val, "Get %v", kv.key)
	}
	time.Sleep(2e9)
	for _, kv := range kvs {
		_, err = client.Get(xl, kv.key)
		assert.Equal(t, err, memcache.ErrCacheMiss, "Get %v", kv.key)
	}
}
