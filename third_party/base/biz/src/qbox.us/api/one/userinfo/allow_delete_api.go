package userinfo

import (
	"net/http"

	"github.com/qiniu/rpc.v2"
	"qbox.us/servend/proxy_auth"
)

type AllowDeleteClient struct {
	URL  string
	Conn rpc.Client
}

func NewAllowDeleteClient(host string, t *proxy_auth.Transport) AllowDeleteClient {
	client := &http.Client{Transport: t}
	return AllowDeleteClient{
		URL:  host + "/user/allow/delete",
		Conn: rpc.Client{client},
	}
}

func (c AllowDeleteClient) PutIPWhitelist(l rpc.Logger, whitelist []string) (err error) {
	err = c.Conn.CallWithJson(l, nil, "PUT", c.URL+"/ipwhitelist", whitelist)
	return
}

func (c AllowDeleteClient) GetIPWhitelist(l rpc.Logger) (whitelist []string, err error) {
	err = c.Conn.Call(l, &whitelist, "GET", c.URL+"/ipwhitelist")
	return
}
