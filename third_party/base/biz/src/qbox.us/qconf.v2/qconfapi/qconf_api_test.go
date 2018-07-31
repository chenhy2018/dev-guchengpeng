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
	"gopkg.in/redis.v5"
	"launchpad.net/mgo/bson"
)

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

func skipIfNoRedis(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Println("skipped Redis test")
		t.Skip("no redis available")

	}
	conn.Close()

}

func TestRedisClient(t *testing.T) {
	skipIfNoRedis(t)

	KEY := "qconfapi.v2.TestGetFromMcache"
	VAL := []byte("val")

	redisS := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	assert.NoError(t, redisS.Del(KEY).Err())

	log := xlog.NewWith("TestGetFromRedis")
	redisCfg := RedisCfg{
		Hosts: []string{"127.0.0.1:6379", "127.0.0.1:6379"},
	}
	cfg := &Config{
		MasterHosts:       []string{""},
		RedisCfg:          redisCfg,
		RedisCacheExpires: 500,
	}
	cli := New(cfg)

	assert.Equal(t, 0, atomic.LoadUint32(&cli.redisIdxBegin))

	for i := 0; i < 3; i++ {
		_, _, _, err := cli.getFromRedis(log, KEY, Cache_Normal, time.Now().UnixNano())
		assert.Equal(t, redis.Nil, err)
		assert.Equal(t, uint32(1-i%2), atomic.LoadUint32(&cli.redisIdxBegin)%2)
	}

	timeS := time.Now().UnixNano()
	item := CacheItem{
		Val:  VAL,
		T:    timeS,
		Code: 200,
	}

	bytes, err := bson.Marshal(item)
	assert.NoError(t, err)
	err = redisS.Set(KEY, bytes, 0).Err()
	assert.NoError(t, err)

	for i := 0; i < 3; i++ {
		code, val, exp, err := cli.getFromRedis(log, KEY, Cache_Normal, timeS)
		assert.NoError(t, err)
		assert.Equal(t, VAL, val)
		assert.Equal(t, 200, code)
		assert.Equal(t, 1, atomic.LoadUint32(&cli.redisIdxBegin)%2)
		assert.Equal(t, exp, false)
	}
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

func TestRedisExpires(t *testing.T) {
	skipIfNoRedis(t)
	count := uint32(0)
	VAL, _ := bson.Marshal(map[string]string{"a": "haha"})
	masterSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddUint32(&count, 1)
		w.Write(VAL)
	}))

	log := xlog.NewWith("TestRedisExpires")
	redisCfg := RedisCfg{
		Hosts: []string{"127.0.0.1:6379", "127.0.0.1:6379"},
	}
	cfg := &Config{
		MasterHosts:       []string{masterSvr.URL},
		RedisCfg:          redisCfg,
		RedisCacheExpires: 500,
	}
	cli := New(cfg)
	go cli.routine()

	id := "qconfapi.v2.TestRedisExpires"

	redisS := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	assert.NoError(t, redisS.Del(id).Err())

	var ret map[string]string
	err := cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)
	assert.Equal(t, ret, map[string]string{"a": "haha"})
	assert.Equal(t, 1, count)

	timeS := time.Now().UnixNano()
	item := CacheItem{
		Val:  VAL,
		T:    timeS,
		Code: 200,
	}

	bytes, err := bson.Marshal(item)
	assert.NoError(t, err)
	err = redisS.Set(id, bytes, 0).Err()
	assert.NoError(t, err)

	time.Sleep(1e9)
	err = cli.Get(log, &ret, id, Cache_Normal)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	assert.Equal(t, 2, count)

}

func TestGetFromMaster(t *testing.T) {
	VAL, _ := bson.Marshal(map[string]string{"a": "haha"})
	masterSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write(VAL)
	}))

	log := xlog.NewWith("TestGetFromMaster")
	cfg := &Config{
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
