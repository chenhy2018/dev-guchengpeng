package memcache

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qiniu/xlog.v1"
)

func setupMcd(xl *xlog.Logger, server string) bool {

	c, err := net.Dial("tcp", server)
	if err != nil {
		xl.Warn("skipping test; no server running at %s", server)
		return false
	}
	c.Write([]byte("flush_all\r\n"))
	c.Close()
	return true
}

type mcValue struct {
	A int
	B string
	C []int
	D []string
}

type keyValue struct {
	key string
	val mcValue
}

func TestMcService(t *testing.T) {

	xl := xlog.NewDummy()
	if !setupMcd(xl, "localhost:11211") {
		return
	}
	if !setupMcd(xl, "localhost:11212") {
		return
	}

	conns := []Conn{
		{[]string{"key1", "key2"}, "localhost:11211"},
		{[]string{"key3", "key4"}, "localhost:11212"},
	}
	client := New(conns)

	kvs := []keyValue{
		{"hello", mcValue{123, "abc", nil, nil}},
		{"world", mcValue{456, "", []int{12, 34}, []string{"aa", "bb"}}},
		{"hahaa", mcValue{0, "def", nil, []string{"dd"}}},
		{string(bytes.Repeat([]byte("A"), 251)), mcValue{A: 4}},
		{string(bytes.Repeat([]byte("B"), 251)), mcValue{A: 78}},
		{string(bytes.Repeat([]byte("C"), 251)), mcValue{A: 12}},
	}

	var err error
	for _, kv := range kvs {
		var ret mcValue
		err = client.Get(xl, kv.key, &ret)
		assert.Equal(t, err, ErrCacheMiss, "Get %v", kv.key)
		err = client.Set(xl, kv.key, &kv.val)
		assert.NoError(t, err, "Set %v", kv.key)
		err = client.Get(xl, kv.key, &ret)
		assert.NoError(t, err, "Get %v", kv.key)
		assert.Equal(t, ret, kv.val, "Get %v", kv.key)
	}

	for _, kv := range kvs[:2] {
		var ret mcValue
		err = client.Del(xl, kv.key)
		assert.Equal(t, err, nil)
		err = client.Get(xl, kv.key, &ret)
		assert.Equal(t, err, ErrCacheMiss, "Get %v", kv.key)
	}

	for _, kv := range kvs[2:] {
		var ret mcValue
		err = client.Get(xl, kv.key, &ret)
		assert.NoError(t, err, "Get %v", kv.key)
		assert.Equal(t, ret, kv.val, "Get %v", kv.key)
	}
}
