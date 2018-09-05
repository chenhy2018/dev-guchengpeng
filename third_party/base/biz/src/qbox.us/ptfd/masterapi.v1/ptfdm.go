package masterapi

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
)

const (
	StatusExists = 614
)

type lbConfig struct {
	lb.Config
	Transport lb.TransportConfig `json:"transport"`
}

type Config struct {
	Default  lbConfig `json:"default"`
	Failover lbConfig `json:"failover"`
}

type Client struct {
	conn *lb.Client
}

func shouldRetry(code int, err error) bool {
	if code == http.StatusServiceUnavailable {
		return true
	}
	return lb.ShouldRetry(code, err)
}

func shouldReproxy(code int, err error) bool {
	if code == http.StatusServiceUnavailable {
		return false
	}
	return lb.ShouldReproxy(code, err)
}

func shouldFailover(code int, err error) bool {
	return shouldRetry(code, err) || shouldReproxy(code, err)
}

func setDefaultConfig(cfg *lbConfig) {
	if cfg.TryTimes == 0 {
		cfg.TryTimes = uint32(len(cfg.Hosts))
	}
	if cfg.FailRetryIntervalS == 0 {
		cfg.FailRetryIntervalS = -1
	}
	cfg.ShouldRetry = shouldRetry

	tr := &cfg.Transport
	if tr.DialTimeoutMS == 0 {
		tr.DialTimeoutMS = 1500
	}
	if tr.TryTimes == 0 {
		tr.TryTimes = uint32(len(tr.Proxys))
	}
	if tr.FailRetryIntervalS == 0 {
		tr.FailRetryIntervalS = -1
	}
	tr.ShouldReproxy = shouldReproxy
}

func New(cfg *Config) Client {
	setDefaultConfig(&cfg.Default)
	setDefaultConfig(&cfg.Failover)
	defaultTr := lb.NewTransport(&cfg.Default.Transport)
	failoverTr := lb.NewTransport(&cfg.Failover.Transport)
	if len(cfg.Failover.Hosts) > 0 {
		return Client{lb.NewWithFailover(&cfg.Default.Config, &cfg.Failover.Config, defaultTr, failoverTr, shouldFailover)}
	} else {
		return Client{lb.New(&cfg.Default.Config, defaultTr)}
	}
}

func (self Client) Create(l rpc.Logger, fh []byte, fsize int64, eblocks []string, idc string) (ctime time.Time, err error) {

	var ret struct {
		Ctime time.Time `json:"ctime"`
	}
	path := fmt.Sprintf("/v1/create/%v/fsize/%v/idc/%v", base64.URLEncoding.EncodeToString(fh), fsize, idc)
	err = self.conn.CallWithJson(l, &ret, path, eblocks)
	return ret.Ctime, err
}

func (self Client) Delete(l rpc.Logger, fh []byte) (err error) {

	path := fmt.Sprintf("/v1/delete/%v", base64.URLEncoding.EncodeToString(fh))
	err = self.conn.Call(l, nil, path)
	return
}

type Entry struct {
	Fh      []byte    `json:"fh"`
	Eblocks []string  `json:"eblocks"`
	Fsize   int64     `json:"fsize"`
	Ctime   time.Time `json:"ctime"`
	Idc     string    `json:"idc"`
}

func (self Client) Query(l rpc.Logger, fh []byte) (e *Entry, err error) {

	e = new(Entry)
	path := fmt.Sprintf("/v1/query/%v", base64.URLEncoding.EncodeToString(fh))
	err = self.conn.Call(l, e, path)
	return
}

func (self Client) Transfer(l rpc.Logger, idc string) (e *Entry, err error) {

	e = new(Entry)
	path := "/v1/transfer/" + idc
	err = self.conn.Call(l, e, path)
	return
}

func (self Client) Referenced(l rpc.Logger, dgid, round uint32) (e *Entry, err error) {
	e = new(Entry)
	path := fmt.Sprintf("/v1/referenced/%v/round/%v", dgid, round)
	err = self.conn.Call(l, e, path)
	return
}

type CountRet struct {
	N int `json:"n"`
}

func (self Client) Count(l rpc.Logger, egid string) (n int, err error) {
	path := fmt.Sprintf("/v1/count/%s", egid)
	var ret CountRet
	err = self.conn.Call(l, &ret, path)
	return ret.N, err
}

type CountTransferRets struct {
	OldMgr int `json:"old_mgr"`
	NewMgr int `json:"new_mgr"`
}

func (self Client) TransferStat(l rpc.Logger, min int) (ret CountTransferRets, err error) {
	path := fmt.Sprintf("/v1/transferstat/%d", min)
	err = self.conn.GetCall(l, &ret, path)
	return
}
