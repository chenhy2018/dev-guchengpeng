package mq

import (
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

type Service struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

func (p *Service) Make(l rpc.Logger, mqId string, expires int) (err error) {
	url := p.Host + "/admin-make/" + mqId + "/expires/" + strconv.Itoa(expires)
	return p.Conn.Call(l, nil, url)
}

func (p *Service) Update(l rpc.Logger, mqId string, expires int) (err error) {
	url := p.Host + "/admin-update/" + mqId + "/expires/" + strconv.Itoa(expires)
	return p.Conn.Call(l, nil, url)
}

func (p *Service) FilterMsgs(l rpc.Logger, mqId string, byUid uint32, toMqId string) (err error) {
	url := p.Host + "/admin-filter/" + mqId + "/by/" + strconv.FormatUint(uint64(byUid), 10) + "/to/" + toMqId
	return p.Conn.Call(l, nil, url)
}

func (p *Service) Stat(l rpc.Logger, mqId string) (ret map[string]int, err error) {
	url := p.Host + "/admin-stat/" + mqId
	err = p.Conn.Call(l, &ret, url)
	return
}
