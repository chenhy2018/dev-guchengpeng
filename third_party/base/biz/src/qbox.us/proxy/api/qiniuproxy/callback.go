package qiniuproxy

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2"
	"qbox.us/proxy/api/proto"
)

type callbackInstance struct {
	conn *lb.Client
}

func NewCallbackInstance(proxyHosts []string) proto.CallbackProxy {

	conn, err := lb.New(proxyHosts, &lb.Config{
		Http:              OneSecondClient,
		TryTimes:          uint32(len(proxyHosts)),
		FailRetryInterval: -1,
		ShouldRetry:       shouldRetry,
	})
	if err != nil {
		panic(err)
	}
	return &callbackInstance{
		conn: conn,
	}
}

func (self *callbackInstance) Callback(l rpc.Logger,
	URLs []string, host, bodyType string, body string,
	accessKey string, config proto.CallbackConfig) (resp *http.Response, err error) {

	m := url.Values{}
	m["url"] = URLs
	if host != "" {
		m.Add("host", host)
	}
	if accessKey != "" {
		m.Add("ak", accessKey)
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
