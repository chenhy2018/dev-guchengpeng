package cc

import (
	"testing"
)

func DoGet(p IntCache, k string, v2 int32, t *testing.T) {
	v := p.Get(k)
	if v != v2 {
		t.Fatal("Cache.Get fail")
	}
}

func DoCheck(p IntCache, k string, v2 int32, t *testing.T) {
	v := p.Check(k)
	if v != v2 {
		t.Fatalf("Cache.Check fail: %v, %v", k, v)
	}
}

func DoSet(p IntCache, k string, v2 int32) {
	p.Set(k, v2)
}

func DoTestCache(p IntCache, t *testing.T) {
	DoGet(p, "1", -1, t)

	DoSet(p, "1", 1)
	DoCheck(p, "1", 1, t)

	DoSet(p, "2", 2)
	DoSet(p, "3", 3)
	DoCheck(p, "1", 1, t)

	DoSet(p, "4", 4)
	DoCheck(p, "1", -1, t)
	//	return

	DoGet(p, "2", 2, t)
	DoSet(p, "5", 5)
	DoGet(p, "3", -1, t)
	DoCheck(p, "2", 2, t)
}

func TestIntCache(t *testing.T) {
	p := NewSimpleIntCache(3)
	DoTestCache(p, t)
}
