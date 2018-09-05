package v2

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

type Conn struct {
	host    string
	httpcli rpc.Client
}

func NewConn(host string, t http.RoundTripper) *Conn {
	cli := &http.Client{Transport: t}
	return &Conn{host: host, httpcli: rpc.Client{cli}}
}

func (p *Conn) Get(l rpc.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) (n int64, retry bool, err error) {
	url := p.host + "/get?key=" + base64.URLEncoding.EncodeToString(key)
	url += "&from=" + strconv.Itoa(from)
	url += "&to=" + strconv.Itoa(to)
	url += "&idc=" + strconv.Itoa(int(bds[3]))
	for _, bd := range bds[:3] {
		url += "&bds=" + strconv.FormatUint(uint64(bd), 10)
	}

	xl := xlog.NewWith(l)
	xl.Info("Conn.Get:", url)

	resp, err := p.httpcli.Get(xl, url)
	if err != nil {
		xl.Warn("Conn.Get httpcli.Get failed:", err)
		retry = true
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != 200 {
		if shouldRetry(code) {
			retry = true
		}
		err = errors.New(resp.Status)
		return
	}

	n, err = io.Copy(w, resp.Body)
	return
}

func (p *Conn) Put(l rpc.Logger, r io.Reader, n int, verifiedKey []byte, doCache bool, bds [3]uint16) (key []byte, retry bool, err error) {
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

	xl := xlog.NewWith(l)
	xl.Info("Conn.Put:", url)

	resp, err := p.httpcli.PostWith(xl, url, "application/octet-stream", io.LimitReader(r, int64(n)), n)
	if err != nil {
		xl.Warn("Conn.Put httpcli.PostWith failed:", err)
		retry = true
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != 200 {
		if shouldRetry(code) {
			retry = true
		}
		err = errors.New(resp.Status)
		return
	}

	key = make([]byte, 20)
	_, err = io.ReadFull(resp.Body, key)
	return
}

// 服务端返回 500 和 57X 需要做重试
func shouldRetry(code int) bool {
	if code == 500 || code/10 == 57 {
		return true
	}
	return false
}
