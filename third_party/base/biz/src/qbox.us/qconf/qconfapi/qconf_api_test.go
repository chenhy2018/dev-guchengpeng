package qconfapi

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"launchpad.net/mgo/bson"
)

func skipIfNoMc(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:11211")
	if err != nil {
		log.Println("skipped")
		t.Skip("no memcached available")
	}
	conn.Close()
}

func isMcAlive() bool {
	conn, err := net.Dial("tcp", "127.0.0.1:11211")
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func TestGetFromMcache(t *testing.T) {
	skipIfNoMc(t)

	KEY := "qconfapi.v2.TestGetFromMcache"
	VAL := []byte("val")
	mcCli := memcache.New("127.0.0.1:11211")
	mcCli.Delete(KEY)

	log := xlog.NewWith("TestGetFromMcache")
	cfg := &Config{
		MasterHosts: []string{""},
		McHosts:     []string{"127.0.0.1:11211", "127.0.0.1:11211"},
	}
	cli := New(cfg)

	assert.Equal(t, 0, atomic.LoadUint32(&cli.mcIdxBegin))

	for i := 0; i < 3; i++ {
		_, _, err := cli.getFromMcache(log, KEY)
		assert.Equal(t, memcache.ErrCacheMiss, err)
		assert.Equal(t, uint32(1-i%2), atomic.LoadUint32(&cli.mcIdxBegin)%2)
	}

	err := mcCli.Set(&memcache.Item{
		Key:   KEY,
		Value: VAL,
		Flags: uint32(200),
	})
	assert.NoError(t, err)

	for i := 0; i < 3; i++ {
		code, val, err := cli.getFromMcache(log, KEY)
		assert.NoError(t, err)
		assert.Equal(t, VAL, val)
		assert.Equal(t, 200, code)
		assert.Equal(t, 1, atomic.LoadUint32(&cli.mcIdxBegin)%2)
	}
}

func TestMcExpires(t *testing.T) {
	skipIfNoMc(t)

	VAL, _ := bson.Marshal(map[string]string{"a": "haha"})
	masterSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write(VAL)
	}))

	log := xlog.NewWith("TestMcExpires")
	cfg := &Config{
		McHosts:     []string{"127.0.0.1:11211", "127.0.0.1:11211"},
		MasterHosts: []string{masterSvr.URL},
		McExpires:   1,
	}
	cli := New(cfg)
	id := "qconfapi.v2.TestMcExpires"
	mcCli := memcache.New("127.0.0.1:11211")
	mcCli.Delete(id)

	var ret map[string]string
	err := cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	item, err := mcCli.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, id, item.Key)
	assert.Equal(t, VAL, item.Value)
	assert.Equal(t, 200, item.Flags)
	time.Sleep(1.2e9)
	_, err = mcCli.Get(id)
	assert.Equal(t, err, memcache.ErrCacheMiss)
}

func TestLcacheDuration(t *testing.T) {
	count := uint32(0)
	VAL, _ := bson.Marshal(map[string]string{"a": "haha"})
	masterSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddUint32(&count, 1)
		w.Write(VAL)
	}))

	log := xlog.NewWith("TestLcacheDuration")
	cfg := &Config{
		MasterHosts:    []string{masterSvr.URL},
		LcacheExpires:  10000,
		LcacheDuration: 500,
	}
	cli := New(cfg)
	id := "qconfapi.v2.TestLcacheDuration"
	var ret map[string]string
	// 第一次，没有lc
	err := cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	assert.Equal(t, 1, count)

	// 命中lc
	err = cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	assert.Equal(t, 1, count)

	time.Sleep(1e9)

	// 命中lc，触发异步刷
	err = cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	assert.Equal(t, 1, count)

	time.Sleep(0.5e9)
	// 异步刷完成
	assert.Equal(t, 2, count)

	// 命中lc, lc中是 []byte
	assert.True(t, len(cli.Lcache.cache[id].val) > 0)
	assert.True(t, cli.Lcache.cache[id].ret == nil)
	err = cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	assert.Equal(t, 2, count)

	// 命中lc, lc中是 interface{}
	assert.True(t, len(cli.Lcache.cache[id].val) == 0)
	assert.True(t, cli.Lcache.cache[id].ret != nil)
	err = cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	assert.Equal(t, 2, count)
}

