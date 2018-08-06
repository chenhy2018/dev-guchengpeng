package lbdc

import (
	"io"
	"qbox.us/cc/time"
	"qbox.us/dht"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"

	ioadmin "qbox.us/admin_api/io"
)

type Reseter interface {
	Reset()
}

type LBdClient struct {
	clients       map[string]*Conn
	dht           dht.Interface
	retryInterval int64
	retryTimes    int // ttl
}

func New(dht dht.Interface, clients map[string]*Conn, retryInterval int64, retryTimes int) *LBdClient {
	return &LBdClient{clients, dht, retryInterval, retryTimes}
}

func (p *LBdClient) ServiceStat() (info []*ioadmin.CacheInfoEx, err error) {
	n := len(p.clients)
	info = make([]*ioadmin.CacheInfoEx, 0, n)
	for host, c := range p.clients {
		ret, _, err1 := c.ServiceStat()
		if err1 != nil {
			log.Warn("LBdClient.ServiceStat failed:", host, err1)
			continue
		}
		cache := &ioadmin.CacheInfoEx{
			Host:    host,
			Missing: ret.Missing,
			Total:   ret.Total,
			Wtotal:  ret.Wtotal,
		}
		info = append(info, cache)
	}
	return
}

func (p *LBdClient) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) (err error) {

	xl.Debug("LBdClient.Get:", key, bds)
	now := time.Seconds()
	routers := p.dht.Route(key, p.retryTimes)
	for _, router := range routers {
		client, ok := p.clients[router.Host]
		if !ok {
			continue
		}
		lastFailedTime := client.GetLastFailedTime()
		if lastFailedTime != 0 && now-lastFailedTime < p.retryInterval {
			continue
		}
		var n int64
		n, err = client.Get(xl, key, w, from, to, bds)
		if err == nil {
			return
		}

		if reseter, ok := w.(Reseter); ok { // retry
			reseter.Reset()
			continue
		}
		if n == 0 { // retry
			continue
		}

		xl.Warn("LbdClient.Get: can NOT retry by dirty:", err)
		return err
	}
	return err
}

func (p *LBdClient) GetLocal(xl *xlog.Logger, key []byte) (rc io.ReadCloser, n int, err error) {

	xl.Debug("lbdc.GetLocal:", key)
	now := time.Seconds()
	routers := p.dht.Route(key, p.retryTimes)

	err = EServerNotAvailable
	for _, router := range routers {
		client, ok := p.clients[router.Host]
		if !ok {
			continue
		}
		lastFailedTime := client.GetLastFailedTime()
		if lastFailedTime != 0 && now-lastFailedTime < p.retryInterval {
			continue
		}
		rc, n, err = client.GetLocal(xl, key)
		if err == nil {
			return
		}
	}
	return
}

func (p *LBdClient) Put(xl *xlog.Logger, key []byte, r io.Reader, n int, doCache bool, bds [3]uint16) (err error) {

	now := time.Seconds()
	xl.Debug("LBdClient.Put: pick client key:", key)
	routers := p.dht.Route(key, p.retryTimes)
	xl.Debug("LbdClient.Put: routers:", routers)
	for _, router := range routers {
		client, ok := p.clients[router.Host]
		if !ok {
			continue
		}
		lastFailedTime := client.GetLastFailedTime()
		if lastFailedTime != 0 && now-lastFailedTime < p.retryInterval {
			continue
		}
		_, err = client.Put2(xl, r, n, key, doCache, bds)
		if err == nil {
			return
		}

		if seeker, ok := r.(io.Seeker); ok { // retry
			seeker.Seek(0, 0)
			continue
		}

		xl.Warn("LbdClient.Put: can NOT retry by dirty :", err)
		return err
	}
	return err
}

func (p *LBdClient) PutLocal(xl *xlog.Logger, key []byte, r io.Reader, n int) (err error) {

	now := time.Seconds()
	xl.Debug("lbd.PutLocal, pick client key:", key)
	routers := p.dht.Route(key, p.retryTimes)
	xl.Debug("lbd.PutLocal, routers:", routers)
	for _, router := range routers {
		client, ok := p.clients[router.Host]
		if !ok {
			continue
		}
		lastFailedTime := client.GetLastFailedTime()
		if lastFailedTime != 0 && now-lastFailedTime < p.retryInterval {
			continue
		}
		err = client.PutLocal(xl, key, r, n)
		if err == nil {
			return
		}

		if seeker, ok := r.(io.Seeker); ok { // retry
			seeker.Seek(0, 0)
			continue
		}

		xl.Warn("lbd.PutLocal, can NOT retry by dirty :", err)
		return err
	}
	return err
}
