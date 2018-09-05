package stat

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleApiCall struct {
	Host   string
	Client *rpc.Client
}

func NewHandleApiCall(host string, client *rpc.Client) *HandleApiCall {
	return &HandleApiCall{host, client}
}

func (r HandleApiCall) QueryTime(logger rpc.Logger, req ReqApiCallTimeQuery) (resp RespRtTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	value.Add("type", req.ApiType)
	err = r.Client.Call(logger, &resp, r.Host+"/v2/apicall/query/time?"+value.Encode())
	return
}

func (r HandleApiCall) DomainQueryTime(logger rpc.Logger, req ReqDomainTimeQuery) (resp RespRtTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("domain", req.Domain)
	if req.ApiType != nil {
		value.Add("type", *req.ApiType)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/apicall/domain/query/time?"+value.Encode())
	return
}

func (r HandleApiCall) QueryBuckets(logger rpc.Logger, req ReqApiCallTimeQuery) (resp RespBucketsQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	value.Add("type", req.ApiType)
	err = r.Client.Call(logger, &resp, r.Host+"/v2/apicall/query/buckets?"+value.Encode())
	return
}
