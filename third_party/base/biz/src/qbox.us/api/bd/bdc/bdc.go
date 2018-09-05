package bdc

import (
	"io"
	"strings"
	"sync/atomic"

	qio "github.com/qiniu/io"

	"github.com/qiniu/xlog.v1"
	"qbox.us/cc/time"
)

type Reseter interface {
	Reset()
}

func isWriterErr(err error) bool {

	return err == io.ErrShortWrite || strings.Contains(err.Error(), "write tcp")
}

type BdClient struct {
	clients       []*Conn
	clientIndex   uint32
	retryTimes    int32
	retryInterval int64
}

func NewBdClient(clients []*Conn, retryInterval int64, retryTimes int) *BdClient {
	return &BdClient{clients, 0, int32(retryTimes), retryInterval}
}

func (p *BdClient) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) (err error) {

	err = EServerNotAvailable
	fromIdx := atomic.AddUint32(&p.clientIndex, 1)
	for i := 0; i < int(p.retryTimes); i++ {
		client, pickIdx := p.pickClient(xl, fromIdx)
		if client == nil {
			return
		}
		fromIdx = pickIdx + 1
		var n int64
		n, err = client.Get(xl, key, w, from, to, bds)
		if err == nil {
			return
		}
		if isWriterErr(err) {
			return
		}
		from += int(n)
	}
	return
}

func (p *BdClient) Put(xl *xlog.Logger, key []byte, r io.Reader, n int, bds [3]uint16) (err error) {

	err = EServerNotAvailable
	fromIdx := atomic.AddUint32(&p.clientIndex, 1)
	for i := 0; i < int(p.retryTimes); i++ {
		client, pickIdx := p.pickClient(xl, fromIdx)
		if client == nil {
			return
		}
		fromIdx = pickIdx + 1
		_, err = client.Put(xl, r, n, key, bds)
		if err == nil {
			return
		}
		xl.Warn("BdClient.Put: Put failed", err)
		if rt, ok := r.(io.Seeker); ok {
			_, err1 := rt.Seek(0, 0)
			if err1 == nil {
				continue // 支持reset的流，重试
			}
		}
		return
	}
	return
}

func (p *BdClient) PutEx(xl *xlog.Logger, key []byte, r io.ReaderAt, n int, bds [3]uint16) (err error) {

	err = EServerNotAvailable
	fromIdx := atomic.AddUint32(&p.clientIndex, 1)
	for i := 0; i < int(p.retryTimes); i++ {
		client, pickIdx := p.pickClient(xl, fromIdx)
		if client == nil {
			return
		}
		fromIdx = pickIdx + 1
		_, err = client.Put(xl, &qio.Reader{ReaderAt: r}, n, key, bds)
		if err == nil {
			return
		}
		xl.Info("BdClient.PutEx: Put failed", err)
	}
	return
}

func (p *BdClient) pickClient(xl *xlog.Logger, fromIdx uint32) (*Conn, uint32) {
	if p.clients == nil || len(p.clients) <= 0 {
		return nil, 0
	}
	num := uint32(len(p.clients))
	now := time.Seconds()
	for i := uint32(0); i < num; i++ {
		pickIdx := fromIdx + i
		conn := p.clients[pickIdx%num]
		fail := conn.GetLastFailedTime()
		if fail == 0 || now-fail >= p.retryInterval {
			return conn, pickIdx
		}
	}
	return nil, 0
}
