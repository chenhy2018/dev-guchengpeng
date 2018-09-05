package ufop

import (
	"github.com/qiniu/rpc.v1"
	brpc "github.com/qiniu/rpc.v1/brpc/lb.v2.1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"net/http"
)

type Client struct {
	Host  string     // 功能废弃，兼容保留
	Conn  rpc.Client // 功能废弃，兼容保留
	Bconn *brpc.Client
}

func New(host string, t http.RoundTripper) *Client {
	cfg := &brpc.Config{
		lb.Config{
			Hosts:    []string{host},
			TryTimes: 1,
		},
	}
	client := brpc.New(cfg, t)
	client2 := &http.Client{Transport: t}
	return &Client{
		Host:  host,
		Conn:  rpc.Client{client2},
		Bconn: client,
	}
}

func NewWithMultiHosts(hosts []string, t http.RoundTripper) *Client {
	cfg := &brpc.Config{
		lb.Config{
			Hosts:    hosts,
			TryTimes: uint32(len(hosts)),
		},
	}
	client := brpc.New(cfg, t)
	return &Client{
		Bconn: client,
	}
}

type AclEntry struct {
	Ufop    string   `bson:"ufop"`
	AclMode byte     `bson:"acl_mode"`
	AclList []uint32 `bson:"acl_list"`
	Url     string   `bson:"url"`
	Method  byte     `bson:"method"`
}

type ListRet struct {
	Entries []AclEntry `bson:"entries"`
}

func (p *Client) ListAll(l rpc.Logger) (ret ListRet, err error) {

	err = p.Bconn.CallWithForm(l, &ret, "/listall", nil)
	return
}

type ListUfopsRet struct {
	Ufops []string `bson:"ufops"`
}

func (p *Client) ListUfops(l rpc.Logger) (ret ListUfopsRet, err error) {

	err = p.Bconn.CallWithForm(l, &ret, "/list/ufops", nil)
	return
}
