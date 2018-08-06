package localcache

import (
	"testing"
	"time"
)

func TestMemCache(t *testing.T) {
	cache, _ := NewMemCache(time.Millisecond * 500)
	data1 := []byte("123456")
	data2 := []byte("abcdefg")

	cache.Set("data1", data1)
	time.Sleep(time.Millisecond * 300)
	cache.Set("data2", data2)

	ret1, err := cache.Get("data1")
	t.Logf("cache.Get data1, got %v, %v, ", string(ret1), err)
	if err != nil {
		t.Fatal("cache.Get data1, expect: nil, but got:", err)
	}
	if string(ret1) != string(data1) {
		t.Fatalf("cache.Get data1, expect: %v, but got: %v", string(data1), string(ret1))
	}

	// after more 300ms, data1 expired, data2 not expired
	time.Sleep(time.Millisecond * 300)
	ret1, err = cache.Get("data1")
	t.Logf("cache.Get data1, got: %v, %v, ", string(ret1), err)
	if err != ErrNoSuchEntry {
		t.Fatalf("cache.Get data1, expect: %v, but got: %v", ErrNoSuchEntry, err)
	}

	ret2, err := cache.Get("data2")
	t.Logf("cache.Get data2, got: %v, %v, ", string(ret2), err)
	if err != nil {
		t.Fatal("cache.Get data2, expect: nil, but got:", err)
	}
	if string(ret2) != string(data2) {
		t.Fatalf("cache.Get data2, expect: %v, but got: %v", string(data2), string(ret2))
	}

	// after more 300ms, data2 expired
	time.Sleep(time.Millisecond * 300)
	_, err = cache.Get("data2")
	t.Logf("cache.Get data2, got: %v, %v, ", string(ret1), err)
	if err != ErrNoSuchEntry {
		t.Fatalf("cache.Get data2, expect: %v, but got: %v", ErrNoSuchEntry, err)
	}
}
