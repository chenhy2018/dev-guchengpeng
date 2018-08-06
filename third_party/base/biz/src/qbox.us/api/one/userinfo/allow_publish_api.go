package userinfo

import (
	"github.com/qiniu/rpc.v2"
	"net/http"
	"qbox.us/servend/proxy_auth"
)

type AllowPublishClient struct {
	URL  string
	Conn rpc.Client
}

func NewAllowPublishClient(host string, t *proxy_auth.Transport) AllowPublishClient {
	var rt http.RoundTripper
	if t != nil {
		rt = t
	}
	client := &http.Client{Transport: rt}
	return AllowPublishClient{
		URL:  host + "/user/allow/publish",
		Conn: rpc.Client{client},
	}
}

func (c AllowPublishClient) Put(l rpc.Logger, allowPublish []string) (err error) {

	err = c.Conn.CallWithJson(l, nil, "PUT", c.URL, allowPublish)
	return
}

func (c AllowPublishClient) Get(l rpc.Logger) (allowPublish []string, err error) {
	err = c.Conn.Call(l, &allowPublish, "GET", c.URL)
	return
}
