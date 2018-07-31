package clients

import (
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"io"
	"net/http"
	"qbox.us/errors"
	"qbox.us/fop"
	"sync/atomic"
)

type Fopg struct {
	conns      []*Conn
	retryTimes int
	lastIndex  uint32
}

func NewFopg(hosts []string, retryTimes int, transport http.RoundTripper) *Fopg {

	return new(Fopg).init(hosts, retryTimes, transport)
}

func (p *Fopg) init(hosts []string, retryTimes int, transport http.RoundTripper) *Fopg {

	p.conns = make([]*Conn, len(hosts))
	p.retryTimes = retryTimes
	p.lastIndex = 0
	for i, h := range hosts {
		p.conns[i] = NewConn(h, transport)
	}
	return p
}

func (p *Fopg) List(xl rpc.Logger) (fops []string, err error) {

	return p.conns[0].List(xl)
}

func (p *Fopg) Op(xl *xlog.Logger, fh []byte, fsize int64, fopCtx *fop.FopCtx, nop bool) (
	r io.ReadCloser, length int64, mime string, needCache bool, err error) {

	conns := p.conns
	err = errors.New("No fopg server is available.")
	if len(conns) == 0 {
		xl.Warn("Fopg.Op: No fopg server is available:", *fopCtx)
		return
	}

	index := atomic.AddUint32(&p.lastIndex, 1) % uint32(len(conns))
	for i := 0; i < p.retryTimes; i++ {
		if index >= uint32(len(conns)) {
			index = 0
		}
		retry := false
		r, length, mime, needCache, err, retry = conns[index].Op(xl, fh, fsize, fopCtx, nop)
		xl.Debug("Fopg.Op:", r, length, err)
		if !retry {
			break
		}
		index++
	}
	if err != nil {
		xl.Warn("Fopg.Op: error:", err)
	}
	return
}

func (p *Fopg) OpWithRT(xl *xlog.Logger, fh []byte, fsize int64, fopCtx *fop.FopCtx, nop bool, conn *http.Client) (
	r io.ReadCloser, length int64, mime string, needCache bool, err error) {

	conns := p.conns
	err = errors.New("No fopg server is available.")
	if len(conns) == 0 {
		xl.Warn("Fopg.OpWithRT: No fopg server is available:", *fopCtx)
		return
	}

	index := atomic.AddUint32(&p.lastIndex, 1) % uint32(len(conns))
	for i := 0; i < p.retryTimes; i++ {
		if index >= uint32(len(conns)) {
			index = 0
		}
		retry := false
		r, length, mime, needCache, err, retry = conns[index].OpWithRT(xl, fh, fsize, fopCtx, nop, conn)
		xl.Debug("Fopg.OpWithRT:", r, length, err)
		if !retry {
			break
		}
		index++
	}
	if err != nil {
		xl.Warn("Fopg.OpWithRT: error:", err)
	}
	return
}

func (p *Fopg) RangeOp(xl *xlog.Logger, fh []byte, fsize int64, fopCtx *fop.FopCtx, xrange string, conn *http.Client) (
	r io.ReadCloser, length int64, mime string, contentRange string, err error) {

	conns := p.conns
	err = errors.New("No fopg server is available.")
	if len(conns) == 0 {
		xl.Warn("Fopg.RangeOp: No fopg server is available")
		return
	}

	index := atomic.AddUint32(&p.lastIndex, 1) % uint32(len(conns))
	for i := 0; i < p.retryTimes; i++ {
		if index >= uint32(len(conns)) {
			index = 0
		}
		retry := false
		r, length, mime, contentRange, err, retry = conns[index].RangeOp(xl, fh, fsize, fopCtx, xrange, conn)
		if !retry {
			break
		}
		index++
	}
	if err != nil {
		xl.Warn("Fopg.RangeOp error:", err)
	}
	return
}

func (p *Fopg) PersistentOp(xl *xlog.Logger, ret interface{}, fh []byte, fsize int64, fopCtx *fop.FopCtx, httpConn *http.Client) (err error) {

	conns := p.conns
	err = errors.New("No fopg server is available.")
	if len(conns) == 0 {
		xl.Warn("Fopg.PersistentOp: No fopg server is available:", *fopCtx)
		return
	}

	index := atomic.AddUint32(&p.lastIndex, 1) % uint32(len(conns))
	for i := 0; i < p.retryTimes; i++ {
		if index >= uint32(len(conns)) {
			index = 0
		}
		if err = conns[index].PersistentOp(xl, ret, fh, fsize, fopCtx, httpConn); err == nil {
			break
		}
		index++
	}
	if err != nil {
		xl.Warn("Fopg.PersistentOp: error:", err)
	}
	return
}
