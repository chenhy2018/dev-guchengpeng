package api

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
)

const (
	MaxDataSize    = 4 * 1024 * 1024 // 4M
	StatusConflict = 409
)

type Client struct {
	transfer    bool
	getCli      rpc.Client
	putCli      rpc.Client
	proxyGetCli rpc.Client
}

type TimeoutOptions struct {
	DialMs         int `json:"dial_ms"`
	GetRespMs      int `json:"get_resp_ms"`
	ProxyGetRespMs int `json:"proxy_get_resp_ms"`
	PutClientMs    int `json:"put_client_ms"`
}

func dialTimeout(ms int) *http.Transport {

	return &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   time.Duration(ms) * time.Millisecond,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}
}

type Options struct {
	TimeoutOptions
	Transfer bool
	Proxys   []string `json:"proxys"`
}

func shouldReproxy(code int, err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), "connecting to proxy") && strings.Contains(err.Error(), "dial tcp")
	}
	return code == http.StatusServiceUnavailable
}

func NewClient(options *Options) Client {
	c := NewWithTimeout(&options.TimeoutOptions)
	c.proxyGetCli = c.getCli
	c.transfer = options.Transfer
	if len(options.Proxys) > 0 {
		proxyTr := lb.NewTransport(&lb.TransportConfig{
			DialTimeoutMS: options.DialMs,
			RespTimeoutMS: options.ProxyGetRespMs,
			Proxys:        options.Proxys,
			ShouldReproxy: shouldReproxy,
		})
		c.proxyGetCli = rpc.Client{Client: &http.Client{Transport: proxyTr}}
	}
	return c
}

func NewWithTimeout(options *TimeoutOptions) Client {

	getTr := dialTimeout(options.DialMs)
	getTr.ResponseHeaderTimeout = time.Duration(options.GetRespMs) * time.Millisecond
	getClient := rpc.Client{Client: &http.Client{Transport: getTr}}

	putTr := dialTimeout(options.DialMs)
	putClient := rpc.Client{Client: &http.Client{Transport: putTr, Timeout: time.Duration(options.PutClientMs) * time.Millisecond}}

	return Client{getCli: getClient, putCli: putClient}
}

// -----------------------------------------------------------------------------

type StgRet struct {
	Ctx string `json:"ctx"`
}

func (p Client) Create(l rpc.Logger, host string, max uint32, r io.Reader, length uint32) (ret StgRet, err error) {

	if length == 0 {
		// see https://github.com/golang/go/issues/20257
		r = nil
	}
	url := fmt.Sprintf("%v/v1/create/%v", host, max)
	err = p.putCli.CallAfterCrcEncoded(l, &ret, url, "application/octet-stream", r, int64(length))
	return
}

// -----------------------------------------------------------------------------

func (p Client) Put(l rpc.Logger, host string, ctx string, off uint32, r io.Reader, length uint32) (ret StgRet, err error) {

	if length == 0 {
		// see https://github.com/golang/go/issues/20257
		r = nil
	}
	url := fmt.Sprintf("%v/v1/put/%v/at/%v", host, ctx, off)
	err = p.putCli.CallAfterCrcEncoded(l, &ret, url, "application/octet-stream", r, int64(length))
	return
}

// -----------------------------------------------------------------------------

func (p Client) Fwd(l rpc.Logger, host, ctx string, r io.Reader, length uint32) error {

	if length == 0 {
		// see https://github.com/golang/go/issues/20257
		r = nil
	}
	url := fmt.Sprintf("%v/v1/fwd/%v", host, ctx)
	return p.putCli.CallAfterCrcEncoded(l, nil, url, "application/octet-stream", r, int64(length))
}

// -----------------------------------------------------------------------------

func (p Client) Get(l rpc.Logger, host, eblock string, from, to uint32) (rc io.ReadCloser, err error) {
	return p.get(l, p.getCli, host, eblock, from, to)
}

func (p Client) ProxyGet(l rpc.Logger, host, eblock string, from, to uint32) (rc io.ReadCloser, err error) {
	return p.get(l, p.proxyGetCli, host, eblock, from, to)
}

func (p Client) get(l rpc.Logger, cli rpc.Client, host, eblock string, from, to uint32) (rc io.ReadCloser, err error) {

	var url string
	if p.transfer {
		url = fmt.Sprintf("%v/v1/transget/%v/from/%v/to/%v", host, eblock, from, to)
	} else {
		url = fmt.Sprintf("%v/v1/get/%v/from/%v/to/%v", host, eblock, from, to)
	}
	resp, err := cli.PostWithCrcCheck(l, url)
	if err != nil {
		return
	}
	if resp.StatusCode/100 != 2 {
		err = rpc.ResponseError(resp)
		resp.Body.Close()
		return
	}
	rc = resp.Body
	return
}

// -----------------------------------------------------------------------------

type StatusRet struct {
	DiskInfo []DiskInfo `json:"disk_info"`
}

type DiskInfo struct {
	Dgid  uint32 `json:"dgid"`
	Path  string `json:"path"`
	Avail int64  `json:"avail"`
	Total int64  `json:"total"`
}

func (p Client) Status(l rpc.Logger, host string) (ret StatusRet, err error) {
	url := fmt.Sprintf("%v/v1/status", host)
	err = p.getCli.GetCall(l, &ret, url)
	return
}
