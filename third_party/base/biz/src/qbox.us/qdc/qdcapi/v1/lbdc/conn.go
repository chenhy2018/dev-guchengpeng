package lbdc

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"

	"qbox.us/cc/time"
	"qbox.us/rpc"
	"github.com/qiniu/xlog.v1"

	qrpc "github.com/qiniu/rpc.v1"
	ioadmin "qbox.us/admin_api/io"
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

func (p *Conn) ServiceStat() (info ioadmin.CacheInfo, code int, err error) {
	code, err = rpc.DefaultClient.Call(&info, p.host+"/service-stat")
	return
}

// bds的前3个uint16表示bd号，最后一个uint16表示idc号。
// 这里的逻辑被BdClient和LBdClient所共用。
// 概念上认为BdClient后端所连接的是bd，目前实际上是不处理bds的。
func (p *Conn) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) (n int64, err error) {

	client := &qrpc.Client{&http.Client{Transport: p.roundTripper}}
	p.Setsucceed()
	url := p.host + "/get?key=" + base64.URLEncoding.EncodeToString(key)
	url += "&from=" + strconv.Itoa(from)
	url += "&to=" + strconv.Itoa(to)
	url += "&idc=" + strconv.Itoa(int(bds[3]))
	for _, bd := range bds[:3] {
		url += "&bds=" + strconv.FormatUint(uint64(bd), 10)
	}
	xl.Debug("Conn.Get:", url)
	resp, err := client.Get(xl, url)
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
		return
	}
	n, err = io.Copy(w, resp.Body)
	if n == 0 && err == nil {
		err = io.EOF
	}
	return
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
		if resp.StatusCode == 404 {
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

func (p *Conn) Put(xl *xlog.Logger, r io.Reader, n int, verifiedKey []byte, bds [3]uint16) (key []byte, err error) {
	return p.Put2(xl, r, n, verifiedKey, false, bds)
}

func (p *Conn) Put2(xl *xlog.Logger, r io.Reader, n int, verifiedKey []byte, doCache bool, bds [3]uint16) (key []byte, err error) {
	client := &qrpc.Client{&http.Client{Transport: p.roundTripper}}
	p.Setsucceed()
	url := p.host + "/put?len=" + strconv.Itoa(n)
	if verifiedKey != nil {
		url += "&key=" + base64.URLEncoding.EncodeToString(verifiedKey)
	}
	if doCache {
		url += "&cache=1"
	}
	for _, bd := range bds[:3] {
		url += "&bds=" + strconv.FormatUint(uint64(bd), 10)
	}
	xl.Debug("Conn.Put2: url:", url)
	resp, err := client.PostWith(xl, url, "application/octet-stream", io.LimitReader(r, int64(n)), n)
	if err != nil {
		xl.Warn("Conn.Put2: client.PostWith:", url, n, err)
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
		return
	}

	key = make([]byte, 20) // 20 = len of sha-1
	n, err = io.ReadFull(resp.Body, key)
	if err != nil {
		return
	}
	if n != 20 {
		err = EDataError
		return
	}
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
