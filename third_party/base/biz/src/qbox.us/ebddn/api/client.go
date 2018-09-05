package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"time"

	route "qbox.us/api/one/route"
	"qbox.us/ebd/api/types"
	pfdcfg "qbox.us/pfdcfg/api"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"github.com/qiniu/xlog.v1"
)

// ===========================================================

const DefaultExpires = 20 * time.Minute
const DefaultFailRetryIntervalS = 30

type Config struct {
	Hosts           []string `json:"hosts"`
	Proxies         []string `json:"proxies"`
	FailoverHosts   []string `json:"failover_hosts"`
	FailoverProxies []string `json:"failover_proxies"`
	OneHost         []string `json:"one_hosts"`

	AccessKey          string `json:"access"`
	SecretKey          string `json:"secret"`
	DnsResolve         bool   `json:"dns_resolve"`
	DnsCacheTimeS      int64  `json:"dns_cache_time_s"`
	DialTimeoutMS      int    `json:"dial_timeout_ms"`
	RespTimeoutMS      int    `json:"resp_timeout_ms"`
	FailRetryIntervalS int64  `json:"fail_retry_interval_s`

	SpeedLimit      lb.SpeedLimit `json:"speed_limit"`
	MaxFails        int           `json:"max_fails"`
	MaxFailsPeriodS int64         `json:"max_fails_period_s"`
	NoBlocks        bool          `json:"no_blocks"`
	NoCached        bool          `json:"no_cached"`
}

type Client struct {
	conn *lb.Client
}

var shouldRetry = func(code int, err error) bool {
	if code == httputil.StatusGracefulQuit || code == httputil.StatusOverload {
		return true
	}
	return lb.ShouldRetry(code, err)
}

var shouleFailover = func(code int, err error) bool {
	return shouldRetry(code, err) || lb.ShouldReproxy(code, err)
}

func New(cfg *Config) (c *Client, err error) {

	if cfg.FailRetryIntervalS == 0 {
		cfg.FailRetryIntervalS = DefaultFailRetryIntervalS
	}
	if len(cfg.FailoverHosts) == 0 {
		cfg.FailRetryIntervalS = -1
	}

	var lookupHost func(host string) (addrs []string, err error) = net.LookupHost
	if len(cfg.OneHost) > 0 {
		routeCli := route.NewWithMultiHosts(cfg.OneHost, nil)
		xl := xlog.NewDummy()
		lookupHost = func(host string) (addrs []string, err error) {
			ret, err := routeCli.GetHost(xl, host)
			if err != nil {
				xl.Info("routeCli.GetHost failed", err)
				return net.LookupHost(host)
			} else {
				addrs = ret.Addrs
			}
			return
		}
	}

	main := &lb.Config{
		Hosts:              cfg.Hosts,
		FailRetryIntervalS: cfg.FailRetryIntervalS,
		ShouldRetry:        shouldRetry,
		DnsResolve:         cfg.DnsResolve,
		DnsCacheTimeS:      cfg.DnsCacheTimeS,
		LookupHost:         lookupHost,
		SpeedLimit:         cfg.SpeedLimit,
		MaxFails:           cfg.MaxFails,
		MaxFailsPeriodS:    cfg.MaxFailsPeriodS,

		LookupHostNotHoldHost: true,
	}
	mainTransport := lb.NewTransport(&lb.TransportConfig{
		DialTimeoutMS:      cfg.DialTimeoutMS,
		RespTimeoutMS:      cfg.RespTimeoutMS,
		Proxys:             cfg.Proxies,
		FailRetryIntervalS: cfg.FailRetryIntervalS,
	})
	failover := &lb.Config{
		Hosts:       cfg.FailoverHosts,
		ShouldRetry: shouldRetry,
	}
	failoverTransport := lb.NewTransport(&lb.TransportConfig{
		DialTimeoutMS: cfg.DialTimeoutMS,
		RespTimeoutMS: cfg.RespTimeoutMS,
		Proxys:        cfg.FailoverProxies,
	})
	if cfg.AccessKey != "" {
		mac := &digest.Mac{
			AccessKey: cfg.AccessKey,
			SecretKey: []byte(cfg.SecretKey),
		}
		mainTransport = digest.NewTransport(mac, mainTransport)
		failoverTransport = digest.NewTransport(mac, failoverTransport)
	}
	conn := lb.NewWithFailover(main, failover, mainTransport, failoverTransport, shouleFailover)
	return &Client{conn: conn}, nil
}

type nullReadCloser struct{}

func (self nullReadCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (self nullReadCloser) Close() error {
	return nil
}

func (self *Client) Get(l rpc.Logger,
	fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {

	fhi, err := types.DecodeFh(fh)
	if err != nil {
		err = errors.Info(err, "types.DecodeFh")
		return
	}

	if from >= to || fhi.Fsize == 0 {
		return nullReadCloser{}, 0, nil
	}

	u := fmt.Sprintf("/get/%s?e=%d", base64.URLEncoding.EncodeToString(fh), time.Now().Add(DefaultExpires).Unix())
	req, err := lb.NewRequest("GET", u, nil)
	if err != nil {
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", from, to-1))

	resp, err := self.conn.Do(l, req)
	if err != nil {
		return
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}
	return resp.Body, resp.ContentLength, nil
}

func (self *Client) Hash(l rpc.Logger, fh []byte) (hash [20]byte, err error) {
	ret := struct {
		Hash string `json:"hash"`
	}{}
	u := fmt.Sprintf("/hash/%s?e=%d", base64.URLEncoding.EncodeToString(fh), time.Now().Add(DefaultExpires).Unix())
	err = self.conn.Call(l, &ret, u)
	if err != nil {
		return
	}
	hash0, err := base64.URLEncoding.DecodeString(ret.Hash)
	copy(hash[:], hash0)
	return
}

func (self *Client) Md5(l rpc.Logger, fh []byte) (md5 []byte, err error) {
	ret := struct {
		Md5 string `json:"md5"`
	}{}
	u := fmt.Sprintf("/md5/%s?e=%d", base64.URLEncoding.EncodeToString(fh), time.Now().Add(DefaultExpires).Unix())
	err = self.conn.Call(l, &ret, u)
	if err != nil {
		return
	}
	md5, err = base64.URLEncoding.DecodeString(ret.Md5)
	return
}

func (c *Client) GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error) {
	return pfdcfg.DEFAULT, nil
}

// -----------------------------------------------------
