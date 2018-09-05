package qrefresher

import (
	"time"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"github.com/qiniu/xlog.v1"
)

var ErrSlaveRefreshFailed = errors.New("Call qconf slave.refresh failed")

// ------------------------------------------------------------------------

type Config struct {
	MgrAccessKey  string     `json:"access_key"`
	MgrSecretKey  string     `json:"secret_key"`  // 向 slave 发送指令时的帐号
	SlaveHosts    [][]string `json:"slave_hosts"` // [[idc1_slave1, idc1_slave2], [idc2_slave1, idc2_slave2], ...]
	DialTimeoutMs int        `json:"dial_timeout_ms"`
}

type Client struct {
	lbClients  []*lb.Client
	SlaveHosts [][]string // [[idc1_slave1, idc1_slave2], [idc2_slave1, idc2_slave2], ...]
}

func new(accessKey, secretKey string, slaveHosts [][]string, dialTimeoutMs int) Client {

	mac := &digest.Mac{
		AccessKey: accessKey,
		SecretKey: []byte(secretKey),
	}
	dialTimeout := time.Duration(dialTimeoutMs) * time.Millisecond
	if dialTimeoutMs == 0 {
		dialTimeout = time.Second
	}
	tr := digest.NewTransport(mac, rpc.NewTransportTimeout(dialTimeout, 0))

	lbClients := make([]*lb.Client, len(slaveHosts))
	for i, idcHosts := range slaveHosts {
		lbClients[i] = lb.New(&lb.Config{
			Hosts:              idcHosts,
			FailRetryIntervalS: -1,
			TryTimes:           uint32(len(idcHosts)),
		}, tr)
	}

	return Client{lbClients: lbClients, SlaveHosts: slaveHosts}
}

func New(accessKey, secretKey string, slaveHosts [][]string) Client {
	return new(accessKey, secretKey, slaveHosts, 0)
}

func NewWith(cfg *Config) Client {

	return new(cfg.MgrAccessKey, cfg.MgrSecretKey, cfg.SlaveHosts, cfg.DialTimeoutMs)
}

func (p Client) Refresh(l rpc.Logger, id string) (err error) {

	for i, cli := range p.lbClients {
		err1 := cli.CallWithForm(l, nil, "/refresh", map[string][]string{
			"id": {id},
		})
		if err1 != nil {
			log := xlog.NewWith(l)
			log.Warn("Call qconf slave.refresh failed:", err1, p.SlaveHosts[i])
			err = errors.Info(ErrSlaveRefreshFailed, "Hosts:", p.SlaveHosts[i])
		}
	}
	return
}

func (p Client) MultiRefresh(l rpc.Logger, id ...string) (err error) {

	for i, cli := range p.lbClients {
		err1 := cli.CallWithForm(l, nil, "/refresh", map[string][]string{
			"id": id,
		})
		if err1 != nil {
			log := xlog.NewWith(l)
			log.Warn("Call qconf slave.refresh failed:", err1, p.SlaveHosts[i])
			err = errors.Info(ErrSlaveRefreshFailed, "Hosts:", p.SlaveHosts[i])
		}
	}
	return
}

// ------------------------------------------------------------------------
