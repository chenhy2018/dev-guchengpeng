package api

import (
	"net"
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1/lb.v3"
	"qiniu.com/auth/qiniumac.v1"
)

var (
	defaultTransport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}
)

type Client struct {
	cfg Config
	*lb.Client
}

func New(cfg Config) (client *Client) {
	cfg.Clean()
	if cfg.Transport == nil {
		cfg.Transport = defaultTransport
	}

	p := &Client{
		cfg: cfg,
	}

	mac := qiniumac.Mac{cfg.AccessKey, []byte(cfg.SecretKey)}
	cli := qiniumac.NewClient(&mac, cfg.Transport)

	p.Client = lb.New(&lb.Config{
		Hosts: cfg.Hosts,
		Http:  cli,
	})
	return p
}
