package route

import (
	"net/http"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
)

type Client struct {
	Conn *lb.Client
}

func New(host string, t http.RoundTripper) *Client {
	cfg := &lb.Config{
		Hosts:    []string{host},
		TryTimes: 1,
	}
	client := lb.New(cfg, t)
	return &Client{
		Conn: client,
	}
}

func NewWithMultiHosts(hosts []string, t http.RoundTripper) Client {
	cfg := &lb.Config{
		Hosts:    hosts,
		TryTimes: uint32(len(hosts)),
	}
	client := lb.New(cfg, t)
	return Client{
		Conn: client,
	}
}

func (c *Client) SwitchLine(l rpc.Logger, host, line string) (err error) {
	err = c.Conn.CallWithForm(l, nil, "/route/switchline", map[string][]string{
		"host": []string{host},
		"line": []string{line},
	})
	return
}

type GethostRet struct {
	Addrs []string `json:"addrs"`
}

func (c *Client) GetHost(l rpc.Logger, host string) (ret GethostRet, err error) {
	err = c.Conn.CallWithForm(l, &ret, "/route/gethost", map[string][]string{
		"host": []string{host},
	})
	return
}

type GetAllHostsRet map[string]map[string][]string

func (c *Client) GetAllHosts(l rpc.Logger) (ret GetAllHostsRet, err error) {
	err = c.Conn.Call(l, &ret, "/route/all/hosts")
	return
}
