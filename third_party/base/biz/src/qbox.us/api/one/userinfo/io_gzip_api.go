package userinfo

import (
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"github.com/qiniu/rpc.v2"
	"qbox.us/servend/proxy_auth"
)

type IOGZipSettingClient struct {
	Path string
	Conn *lb.Client
}

func NewIOGZipSettingClient(hosts []string, t *proxy_auth.Transport) IOGZipSettingClient {

	cfg := &lb.Config{
		Hosts:    hosts,
		TryTimes: uint32(len(hosts)),
	}

	client := lb.New(cfg, t)
	return IOGZipSettingClient{
		Path: "/user/gzip_mime_types",
		Conn: client,
	}
}

func (c IOGZipSettingClient) Put(l rpc.Logger, mimeTypes map[string]bool) (err error) {
	err = c.Conn.PutWithJson(l, c.Path, mimeTypes)
	return
}

func (c IOGZipSettingClient) Get(l rpc.Logger) (map[string]bool, error) {
	m := make(map[string]bool)
	err := c.Conn.GetCall(l, &m, c.Path)
	return m, err
}
