package memcache

import (
	"encoding/json"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qiniu/xlog.v1"
)

type MemcacheEx interface {
	Memcache
	SetWithExpires(xl *xlog.Logger, key string, val interface{}, expires int32) error
}

type mcServiceEx struct {
	*mcService
}

func NewEx(conns []Conn) MemcacheEx {

	sel, err := newMcSelector(conns)
	if err != nil {
		panic("newMcService: newMcSelector failed => " + err.Error())
	}
	mcService := &mcService{
		Client: memcache.NewFromSelector(sel),
	}
	return &mcServiceEx{
		mcService: mcService,
	}
}

func NewExWithTimeout(conns []Conn, timeout time.Duration) MemcacheEx {
	m := NewEx(conns).(*mcServiceEx)
	m.Timeout = timeout
	return m
}

// ----------------------------------------------------------------------------
func (p *mcServiceEx) SetWithExpires(xl *xlog.Logger, key1 string, value interface{}, expires int32) error {
	var xerr error
	defer xl.Xtrack("mc.s", time.Now(), &xerr)

	key := genMcKey(key1)
	b, err := json.Marshal(value)
	if err != nil {
		xerr = errMcBad
		xl.Warn("memcache.Set: bad value =>", err)
		return err
	}
	item := &memcache.Item{
		Key:        key,
		Value:      b,
		Expiration: expires,
	}
	err = p.Client.Set(item)
	if err != nil {
		xerr = errMc500
		xl.Warn("memcache.Set: call failed =>", err)
		return err
	}
	return nil
}
