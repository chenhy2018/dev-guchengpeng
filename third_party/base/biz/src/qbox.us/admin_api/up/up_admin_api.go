package up

import (
	"net/http"
	"qbox.us/rpc"
)

// -----------------------------------------------------------

type Service struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

// -----------------------------------------------------------

func (r *Service) Stat(ret interface{}) (code int, err error) {

	return r.Conn.Call(ret, r.Host+"/admin/service-stat")
}

// -----------------------------------------------------------
