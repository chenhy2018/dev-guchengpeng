package memcache

import (
	"qbox.us/cc/time"
	"github.com/qiniu/ts"
	"testing"
	gtime "time"
)

func DoGet(p *Service, k string, v2 int32, t *testing.T) {
	v, _ := p.Get(k)
	if v == nil {
		if v2 != -1 {
			ts.Fatal(t, "Cache.Get fail", k, v, v2)
		}
		return
	}
	if v.(int32) != v2 {
		ts.Fatal(t, "Cache.Get fail", k, v, v2)
	}
}

func DoGet2(p *Service, k string, v2 int32, ok2 bool, expireTime int64, t *testing.T) {
	v, ok := p.Get2(k, expireTime)
	if v == nil {
		if v2 != -1 {
			ts.Fatal(t, "Cache.Get fail", k, v, v2)
		}
		return
	}
	if ok2 != ok {
		ts.Fatal(t, "Cache.Get fail", k, ok2, ok)
	}
	if v.(int32) != v2 {
		ts.Fatal(t, "Cache.Get fail", k, v, v2)
	}
}

func DoCheck(p *Service, k string, v2 int32, t *testing.T) {
	v := p.Check(k)
	if v == nil {
		if v2 != -1 {
			ts.Fatal(t, "Cache.Get fail", k, v, v2)
		}
		return
	}
	if v.(int32) != v2 {
		ts.Fatalf(t, "Cache.Check fail: %v, %v", k, v)
	}
}

func DoSet(p *Service, k string, v2 int32) {
	p.Set(k, v2, 10e9)
}

func DoDel(p *Service, k string) {
	p.Del(k)
}

func DoTestCache(p *Service, t *testing.T) {
	DoGet(p, "1", -1, t)

	DoSet(p, "1", 1)
	DoCheck(p, "1", 1, t)

	DoSet(p, "2", 2)
	DoSet(p, "3", 3)
	DoCheck(p, "1", 1, t)

	DoSet(p, "4", 4)
	DoCheck(p, "1", -1, t)

	DoGet(p, "2", 2, t)
	DoSet(p, "5", 5)
	DoGet(p, "3", -1, t)
	DoCheck(p, "2", 2, t)

	DoSet(p, "1", 11)
	DoGet(p, "1", 11, t)

	DoSet(p, "a", 1)
	DoSet(p, "b", 2)
	DoDel(p, "a")
	DoGet(p, "a", -1, t)
	DoGet(p, "b", 2, t)

	DoSet(p, "x", 1)
	gtime.Sleep(1e9)
	DoGet2(p, "x", 1, false, time.Nanoseconds(), t)
	DoGet2(p, "x", 1, true, 0, t)
}

func TestIntCache(t *testing.T) {
	p := New(3)
	DoTestCache(p, t)
}
