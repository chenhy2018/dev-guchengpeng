package eu

import (
	"net/http"
	"qbox.us/api/eu"
	"qbox.us/cc/strconv"
	"qbox.us/rpc"
)

// ----------------------------------------------------------

type Service struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

// ----------------------------------------------------------

func (p *Service) GetWatermark(owner uint, customer string) (ret eu.Watermark, code int, err error) {

	params := map[string][]string{
		"id":       {strconv.Uitoa(owner)},
		"customer": {customer},
	}
	code, err = p.Conn.CallWithForm(&ret, p.Host+"/admin/wmget", params)
	return
}
