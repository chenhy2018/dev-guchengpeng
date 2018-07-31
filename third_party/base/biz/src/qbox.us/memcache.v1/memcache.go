package memcache

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/qiniu/xlog.v1"
)

var ErrCacheMiss = memcache.ErrCacheMiss

type Memcache interface {
	Set(xl *xlog.Logger, key string, val interface{}) (err error)
	Get(xl *xlog.Logger, key string, ret interface{}) (err error)
	Del(xl *xlog.Logger, key string) (err error)
}

type mcService struct {
	*memcache.Client
}

func New(conns []Conn) Memcache {

	sel, err := newMcSelector(conns)
	if err != nil {
		panic("newMcService: newMcSelector failed => " + err.Error())
	}
	return &mcService{
		Client: memcache.NewFromSelector(sel),
	}
}

func NewWithTimeout(conns []Conn, timeout time.Duration) Memcache {
	m := New(conns).(*mcService)
	m.Timeout = timeout
	return m
}

// -----------------------------------------------------------------------------

var errMc404 = errors.New("404")
var errMc500 = errors.New("500")
var errMcBad = errors.New("bad")

// Same as qboxmc.
func genMcKey(key1 string) string {

	key := base64.URLEncoding.EncodeToString([]byte(key1))
	if len(key) <= 250 {
		return key
	}
	h := sha1.New()
	io.WriteString(h, key)
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func (p *mcService) Set(xl *xlog.Logger, key1 string, value interface{}) error {

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
		Key:   key,
		Value: b,
	}
	err = p.Client.Set(item)
	if err != nil {
		xerr = errMc500
		xl.Warn("memcache.Set: call failed =>", err)
		return err
	}
	return nil
}

func (p *mcService) Get(xl *xlog.Logger, key1 string, ret interface{}) error {

	var xerr error
	defer xl.Xtrack("mc.g", time.Now(), &xerr)

	key := genMcKey(key1)
	item, err := p.Client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			xerr = errMc404
			return err
		}
		xerr = errMc500
		xl.Warn("memcache.Get: call failed =>", err)
		return err
	}

	err = json.NewDecoder(bytes.NewReader(item.Value)).Decode(ret)
	if err != nil {
		xerr = errMcBad
		xl.Warn("memcache.Get: bad value =>", err)
		return err
	}
	return nil
}

func (p *mcService) Del(xl *xlog.Logger, key1 string) error {

	var xerr error
	defer xl.Xtrack("mc.d", time.Now(), &xerr)

	key := genMcKey(key1)
	err := p.Client.Delete(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			xerr = errMc404
			return err
		}
		xerr = errMc500
		xl.Warn("memcache.Del: call failed =>", err)
		return err
	}
	return nil
}
