package mcm

import (
	"encoding/base64"
	"github.com/qiniu/rpc.v1"
	"net/http"
)

// ----------------------------------------------------------------------------

type Service struct {
	host string
	conn *rpc.Client
}

func New(host string, tr http.RoundTripper) *Service {

	client := &rpc.Client{&http.Client{Transport: tr}}
	return &Service{
		host: host,
		conn: client,
	}
}

// ----------------------------------------------------------------------------

func (r *Service) Del(l rpc.Logger, key string) (err error) {

	url := r.host + "/del/" + base64.URLEncoding.EncodeToString([]byte(key))
	return r.conn.Call(l, nil, url)
}
