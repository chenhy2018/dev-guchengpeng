package api

import (
	"errors"
	"net/http"

	"github.com/qiniu/rpc.v1"
	cfgapi "qbox.us/pfdcfg/api"
	stgapi "qbox.us/pfdstg/api"
	"qbox.us/pfdtracker/stater"
	"qbox.us/qconf/qconfapi"
)

var (
	EGetDGInfoFailed = errors.New("get dg info from pfdcfg failed.")
)

type Config struct {
	Guid                string          `json:"guid"`
	CfgHosts            []string        `json:"cfg_hosts"`
	TrackerQconf        qconfapi.Config `json:"tracker_qconf"`
	Proxies             []string        `json:"proxies"`
	PutTryTimes         int             `json:"put_try_times"`
	DeleteTryTimes      int             `json:"delete_try_times"`
	DgsRefreshIntervalS int             `json:"dgs_refresh_interval_s"`
	SmallFileLimit      int64           `json:"small_file_limit"`
	Idc                 string          `json:"idc"`
	RemoteIdcOrder      []string        `json:"remote_idc_order"`
	MaxIdleConnPerHost  int             `json:"max_idle_conn_per_host"`

	Timeouts stgapi.TimeoutOption `json:"timeouts"`
}

func New(cfg *Config) (c *Client, err error) {

	cfgcli1, err := cfgapi.New(cfg.CfgHosts, nil)
	if err != nil {
		return
	}

	cfgcli := PfdCfgClient{cfgcli1}

	gidStater := stater.NewGidStater(&cfg.TrackerQconf)
	c, err = NewClientWithConfig(cfg, cfgcli, gidStater)
	if cfg.MaxIdleConnPerHost != 0 {
		stgapi.PostTimeoutTransport.(*http.Transport).MaxIdleConnsPerHost = cfg.MaxIdleConnPerHost
		stgapi.GetTimeoutTransport.(*http.Transport).MaxIdleConnsPerHost = cfg.MaxIdleConnPerHost
	}
	if c != nil {
		c.cfg = cfg
	}
	return
}

type PfdCfgClient struct {
	client cfgapi.Client
}

func (p PfdCfgClient) ListDgs(l rpc.Logger, guid string) ([]*cfgapi.DiskGroupInfo, error) {
	return p.client.AllDgs(l, guid)
}

func (p PfdCfgClient) GetDGInfo(l rpc.Logger, guid string, dgid uint32) (dgInfo *cfgapi.DiskGroupInfo, err error) {

	dgInfo, err = p.client.DGInfo(l, guid, dgid)
	if err != nil {
		return nil, EGetDGInfoFailed
	}
	return
}
