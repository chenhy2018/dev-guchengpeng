package stat

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleFopg struct {
	Host   string
	Client *rpc.Client
}

func NewHandleFopg(host string, client *rpc.Client) *HandleFopg {
	return &HandleFopg{host, client}
}

type ReqFopgValueQuery struct {
	Uid      uint32   `json:"uid"`
	Type     string   `json:"type"`
	PipeName *string  `json:"pipename"`
	PipeType *string  `json:"pipetype"`
	From     Day      `json:"from"`
	To       Day      `json:"to"`
	P        ShowType `json:"p"`
}

type ReqFopgCountQuery struct {
	Uid      uint32   `json:"uid"`
	PipeName *string  `json:"pipename"`
	PipeType *string  `json:"pipetype"`
	From     Day      `json:"from"`
	To       Day      `json:"to"`
	P        ShowType `json:"p"`
}

func (r HandleFopg) QueryValue(logger rpc.Logger, req ReqFopgValueQuery) (resp RespSimpleTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("type", req.Type)
	if req.PipeName != nil {
		value.Add("pipename", *req.PipeName)
	}
	if req.PipeType != nil {
		value.Add("pipetype", *req.PipeType)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/fopg/query/value?"+value.Encode())
	return
}

func (r HandleFopg) QueryGroupbyType(logger rpc.Logger, req ReqFopgCountQuery) (respMap map[string]RespSimpleTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.PipeName != nil {
		value.Add("pipename", *req.PipeName)
	}
	if req.PipeType != nil {
		value.Add("pipetype", *req.PipeType)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &respMap, r.Host+"/v2/fopg/query/groupby/type?"+value.Encode())
	return
}
