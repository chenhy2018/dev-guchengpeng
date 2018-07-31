package lbdc

import (
	"io"
	"strings"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"

	"qbox.us/cc/time"
	"qbox.us/dht"

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
	acl           *Acl
}

func New(dht dht.Interface, clients map[string]*Conn, retryInterval int64, retryTimes int, acl *Acl) *LBdClient {
	return &LBdClient{clients, dht, retryInterval, retryTimes, acl}
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

func isWriteFailed(err error) bool {

	// Such as "write tcp 192.168.0.126:18072: broken pipe" or "write tcp 192.168.0.126:52119: connection reset by peer"
	return strings.Contains(err.Error(), "write ")
}

func (p *LBdClient) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) (err error) {

	now := time.Seconds()
	routers := p.dht.Route(key, len(p.dht.Nodes()))
	xl.Debug("LBdClient.Get:", key, bds, routers)

	picker, err := newPickerProxy(routers)
	if err != nil {
		return
	}

	err = EServerNotAvailable
	for i := 0; i < p.retryTimes; i++ {
		host := picker.One()
		client, ok := p.clients[host]
		if !ok {
			continue
		}
		lastFailedTime := client.GetLastFailedTime()
		if lastFailedTime != 0 && now-lastFailedTime < p.retryInterval {
			xl.Warnf("LbdClient.Get: %v last failed, last %v now %v interval %v", host, lastFailedTime, now, p.retryInterval)
			continue
		}
		var releaseFunc func()
		releaseFunc, err = p.acl.AcquireWithBd(host, bds[0])
		if err != nil {
			xlogAclError(xl, err)
			if shouldRetry(err) {
				continue
			}
			break
		}
		var n int64
		n, err = client.Get(xl, key, w, from, to, bds)
		releaseFunc()
		if err == nil {
			return
		}
		if isWriteFailed(err) {
			xl.Warnf("LbdClient.Get: %v get %v bytes and failed %v, not retry", host, n, err)
			return
		}
		from += int(n)
		xl.Warnf("LbdClient.Get: %v get %v bytes and failed %v", host, n, err)
	}
	return err
}

type releaseCloser struct {
	io.ReadCloser
	releaseFn func()
}

func (p releaseCloser) Close() error {
	p.releaseFn()
	return p.ReadCloser.Close()
}

func (p *LBdClient) GetLocal(xl *xlog.Logger, key []byte) (rc io.ReadCloser, n int, err error) {

	xl.Debug("lbdc.GetLocal:", key)
	now := time.Seconds()
	routers := p.dht.Route(key, len(p.dht.Nodes()))

	picker, err := newPickerProxy(routers)
	if err != nil {
		return
	}

	err = EServerNotAvailable
	for i := 0; i < p.retryTimes; i++ {
		host := picker.One()
		client, ok := p.clients[host]
		if !ok {
			continue
		}
		lastFailedTime := client.GetLastFailedTime()
		if lastFailedTime != 0 && now-lastFailedTime < p.retryInterval {
			continue
		}
		var releaseFunc func()
		releaseFunc, err = p.acl.Acquire(host)
		if err != nil {
			xlogAclError(xl, err)
			if shouldRetry(err) {
				continue
			}
			break
		}
		rc, n, err = client.GetLocal(xl, key)
		if err == nil {
			rc = releaseCloser{rc, releaseFunc}
			return
		}
		releaseFunc()
		if err == EKeyNotFound {
			return
		}
	}
	return
}

func (p *LBdClient) Put(xl *xlog.Logger, key []byte, r io.Reader, n int, doCache bool, bds [3]uint16) (err error) {

	now := time.Seconds()
	xl.Debug("LBdClient.Put: pick client key:", key)
	routers := p.dht.Route(key, len(p.dht.Nodes()))
	xl.Debug("LbdClient.Put: routers:", routers)

	picker, err := newPickerProxy(routers)
	if err != nil {
		return
	}

	err = EServerNotAvailable
	for i := 0; i < p.retryTimes; i++ {
		host := picker.One()
		client, ok := p.clients[host]
		if !ok {
			continue
		}
		lastFailedTime := client.GetLastFailedTime()
		if lastFailedTime != 0 && now-lastFailedTime < p.retryInterval {
			continue
		}
		var releaseFunc func()
		releaseFunc, err = p.acl.AcquireWithBd(host, bds[0])
		if err != nil {
			xlogAclError(xl, err)
			if shouldRetry(err) {
				continue
			}
			break
		}
		_, err = client.Put2(xl, r, n, key, doCache, bds)
		releaseFunc()
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
	routers := p.dht.Route(key, len(p.dht.Nodes()))
	xl.Debug("lbd.PutLocal, routers:", routers)

	picker, err := newPickerProxy(routers)
	if err != nil {
		return
	}

	err = EServerNotAvailable
	for i := 0; i < p.retryTimes; i++ {
		host := picker.One()
		client, ok := p.clients[host]
		if !ok {
			continue
		}
		lastFailedTime := client.GetLastFailedTime()
		if lastFailedTime != 0 && now-lastFailedTime < p.retryInterval {
			continue
		}
		var releaseFunc func()
		releaseFunc, err = p.acl.Acquire(host)
		if err != nil {
			xlogAclError(xl, err)
			if shouldRetry(err) {
				continue
			}
			break
		}
		err = client.PutLocal(xl, key, r, n)
		releaseFunc()
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
