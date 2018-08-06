package ufop

import (
	"net/http"

	"github.com/qiniu/rpc.v1"
)

type Service struct {
	Host string
	Conn rpc.Client
}

func NewService(host string, t http.RoundTripper) *Service {
	return &Service{host, rpc.Client{&http.Client{Transport: t}}}
}

func (ufop Service) RegUfop(l rpc.Logger, ufopname, uapp, owneruid string) (err error) {

	params := map[string][]string{
		"op":    {ufopname},
		"uapp":  {uapp},
		"owner": {owneruid},
	}
	return ufop.Conn.CallWithForm(l, nil, ufop.Host+"/fopg/regufop", params)
}

func (ufop Service) UnregUfop(l rpc.Logger, ufopname, owneruid string) (err error) {

	params := map[string][]string{
		"op":    {ufopname},
		"owner": {owneruid},
	}
	return ufop.Conn.CallWithForm(l, nil, ufop.Host+"/fopg/unregufop", params)
}

type UfopInfo struct {
	Ufop      string `json:"ufopname"`
	Owner     uint32 `json:"uid"`
	Uapp      string `json:"uapp"`
	DiskCache int    `json:"diskcache"`
}

func (ufop Service) InfoUfop(l rpc.Logger, ufopname string) (irs UfopInfo, err error) {

	params := map[string][]string{
		"op": {ufopname},
	}
	err = ufop.Conn.CallWithForm(l, &irs, ufop.Host+"/fopg/ufopinfo", params)
	return
}

type UfopModifyParam struct {
	Ufop      string `json:"ufop"`
	Uapp      string `json:"uapp"`
	DiskCache int    `json:"diskcache"`
}

func (ufop Service) UfopModify(l rpc.Logger, ump *UfopModifyParam) (err error) {

	return ufop.Conn.CallWithJson(l, nil, ufop.Host+"/modifyufop", ump)
}
