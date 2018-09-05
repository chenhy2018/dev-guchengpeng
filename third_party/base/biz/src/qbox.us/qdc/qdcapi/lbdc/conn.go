package lbdc

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"

	"qbox.us/cc/time"
	"github.com/qiniu/xlog.v1"

	qrpc "github.com/qiniu/rpc.v1"
)

type Conn struct {
	host          string
	roundTripper  http.RoundTripper
	lastFaildTime int64
	*sync.Mutex
}

func NewConn(host string, r http.RoundTripper) *Conn {
	return &Conn{host, r, 0, new(sync.Mutex)}
}

func (p *Conn) GetLocal(xl *xlog.Logger, key []byte) (rc io.ReadCloser, n int, err error) {
	client := &qrpc.Client{&http.Client{Transport: p.roundTripper}}
	p.Setsucceed()
	url := p.host + "/get_local?key=" + base64.URLEncoding.EncodeToString(key)

	resp, err := client.Get(xl, url)
	if err != nil {
		p.SetFailed()
		err = EServerNotAvailable
		return
	}

	defer func() {
		if err != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != 200 {
		if isServerError(resp.StatusCode, err) {
			p.SetFailed()
		}
		if resp.StatusCode == 412 {
			err = EKeyVerifiedError
			return
		}
		if resp.StatusCode == 404 || resp.StatusCode == 612 {
			err = EKeyNotFound
			return
		}
		err = errors.New(resp.Status)
		return
	}

	n = int(resp.ContentLength)
	rc = resp.Body
	return
}

func (p *Conn) PutLocal(xl *xlog.Logger, key []byte, r io.Reader, n int) (err error) {
	client := &qrpc.Client{&http.Client{Transport: p.roundTripper}}
	p.Setsucceed()
	url := p.host + "/put_local?len=" + strconv.Itoa(n)
	url += "&key=" + base64.URLEncoding.EncodeToString(key)

	resp, err := client.PostWith(xl, url, "application/octet-stream", io.LimitReader(r, int64(n)), n)
	if err != nil {
		p.SetFailed()
		err = EServerNotAvailable
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if isServerError(resp.StatusCode, err) {
			p.SetFailed()
		}
		if resp.StatusCode == 412 {
			err = EKeyVerifiedError
			return
		}
		if resp.StatusCode == 404 {
			err = EKeyNotFound
			return
		}
		err = errors.New(resp.Status)
	}
	return
}

func (p *Conn) SetFailed() {
	p.Lock()
	defer p.Unlock()
	p.lastFaildTime = time.Seconds()
}

func (p *Conn) Setsucceed() {
	p.Lock()
	defer p.Unlock()
	p.lastFaildTime = 0
}

func (p *Conn) GetLastFailedTime() int64 {
	p.Lock()
	defer p.Unlock()
	return p.lastFaildTime
}

func isServerError(statusCode int, err error) bool {
	return statusCode >= 500 && statusCode < 600
}
