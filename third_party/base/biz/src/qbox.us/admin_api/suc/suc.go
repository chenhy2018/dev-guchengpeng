package suc

import (
	"github.com/qiniu/rpc.v1"
	"net/http"
	"strings"
)

// ----------------------------------------------------------------------------

type Service struct {
	host string
	conn rpc.Client
}

func New(host string, transport http.RoundTripper) *Service {

	return &Service{
		host: host,
		conn: rpc.Client{
			&http.Client{Transport: transport},
		},
	}
}

// ----------------------------------------------------------------------------

func (r *Service) Set(l rpc.Logger, grp, key, val string) error {

	body := strings.NewReader(val)
	url := r.host + "/set/" + grp + "/key/" + key
	return r.conn.CallWith(l, nil, url, "application/octet-stream", body, body.Len())
}

func (r *Service) Del(l rpc.Logger, grp, key string) error {

	url := r.host + "/del/" + grp + "/key/" + key
	return r.conn.Call(l, nil, url)
}
