package stat

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleTransfer struct {
	Host   string
	Client *rpc.Client
}

func NewHandleTransfer(host string, client *rpc.Client) *HandleTransfer {
	return &HandleTransfer{host, client}
}

func (r HandleTransfer) QueryTime(logger rpc.Logger, req ReqTimeQuery) (resp RespRtTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/transfer/query/time?"+value.Encode())
	return
}

func (r HandleTransfer) QueryBuckets(logger rpc.Logger, req ReqTimeQuery) (resp RespBucketsQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/transfer/query/buckets?"+value.Encode())
	return
}

func (r HandleTransfer) DomainQueryTime(logger rpc.Logger, req ReqDomainTimeQuery) (resp RespRtTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("domain", req.Domain)
	if req.ApiType != nil {
		value.Add("type", *req.ApiType)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/transfer/domain/query/time?"+value.Encode())
	return
}
