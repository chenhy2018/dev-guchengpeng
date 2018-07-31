package cc

import (
	"testing"
)

func TestKeyCache(t *testing.T) {

	flag := true
	outOfLimit := func(space int64, count int) bool {
		return flag
	}

	var cache KeyCache
	cache = NewSimpleKeyCache(outOfLimit)

	var old, key string

	flag = false
	key = cache.Get("a") // -1
	if key != "" {
		t.Fatal("first get failed", key)
	}
	old = cache.Set("a", 0)
	if old != "" {
		t.Fatal("first set failed", old)
	}
	key = cache.Get("a") // 1
	if key != "a" {
		t.Fatal("get cache failed", key)
	}

	old = cache.Set("b", 0)
	if old != "" {
		t.Fatal("set failed", old)
	}
	key = cache.Get("a") // 1
	if key != "a" {
		t.Fatal("get cache failed", key)
	}
	key = cache.Get("b") // 1
	if key != "b" {
		t.Fatal("get cache failed", key)
	}

	flag = true
	old = cache.Set("c", 0)
	if old != "a" {
		t.Fatal("eliminate failed", old)
	}
	key = cache.Get("a") // -1
	if key != "" {
		t.Fatal("get failed after eliminate", key)
	}
	key = cache.Get("b") // 1
	if key != "b" {
		t.Fatal("get cache failed", key)
	}
	key = cache.Get("c") // 1
	if key != "c" {
		t.Fatal("get cache failed", key)
	}

	flag = false
	old = cache.Set("d", 0)
	if old != "" {
		t.Fatal("set failed", old)
	}

}
