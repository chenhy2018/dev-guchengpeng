package qconfapi

import (
	"testing"
	"time"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

var (
	errNoSuchDomain = httputil.NewError(404, "no such domain")
	errNoSuchBucket = httputil.NewError(612, "no such bucket")
)

type mockRedis struct {
	caches map[string][]byte
	get    int
	set    int
	del    int
}

func (mc *mockRedis) Set(xl *xlog.Logger, key string, val interface{}) error {
	mc.set++
	xl.Info("set key:", key, "value", val)
	bytes, err := bson.Marshal(val)
	if err != nil {
		return err
	}
	mc.caches[key] = bytes
	return nil
}

func (mc *mockRedis) Get(xl *xlog.Logger, key string, ret interface{}) error {
	mc.get++
	xl.Info("get key", key)
	val, ok := mc.caches[key]
	xl.Info(key, ok, val)
	if !ok {
		return ErrCacheMiss
	}

	err := bson.Unmarshal(val, ret)
	if err != nil {
		xl.Warn("redis.Get: bad value =>", err)
		return err
	}

	return nil
}

func (mc *mockRedis) Del(xl *xlog.Logger, key string) error {
	mc.del++
	xl.Info("delete key", key)
	delete(mc.caches, key)
	return nil
}

type TestDoc struct {
	Id string `bson:"id" json:"id"`
}

func getB(xl *xlog.Logger, id string) (doc interface{}, err error) {
	xl.Info("getB id: ", id)
	doc = TestDoc{
		Id: id,
	}
	return doc, nil
}

func startRedisClient(expires int64) (*RdsClient, *mockRedis) {
	redis := &mockRedis{caches: make(map[string][]byte)}
	updateChan := make(chan UpdateItem, DefaultBufSize)

	rdCli := &RdsClient{
		Client:     redis,
		updateChan: updateChan,
		expires:    expires,
		updateFn:   getB,
	}
	go rdCli.routine()
	return rdCli, redis
}

func TestRedis(t *testing.T) {
	expires := int64(10)
	rdCli, redisS := startRedisClient(expires)

	xl := xlog.NewDummy()

	// test Get and Set
	var ret TestDoc
	key1 := "domain:test.com"
	errRes, err := rdCli.GetFromRedis(xl, &ret, key1, 0, time.Now().UnixNano())
	assert.Equal(t, errRes, ErrCacheMiss)
	assert.NoError(t, err)
	doc, err := getB(xl, key1)
	assert.NoError(t, err)
	assert.Equal(t, doc.(TestDoc).Id, key1)
	rdCli.UpdateRedis(xl, key1, doc, 0, nil)
	_, err = rdCli.get(xl, key1)
	assert.NoError(t, err)

	timeS := time.Now().UnixNano()
	errRes, err = rdCli.GetFromRedis(xl, &ret, key1, 0, timeS)
	assert.Equal(t, ret.Id, key1)

	// test expires and update
	_, err = rdCli.get(xl, key1)
	assert.NoError(t, err)
	_, _, err = rdCli.getWithCacheFlags(xl, key1, 0, time.Now().UnixNano())
	assert.NoError(t, err)
	_, err = rdCli.get(xl, key1)
	assert.NoError(t, err)

	_, err = rdCli.get(xl, key1)
	assert.NoError(t, err)
	_, _, err = rdCli.getWithCacheFlags(xl, key1, 0, time.Now().UnixNano())
	assert.NoError(t, err)
	_, err = rdCli.get(xl, key1)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	//test update
	//should delete
	xl2 := xlog.NewDummy()
	rdCli.UpdateRedis(xl2, key1, nil, Cache_Normal, errNoSuchDomain)
	err = redisS.Get(xl2, key1, &ret)
	assert.Equal(t, err, ErrCacheMiss)
	errRes, err = rdCli.GetFromRedis(xl2, &ret, key1, Cache_Normal, time.Now().UnixNano())
	assert.Equal(t, errRes, ErrCacheMiss)
	assert.NoError(t, err)

	//should set
	var ret2 TestDoc
	rdCli.UpdateRedis(xl2, key1, nil, Cache_NoSuchEntry, errNoSuchBucket)
	errRes, err = rdCli.GetFromRedis(xl2, &ret2, key1, Cache_NoSuchEntry, time.Now().UnixNano())
	assert.NoError(t, errRes)
	assert.Equal(t, err, errNoSuchBucket)
	assert.Equal(t, "", ret2.Id)

}