func TestLcacheExpires(t *testing.T) {
	count := uint32(0)
	VAL, _ := bson.Marshal(map[string]string{"a": "haha"})
	masterSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddUint32(&count, 1)
		w.Write(VAL)
	}))

	log := xlog.NewWith("TestLcacheExpires")
	cfg := &Config{
		MasterHosts:    []string{masterSvr.URL},
		LcacheExpires:  500,
		LcacheDuration: 500,
	}
	cli := New(cfg)
	id := "qconfapi.v2.TestLcacheExpires"
	var ret map[string]string
	err := cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	assert.Equal(t, 1, count)

	err = cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	assert.Equal(t, 1, count)

	time.Sleep(1e9)

	err = cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	assert.Equal(t, 2, count)
}

func TestGetFromMaster(t *testing.T) {
	VAL, _ := bson.Marshal(map[string]string{"a": "haha"})
	masterSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write(VAL)
	}))

	log := xlog.NewWith("TestGetFromMaster")
	cfg := &Config{
		McHosts:           []string{"127.0.0.1:11211", "127.0.0.1:11211"},
		MasterHosts:       []string{masterSvr.URL},
		LcacheExpires:     300,
		LcacheDuration:    300,
		LcacheChanBufSize: 300,
	}
	cli := New(cfg)
	id := "qconfapi.v2.TestGetFromMaster"
	mcCli := memcache.New("127.0.0.1:11211")
	mcCli.Delete(id)

	mcCli.Set(&memcache.Item{Key: id, Value: []byte("wrongdata"), Flags: 200})

	var ret map[string]string
	err := cli.GetFromMaster(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	if isMcAlive() {
		item, err := mcCli.Get(id)
		assert.NoError(t, err)
		assert.Equal(t, id, item.Key)
		assert.Equal(t, VAL, item.Value)
		assert.Equal(t, 200, item.Flags)
	}

	hit, miss := cli.Lcache.Stat()
	assert.Equal(t, hit, 0)
	assert.Equal(t, miss, 1)
}

func TestUpdateFnLimit(t *testing.T) {
	count := uint32(0)
	VAL, _ := bson.Marshal(map[string]string{"a": "haha"})
	masterSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if atomic.AddUint32(&count, 1) > 1 {
			t.Error("count > 1")
		}
		w.Write(VAL)
	}))

	log := xlog.NewWith("TestUpdateFnLimit")
	cfg := &Config{
		MasterHosts:       []string{masterSvr.URL},
		LcacheExpires:     300000,
		LcacheDuration:    300000,
		LcacheChanBufSize: 300,
	}
	cli := New(cfg)
	id := "qconfapi.v2.TestUpdateFnLimit"

	N := 10000
	wg := sync.WaitGroup{}
	wg.Add(N)
	for i := 0; i < N; i++ {
		log := xlog.NewWith(log.ReqId() + strconv.Itoa(i))
		go func() {
			defer wg.Done()
			var ret map[string]string
			err := cli.Get(log, &ret, id, Cache_Normal)
			assert.NoError(t, err)
			assert.Equal(t, ret, map[string]string{"a": "haha"})
		}()
	}
	wg.Wait()
	hit, miss := cli.Lcache.Stat()
	assert.Equal(t, N-1, hit)
	assert.Equal(t, 1, miss)
}

func TestLcacheBsonUnmarshal(t *testing.T) {
	unitTestFunc = func() {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(2)))
	}
	defer func() {
		unitTestFunc = nil
	}()
	count := uint32(0)
	VAL, _ := bson.Marshal(map[string]string{"a": "haha"})
	masterSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddUint32(&count, 1)
		w.Write(VAL)
	}))

	cfg := &Config{
		MasterHosts:   []string{masterSvr.URL},
		LcacheExpires: 300000,
	}
	cli := New(cfg)
	id := "qconfapi.v2.TestLcacheBsonUnmarshal"
	var wg = sync.WaitGroup{}
	log := xlog.NewDummy()
	var ret map[string]string
	err := cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			log := xlog.NewDummy()
			var ret map[string]string
			err := cli.Get(log, &ret, id, Cache_Normal)
			assert.NoError(t, err)
			assert.Equal(t, ret, map[string]string{"a": "haha"})
		}()
	}
	wg.Wait()
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			log := xlog.NewDummy()
			var ret map[string]string
			err := cli.Get(log, &ret, id, Cache_Normal)
			assert.NoError(t, err)
			assert.Equal(t, ret, map[string]string{"a": "haha"})
		}()
	}
	wg.Wait()
}
