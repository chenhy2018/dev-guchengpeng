package caches

import (
	"testing"
	"time"

	"qbox.us/mgo2"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/stretchr/testify/assert"
)

var (
	cacheCollection = bson.M{
		"name": "test_cache_mgo_provider",
	}

	bizMongo *mgo2.Database
)

type CacheModel struct {
	Id        bson.ObjectId `bson:"_id"`
	Key       string        `bson:"key"`
	Value     interface{}   `bson:"value"`
	ExpiredAt time.Time     `bson:"expired_at"`
}

func connect(callback func(c *mgo.Collection) error) error {
	c := mgo2.CopyCollection(bizMongo.C(cacheCollection))
	defer mgo2.CloseCollection(c)
	return callback(c)
}

func initDatabase(t *testing.T) {
	var err error
	bizMongo, err = mgo2.NewDatabase("mongodb://localhost/qbox_biz_test", "strong")
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	bizMongo.C(cacheCollection).DropCollection()
}

func initMgoProvider(t *testing.T) CacheProvider {
	initDatabase(t)

	config := NewMgoConfig()
	mgoCache := NewMgoProvider(config, connect)

	return mgoCache
}

func testCacheProvider(t *testing.T, cache CacheProvider) {
	var err error

	cache.Set("key", "val")
	assert.True(t, cache.Has("key"))
	assert.Equal(t, cache.Get("key").String(), "val")

	assert.True(t, cache.Get("undefined").IsNil())
	assert.False(t, cache.Has("undefined"))

	cache.Set("key", "none")
	assert.True(t, cache.Has("key"))
	assert.Equal(t, cache.Get("key").String(), "none")

	cache.Delete("key")
	assert.False(t, cache.Has("key"))
	assert.True(t, cache.Get("key").IsNil())

	err = cache.Set("key1", "val")
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	err = cache.Set("key2", "val")
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	err = cache.Set("key3", "val")
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	assert.True(t, cache.Has("key1"))
	assert.True(t, cache.Has("key2"))
	assert.True(t, cache.Has("key3"))
	cache.Clean()
	assert.False(t, cache.Has("key1"))
	assert.False(t, cache.Has("key2"))
	assert.False(t, cache.Has("key3"))

	cache.Set("key", "val", 60)
	assert.True(t, cache.Has("key"))

	cache.Set("key", "val", -1)
	assert.True(t, cache.Has("key"))

	cache.Set("key", "val", 0)
	assert.True(t, cache.Has("key"))
}

func Test_Cache_Mgo_Provider(t *testing.T) {
	cache := initMgoProvider(t)
	testCacheProvider(t, cache)
}

func Test_Cache_Mgo_KeyExpired(t *testing.T) {
	cache := initMgoProvider(t)

	mc := cache.(*MgoProvider)

	cache.Set("key", "val", 0)

	err := connect(func(c *mgo.Collection) error {
		return c.Update(bson.M{
			mc.config.KeyField: "key",
		}, bson.M{
			"$set": bson.M{
				mc.config.ExpiredAtField: time.Now(),
			},
		})
	})

	assert.NoError(t, err)

	assert.False(t, cache.Has("key"))
}
