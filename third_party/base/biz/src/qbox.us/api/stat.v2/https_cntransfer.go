package stat

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleHttpsCnTransfer struct {
	Host   string
	Client *rpc.Client
}

func NewHandleHttpsCnTransfer(host string, client *rpc.Client) *HandleHttpsCnTransfer {
	return &HandleHttpsCnTransfer{host, client}
}

func (r HandleHttpsCnTransfer) QueryTime(logger rpc.Logger, req ReqTimeQuery) (resp RespSimpleTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/https/cn/transfer/query/time?"+value.Encode())
	return
}

func (r HandleHttpsCnTransfer) QueryBuckets(logger rpc.Logger, req ReqTimeQuery) (resp RespBucketsQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/https/cn/transfer/query/buckets?"+value.Encode())
	return
}
