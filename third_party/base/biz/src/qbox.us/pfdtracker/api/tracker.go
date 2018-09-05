package api

import (
	"fmt"
	"net/http"
	"time"

	stgapi "qbox.us/pfdstg/api"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
)

var (
	ErrGidECed = httputil.NewError(612, "gid is ECed")
)

const (
	DefaultDialTimeoutMS   = 1500
	DefaultClientTimeoutMS = 3000
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
	if cfg.ClientTimeoutMS == 0 {
		cfg.ClientTimeoutMS = DefaultClientTimeoutMS
	}
	if cfg.TryTimes == 0 {
		cfg.TryTimes = uint32(len(cfg.Hosts))
	}
	if cfg.FailRetryIntervalS == 0 {
		cfg.FailRetryIntervalS = -1
	}
	cfg.ShouldRetry = shouldRetry

	tr := &cfg.Transport
	if tr.DialTimeoutMS == 0 {
		tr.DialTimeoutMS = DefaultDialTimeoutMS
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

type allocRet struct {
	FidBase uint64 `json:"fidBase"`
	N       uint64 `json:"N"`
}

func (self Client) AllocFid(l rpc.Logger) (fidBase, N uint64, err error) {

	var ret allocRet
	path := "/allocFid/"
	err = self.conn.Call(l, &ret, path)
	return ret.FidBase, ret.N, err
}

type lastRet struct {
	FidBase uint64 `json:"fidBase"`
}

func (self Client) LastFid(l rpc.Logger) (fidBase uint64, err error) {

	var ret lastRet
	path := "/lastFid/"
	err = self.conn.Call(l, &ret, path)
	return ret.FidBase, err
}

func (self Client) Bind(l rpc.Logger, egid string, dgid uint32) (err error) {

	path := fmt.Sprintf("/bind/%v/dgid/%v", egid, dgid)
	err = self.conn.Call(l, nil, path)
	return
}

func (self Client) Unbind(l rpc.Logger, egid string) (err error) {

	path := fmt.Sprintf("/unbind/%v", egid)
	err = self.conn.Call(l, nil, path)
	return
}

type ListItem struct {
	Dgid     uint32 `json:"dgid"`
	Egid     string `json:"egid"`
	Readonly int    `json:"readonly"`
}

func (self Client) ListGids(l rpc.Logger, dgid uint32) (ret []ListItem, err error) {

	path := fmt.Sprintf("/list/%d", dgid)
	err = self.conn.Call(l, &ret, path)
	return
}

func (self Client) EcWithGroup(l rpc.Logger, egid string, group string) (err error) {

	path := fmt.Sprintf("/ec/%v/group/%v", egid, group)
	err = self.conn.Call(l, nil, path)
	return
}

func (self Client) Ec(l rpc.Logger, egid string) (err error) {

	path := fmt.Sprintf("/ec/%v", egid)
	err = self.conn.Call(l, nil, path)
	return
}

func (self Client) Ecing(l rpc.Logger, egid string) (err error) {

	path := fmt.Sprintf("/ecing/%v", egid)
	err = self.conn.Call(l, nil, path)
	return
}

func (self Client) EcDel(l rpc.Logger, egid string) (err error) {
	path := fmt.Sprintf("/ecdel/%v", egid)
	err = self.conn.Call(l, nil, path)
	return
}

func (self Client) StateWithGroup(l rpc.Logger, egid string) (group string, dgid uint32, isECed bool, err error) {

	path := fmt.Sprintf("/state/%v", egid)
	var ret struct {
		Dgid  uint32 `json:"dgid"`
		Ec    bool   `json:"ec"`
		Group string `json:"group"`
	}
	err = self.conn.Call(l, &ret, path)
	return ret.Group, ret.Dgid, ret.Ec, err
}

func (self Client) StateEcDel(l rpc.Logger, egid string) (dgid uint32, ecdel bool, err error) {

	path := fmt.Sprintf("/state/%v", egid)
	var ret struct {
		Dgid  uint32 `json:"dgid"`
		EcDel bool   `json:"ecdel"`
	}
	err = self.conn.Call(l, &ret, path)
	return ret.Dgid, ret.EcDel, err
}

func (self Client) State(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error) {

	path := fmt.Sprintf("/state/%v", egid)
	var ret struct {
		Dgid uint32 `json:"dgid"`
		Ec   bool   `json:"ec"`
	}
	err = self.conn.Call(l, &ret, path)
	return ret.Dgid, ret.Ec, err
}

type StateRet struct {
	Dgid   uint32    `json:"dgid" bson:"dgid"`
	EC     bool      `json:"ec" bson:"ec"`
	ECTime time.Time `json:"ectime,omitempty" bson:"ectime,omitempty"`
	Group  string    `json:"group" bson:"group"`
	EcDel  bool      `json:"ecdel" bson:"ecdel"`
	Ecing  int32     `json:"ecing" bson:"ecing"`
}

func (self Client) StateRet(l rpc.Logger, egid string) (ret StateRet, err error) {

	path := fmt.Sprintf("/state/%v", egid)
	err = self.conn.Call(l, &ret, path)
	return
}

func (self Client) SetCompactStats(l rpc.Logger, egid string, stats *stgapi.CompactStats) error {

	path := "/compactstats/set/gid/" + egid
	return self.conn.CallWithJson(l, nil, path, stats)
}

func (self Client) GetCompactStats(l rpc.Logger, egid string) (*stgapi.CompactStats, error) {

	path := "/compactstats/get/gid/" + egid
	var statsVal stgapi.CompactStats
	err := self.conn.Call(l, &statsVal, path)
	return &statsVal, err
}
