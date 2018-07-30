package dc

import (
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	"qbox.us/api"
)

const (
	NoSuchEntry = 404 // 指定的 Entry 不存在
)

var (
	ENoSuchEntry = httputil.NewError(NoSuchEntry, "no such entry")
)

func hintSet(resp *http.Response) bool {
	if resp.StatusCode/100 == 2 {
		return false
	}
	err := rpc.ResponseError(resp)
	return err.Error() == ENoSuchEntry.Error()
}

type Conn struct {
	host string
	conn *rpc.Client

	rangegetSpeed int64
	setSpeed      int64
}

func NewConn(host string, transport http.RoundTripper) *Conn {
	if t, ok := transport.(*http.Transport); ok {
		transport = &Transport{Transport: t}
	}
	return &Conn{
		host: host,
		conn: &rpc.Client{&http.Client{Transport: transport}},
	}
}

func NewConnWithTimeout(host string, options *TimeoutOptions) *Conn {
	tr := &Transport{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   time.Duration(options.DialMs) * time.Millisecond,
				KeepAlive: 30 * time.Second, // same as DefaultTransport
			}).Dial,
			ResponseHeaderTimeout: time.Duration(options.RespMs) * time.Millisecond,
		},
	}
	cli := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(options.ClientMs) * time.Millisecond,
	}
	return &Conn{
		host:          host,
		conn:          &rpc.Client{Client: cli},
		rangegetSpeed: options.RangegetSpeed,
		setSpeed:      options.SetSpeed,
	}
}

type closer interface {
	Close()
}

func (p *Conn) Close() error {
	if c, ok := p.conn.Client.Transport.(closer); ok {
		c.Close()
	}
	return nil
}

const MinimunSpeedTimeout = 500 * time.Millisecond

func (p *Conn) client(Bps int64, length int64) *rpc.Client {
	if Bps <= 0 || length <= 0 {
		return p.conn
	}
	timeout := time.Duration(float64(length) / float64(Bps) * 1e9)
	if timeout < MinimunSpeedTimeout {
		timeout = MinimunSpeedTimeout
	}
	copyCli := *p.conn.Client
	copyCli.Timeout = timeout
	return &rpc.Client{Client: &copyCli}
}

func (p *Conn) Get(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, err error) {
	return p.GetWithHost(xl, p.host, key)
}

func (p *Conn) GetHint(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, hint bool, err error) {
	return p.GetHintWithHost(xl, p.host, key)
}

func (p *Conn) GetWithHost(xl *xlog.Logger, host string, key []byte) (r io.ReadCloser, length int64, err error) {
	r, length, _, err = p.GetHintWithHost(xl, host, key)
	return
}

func (p *Conn) GetHintWithHost(xl *xlog.Logger, host string, key []byte) (r io.ReadCloser, length int64, hint bool, err error) {

	url := host + "/get/" + base64.URLEncoding.EncodeToString(key)
	xl.Debug("dc.get:", url)

	resp, err := p.conn.Get(xl, url)
	if err != nil {
		return
	}

	code := resp.StatusCode
	if code/100 != 2 {
		defer resp.Body.Close()
		err = api.NewError(code)
		if code == 404 {
			err = ENoSuchEntry
			hint = hintSet(resp)
		}
		return
	}
	return resp.Body, resp.ContentLength, hint, nil
}

func (p *Conn) RangeGet(xl *xlog.Logger, key []byte, from int64, to int64) (r io.ReadCloser, length int64, err error) {
	return p.RangeGetWithHost(xl, p.host, key, from, to)
}

func (p *Conn) RangeGetHint(xl *xlog.Logger, key []byte, from int64, to int64) (r io.ReadCloser, length int64, hint bool, err error) {
	return p.RangeGetHintWithHost(xl, p.host, key, from, to)
}

func (p *Conn) RangeGetAndHost(xl *xlog.Logger, key []byte, from int64, to int64) (host string, r io.ReadCloser, length int64, err error) {
	r, length, err = p.RangeGetWithHost(xl, p.host, key, from, to)
	return p.host, r, length, err
}

func (p *Conn) RangeGetWithHost(xl *xlog.Logger, host string, key []byte, from int64, to int64) (r io.ReadCloser, length int64, err error) {
	r, length, _, err = p.RangeGetHintWithHost(xl, host, key, from, to)
	return
}

func (p *Conn) RangeGetHintWithHost(xl *xlog.Logger, host string, key []byte, from int64, to int64) (r io.ReadCloser, length int64, hint bool, err error) {

	url := host + "/rangeGet/" + base64.URLEncoding.EncodeToString(key) + "/" + strconv.FormatInt(from, 10) + "/" + strconv.FormatInt(to, 10)
	xl.Debug("dc.rangeget:", url)

	resp, err := p.client(p.rangegetSpeed, to-from).Get(xl, url)

	if err != nil {
		return
	}

	code := resp.StatusCode
	if code/100 != 2 {
		defer resp.Body.Close()
		err = api.NewError(code)
		if code == 404 {
			err = ENoSuchEntry
			hint = hintSet(resp)
		}
		return
	}
	return resp.Body, resp.ContentLength, hint, nil
}

func (p *Conn) KeyHost(xl *xlog.Logger, key []byte) (host string, err error) {
	host = p.host
	url := p.host + "/get/" + base64.URLEncoding.EncodeToString(key)
	xl.Debug("Conn.KeyHost - url:", url)

	resp, err := p.conn.Head(xl, url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		err = rpc.ResponseError(resp)
	}
	return
}

func (p *Conn) set(xl *xlog.Logger, key []byte, r io.Reader, length int64, checksum []byte) (host string, err error) {

	host = p.host
	url := p.host + "/set/" + base64.URLEncoding.EncodeToString(key)
	if checksum != nil {
		url += "/sha1/" + base64.URLEncoding.EncodeToString(checksum)
	}
	xl.Debug("dc.set:", url)

	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return
	}

	if length < 0 {
		err = errors.New("dc.conn.set error: length < 0")
		return
	}
	req.ContentLength = length

	resp, err := p.client(p.setSpeed, length).Do(xl, req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code/100 != 2 {
		err = api.NewError(code)
	}
	return
}

func (p *Conn) Set(xl *xlog.Logger, key []byte, r io.Reader, length int64) (err error) {
	_, err = p.set(xl, key, r, length, nil)
	return
}

func (p *Conn) SetEx(xl *xlog.Logger, key []byte, r io.Reader, length int64, checksum []byte) (err error) {
	_, err = p.set(xl, key, r, length, checksum)
	return
}

func (p *Conn) SetWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, length int64) (host string, err error) {
	return p.set(xl, key, r, length, nil)
}

func (p *Conn) SetExWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, length int64, checksum []byte) (host string, err error) {
	return p.set(xl, key, r, length, checksum)
}

func (p *Conn) Delete(xl *xlog.Logger, key []byte) (err error) {

	url := p.host + "/delete/" + base64.URLEncoding.EncodeToString(key)
	xl.Debug("dc.delete:", url)

	resp, err := p.conn.PostEx(xl, url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		err = rpc.ResponseError(resp)
	}
	return
}

type Stats struct {
	Gets       int64
	GetMisses  int64
	Sets       int64
	Removes    int64
	GetBytes   int64
	SetBytes   int64
	GetRunning int
	SetRunning int
}
