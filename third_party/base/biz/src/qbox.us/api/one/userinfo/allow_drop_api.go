package userinfo

import (
	"net/http"

	"github.com/qiniu/rpc.v2"
	"qbox.us/servend/proxy_auth"
)

type AllowDropClient struct {
	URL  string
	Conn rpc.Client
}

func NewAllowDropClient(host string, t *proxy_auth.Transport) AllowDropClient {
	client := &http.Client{Transport: t}
	return AllowDropClient{
		URL:  host + "/user/allow/drop",
		Conn: rpc.Client{client},
	}
}

func (c AllowDropClient) PutIPWhitelist(l rpc.Logger, whitelist []string) (err error) {
	err = c.Conn.CallWithJson(l, nil, "PUT", c.URL+"/ipwhitelist", whitelist)
	return
}

func (c AllowDropClient) GetIPWhitelist(l rpc.Logger) (whitelist []string, err error) {
	err = c.Conn.Call(l, &whitelist, "GET", c.URL+"/ipwhitelist")
	return
}
