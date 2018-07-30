package stat

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type ReqHttpsApiCallTimeQuery struct {
	Uid    uint32   `json:"uid"`
	Bucket *string  `json:"bucket"`
	From   Day      `json:"from"`
	To     Day      `json:"to"`
	P      ShowType `json:"p"`
}

type HandleHttpsApiCall struct {
	Host   string
	Client *rpc.Client
}

func NewHandleHttpsApiCall(host string, client *rpc.Client) *HandleHttpsApiCall {
	return &HandleHttpsApiCall{host, client}
}

func (r HandleHttpsApiCall) QueryTime(logger rpc.Logger, req ReqHttpsApiCallTimeQuery) (resp RespRtTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/apicall/https/query/time?"+value.Encode())
	return
}

func (r HandleHttpsApiCall) QueryBuckets(logger rpc.Logger, req ReqHttpsApiCallTimeQuery) (resp RespBucketsQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/apicall/https/query/buckets?"+value.Encode())
	return
}
