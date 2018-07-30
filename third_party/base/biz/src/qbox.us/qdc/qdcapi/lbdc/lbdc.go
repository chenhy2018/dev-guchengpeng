package lbdc

import (
	"io"
	"qbox.us/cc/time"
	"qbox.us/dht"
	"github.com/qiniu/xlog.v1"
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
