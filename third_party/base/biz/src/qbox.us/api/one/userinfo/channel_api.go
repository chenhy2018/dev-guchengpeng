package userinfo

import (
	"net/http"

	"github.com/qiniu/rpc.v2"
	"qbox.us/servend/proxy_auth"
)

type ChannelClient struct {
	URL  string
	Conn rpc.Client
}

func NewChannelClient(host string, t *proxy_auth.Transport) ChannelClient {
	var rt http.RoundTripper
	if t != nil {
		rt = t
	}
	client := &http.Client{Transport: rt}
	return ChannelClient{
		URL:  host + "/user/channel",
		Conn: rpc.Client{client},
	}
}

func (c ChannelClient) Put(l rpc.Logger, channels []string) (err error) {

	err = c.Conn.CallWithJson(l, nil, "PUT", c.URL, channels)
	return
}

func (c ChannelClient) Get(l rpc.Logger) (channels []string, err error) {
	err = c.Conn.Call(l, &channels, "GET", c.URL)
	return
}
