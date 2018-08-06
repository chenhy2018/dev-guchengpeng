package bdlocker

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v3"
	"github.com/qiniu/xlog.v1"

	"code.google.com/p/go.net/context"
)

type Config struct {
	Hosts              []string `json:"hosts"`
	DialTimeoutMs      int      `json:"dial_timeout_ms"`
	RespTimeoutMs      int      `json:"resp_timeout_ms"`
	RetryTimeoutMs     int      `json:"retry_timeout_ms"`
	MaxIdleConnPerHost int      `json:"max_idle_conn_per_host"`
}

type Client struct {
	conf  Config
	cli   rpc.Client
	lbcli *lb.Client
}

func (c *Client) unLockIn(xl *xlog.Logger, host string, fh []byte) (err error) {
	URL := fmt.Sprintf("%s/unlock/%s", host, base64.URLEncoding.EncodeToString(fh))
	resp, err := c.cli.Get(xl, URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	return rpc.ResponseError(resp)
}

func (c *Client) existIn(xl *xlog.Logger, host string, fh []byte) (exist bool, err error) {
	URL := fmt.Sprintf("%s/exist/%s", host, base64.URLEncoding.EncodeToString(fh))
	resp, err := c.cli.Get(xl, URL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	code := resp.StatusCode
	if code == 200 || code == 612 {
		return code == 200, nil
	}
	return false, rpc.ResponseError(resp)
}

func (c *Client) Exist(xl *xlog.Logger, fh []byte) (exist bool, err error) {
	type Ret struct {
		host  string
		exist bool
		err   error
	}
	ch := make(chan Ret, len(c.conf.Hosts))
	for _, host := range c.conf.Hosts {
		go func(host string) {
			exist, err := c.existIn(xl.Spawn(), host, fh)
			xl.Debug(host, exist, err, base64.URLEncoding.EncodeToString(fh))
			ch <- Ret{host, exist, err}
		}(host)
	}
	for i := 0; i < len(c.conf.Hosts); i++ {
		ret := <-ch
		if ret.exist {
			return true, nil
		}
		if ret.err != nil {
			err = ret.err
			xl.Warn("check fh in locker failed", ret.host, ret.err)
		}
	}
	return false, err
}

func (c *Client) Lock(xl *xlog.Logger, fh []byte) (err error) {
	var ret interface{}
	path := fmt.Sprintf("/lock/%v", base64.URLEncoding.EncodeToString(fh))
	return c.lbcli.CallWith(xlog.NewContext(context.TODO(), xl), ret, path, "", nil, 0)
}

func (c *Client) UnLock(xl *xlog.Logger, fh []byte) (err error) {
	type Ret struct {
		host string
		err  error
	}
	ch := make(chan Ret, len(c.conf.Hosts))
	for _, host := range c.conf.Hosts {
		go func(host string) {
			err := c.unLockIn(xl.Spawn(), host, fh)
			xl.Debug(host, err, base64.URLEncoding.EncodeToString(fh))
			ch <- Ret{host, err}
		}(host)
	}
	for i := 0; i < len(c.conf.Hosts); i++ {
		ret := <-ch
		if ret.err != nil {
			xl.Warn("check fh in locker failed", ret.host, ret.err)
			return ret.err
		}
	}
	return nil
}

func NewClient(conf Config) (cli *Client) {
	dialT := time.Duration(conf.DialTimeoutMs) * time.Millisecond
	respT := time.Duration(conf.RespTimeoutMs) * time.Millisecond
	rpcCli := rpc.NewClientTimeout(dialT, respT)
	if conf.MaxIdleConnPerHost != 0 {
		rpcCli.Transport.(*http.Transport).MaxIdleConnsPerHost = conf.MaxIdleConnPerHost
	}
	return &Client{
		conf: conf,
		cli:  rpcCli,
		lbcli: lb.New(&lb.Config{
			Http:           rpc.NewClientTimeout(dialT, respT).Client,
			Hosts:          conf.Hosts,
			HostRetrys:     len(conf.Hosts),
			RetryTimeoutMs: conf.RetryTimeoutMs,
		}),
	}
}
