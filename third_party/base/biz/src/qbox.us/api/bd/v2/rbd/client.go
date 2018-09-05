package rbd

import (
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	bd "qbox.us/api/bd/v2"
)

var DefaultTryTimes = 2

type Config struct {
	Hosts       []string `json:"hosts"`
	TryTimes    int      `json:"try_times"`
	DialTimeout int      `json:"dial_timeout_ms"`
}

type Client struct {
	conns    []*bd.Conn
	putIndex uint32
	getIndex uint32
	tryTimes int
}

func New(cfg *Config) (*Client, error) {
	if len(cfg.Hosts) == 0 {
		return nil, errors.New("rbdapi: not specify hosts")
	}

	if cfg.TryTimes == 0 {
		cfg.TryTimes = DefaultTryTimes
	}

	var t http.RoundTripper
	if cfg.DialTimeout == 0 {
		t = rpc.DefaultTransport
	} else {
		t = rpc.NewTransportTimeout(time.Duration(cfg.DialTimeout)*time.Millisecond, 0)
	}

	conns := make([]*bd.Conn, len(cfg.Hosts))
	for i, host := range cfg.Hosts {
		conns[i] = bd.NewConn(host, t)
	}

	return &Client{conns: conns, tryTimes: cfg.TryTimes}, nil
}

func (p *Client) Get(l rpc.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) (err error) {

	xl := xlog.NewWith(l)
	getIndex := atomic.AddUint32(&p.getIndex, 1)
	var n int64
	var retry bool
	for i := 0; i < p.tryTimes; i++ {
		index := (getIndex + uint32(i)) % uint32(len(p.conns))
		n, retry, err = p.conns[index].Get(l, key, w, from, to, bds)
		xl.Info("rbd.Client.Get:", index, n, retry, err)
		if !retry { // TODO: w is dirty, from += n && retry
			return
		}
		// 1. client->server network error
		// 2. 500 or 57X <- server
		// now n must be 0, retry.
	}
	return
}

func (p *Client) Put(l rpc.Logger, key []byte, r io.Reader, n int, bds [3]uint16) (err error) {

	xl := xlog.NewWith(l)
	putIndex := atomic.AddUint32(&p.putIndex, 1)
	var retry bool
	for i := 0; i < p.tryTimes; i++ {
		index := (putIndex + uint32(i)) % uint32(len(p.conns))
		_, retry, err = p.conns[index].Put(xl, r, n, key, false, bds)
		xl.Info("rbd.Client.Put:", index, n, retry, err)
		if !retry {
			return
		}
		// 1. client->server network error
		// 2. 500 or 57X <- server
		// now r is dirty, must retry after seek.
		seeker, ok := r.(io.ReadSeeker)
		if !ok { // bdtask 使用 rbdapi，可以 seek
			xl.Warn("r is not Seeker, retry failed")
			break
		}
		if _, err2 := seeker.Seek(0, 0); err2 != nil {
			xl.Warn("seeker.Seek(0, 0) failed:", err2)
			break
		}
		r = seeker
	}
	return
}
