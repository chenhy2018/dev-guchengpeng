package discover

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

const (
	DefaultHeartbeatSecs       = 5
	DefaultDialTimeoutMs       = 1000
	DefaultRespHeaderTimeoutMs = 3000
)

var nilGetAttrs = func() Attrs { return nil }

type HeartbeatConfig struct {
	Addr          string   `json:"addr"` // 一般用服务配置中的 bind_host
	Name          string   `json:"name"`
	DiscoverHosts []string `json:"discover_hosts"`
	DialTimeoutMs int      `json:"dial_timeout_ms"`
	RespTimeoutMs int      `json:"resp_timeout_ms"`
	HeartbeatSecs int      `json:"heartbeat_secs"`
	GetAttrs      func() Attrs
}

type Heartbeat struct {
	c *Client
	HeartbeatConfig
}

func NewHeartbeat(cfg *HeartbeatConfig) (*Heartbeat, error) {
	if len(cfg.DiscoverHosts) == 0 {
		return nil, errors.New("discover: hosts not specified")
	}
	if cfg.Addr == "" || cfg.Name == "" {
		return nil, errors.New("discover: addr or name not specified")
	}

	if cfg.DialTimeoutMs == 0 {
		cfg.DialTimeoutMs = DefaultDialTimeoutMs
	}
	if cfg.RespTimeoutMs == 0 {
		cfg.RespTimeoutMs = DefaultRespHeaderTimeoutMs
	}
	if cfg.HeartbeatSecs == 0 {
		cfg.HeartbeatSecs = DefaultHeartbeatSecs
	}
	if cfg.GetAttrs == nil {
		cfg.GetAttrs = nilGetAttrs
	}
	tr := rpc.NewTransportTimeout(time.Duration(cfg.DialTimeoutMs)*time.Millisecond, time.Duration(cfg.RespTimeoutMs)*time.Millisecond)
	c := New(cfg.DiscoverHosts, tr)
	go func() {
		l := xlog.NewWith("heartbeat." + strconv.Itoa(os.Getpid()))
		for {
			attrs := cfg.GetAttrs()
			l.Debugf("addr: %s, name: %s, attrs: %v", cfg.Addr, cfg.Name, attrs)
			if err := c.ServiceRegister(l, cfg.Addr, cfg.Name, attrs); err != nil {
				l.Warn("discover.register failed:", err)
			}
			time.Sleep(time.Duration(cfg.HeartbeatSecs) * time.Second)
		}
	}()
	return &Heartbeat{c: c, HeartbeatConfig: *cfg}, nil
}

func (h *Heartbeat) Report(attrs Attrs) error {
	return h.c.ServiceRegister(nil, h.Addr, h.Name, attrs)
}
