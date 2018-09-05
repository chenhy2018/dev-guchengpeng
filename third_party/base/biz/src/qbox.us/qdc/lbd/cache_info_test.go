package lbd

import (
	"fmt"
	"os"
	"github.com/qiniu/log.v1"
	"qbox.us/store/cc"
	"testing"
)

func init() {
	log.SetOutputLevel(0)
}

func TestCacheInfo(t *testing.T) {

	bug48(t)
	bug53(t)

	cache := cc.NewSimpleIntCache(258)
	pool := cc.NewChunkPool(258)

	cacheInfo, err := NewCacheInfo("cacheinfo", 258)
	defer os.Remove("cacheinfo")
	if err != nil {
		t.Fatal(err)
	}
	key := make([]byte, 20)

	for i := 0; i < 255; i++ {
		set(t, int32(i), key, cache, cacheInfo)
	}

	cacheInfo, err = NewCacheInfo("cacheinfo", 258)
	if err != nil {
		t.Fatal(err)
	}
	err = cacheInfo.Load(cache, pool)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 255; i++ {
		get(t, int32(i), key, cache, cacheInfo)
	}

	for i := 0; i < 10; i++ {
		n := pool.Alloc()
		m := pool.Alloc()
		fmt.Println("alloc :", n, m)
		pool.Free(n)
		pool.Free(m)
	}
}

func set(t *testing.T, n int32, key []byte, cache *cc.SimpleIntCache, info *CacheInfo) {
	key[0] = byte(n)
	info.Set(key, n)
}

func get(t *testing.T, n int32, key []byte, cache *cc.SimpleIntCache, info *CacheInfo) {
	key[0] = byte(n)
	m := cache.Get(string(key))
	if n != m {
		t.Fatal("check failed:", n, m)
	}
}

func bug53(t *testing.T) {
	d := make([]byte, 25)
	d[20] = 0xBE
	cacheInfo, err := NewCacheInfo("cacheinfo.bug53", 2)
	defer os.Remove("cacheinfo.bug53")
	if err != nil {
		t.Fatal(err)
	}
	err = cacheInfo.Set(d[:20], 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(d)
	if d[20] != 0xBE {
		t.Fatal("bug53")
	}
}

func bug48(t *testing.T) {
	key := make([]byte, 20)
	cache := cc.NewSimpleIntCache(5)
	pool := cc.NewChunkPool(5)
	cacheInfo, err := NewCacheInfo("cacheinfo.bug48", 5)
	defer os.Remove("cacheinfo.bug48")
	if err != nil {
		t.Fatal(err)
	}
	cacheInfo.Set(key, 0)
	cacheInfo.Set(key, 1)
	cacheInfo.Set(key, 2)
	cacheInfo.Clear(0)
	cacheInfo.Clear(1)

	cacheInfo2, err := NewCacheInfo("cacheinfo.bug48", 5)
	if err != nil {
		t.Fatal(err)
	}
	cacheInfo2.Load(cache, pool)
	n1 := pool.Alloc()
	n2 := pool.Alloc()
	n3 := pool.Alloc()
	n4 := pool.Alloc()
	fmt.Println(n1, n2, n3, n4)
}
