package qiniuproxy

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"qbox.us/api/authproxy"
	"qbox.us/proxy/api.v2/proto"
)

type callbackInstanceWithProxy struct {
	conn *lb.Client
}

func NewCallbackInstanceWithAuthProxy(hosts []string, proxys []string, timeoutOption authproxy.AuthProxyTimeoutOption, edgeNode bool) proto.CallbackProxy {
	var transport http.RoundTripper
	if edgeNode {
		transport = lb.NewTransport(&lb.TransportConfig{
			DialTimeoutMS: 1000, //onesecond
			RespTimeoutMS: timeoutOption.ProxyGetRespMs,
			Proxys:        proxys,
			ShouldReproxy: authproxy.ShouldReproxy,
		})
	}

	cfg := &lb.Config{
		Hosts:              hosts,
		ShouldRetry:        shouldRetry,
		FailRetryIntervalS: -1,
		TryTimes:           uint32(len(hosts)),
	}

	conn := lb.New(cfg, transport)
	return &callbackInstanceWithProxy{
		conn: conn,
	}

}

func (self *callbackInstanceWithProxy) Callback(l rpc.Logger,
	URLs []string, host, bodyType string, body string,
	config *proto.CallbackConfig) (resp *http.Response, err error) {

	m := url.Values{}
	m["url"] = URLs
	if host != "" {
		m.Add("host", host)
	}
	if config.AccessKey != "" {
		m.Add("ak", config.AccessKey)
	}
	m.Add("timeout", config.Timeout.String())
	m.Add("uid", strconv.Itoa(int(config.Uid)))
	req, err := lb.NewRequest("POST", "/callback?"+m.Encode(), strings.NewReader(body))
	if err != nil {
		return
	}
	if host != "" {
		req.Host = host
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = int64(len(body))
	resp, err = self.conn.Do(l, req)
	if err != nil {
		err = errors.Info(err, "callback via qiniuproxy").Detail(err)
	}
	return
}
