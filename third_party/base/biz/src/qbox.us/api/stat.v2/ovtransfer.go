package stat

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleOvTransfer struct {
	Host   string
	Client *rpc.Client
}

func NewHandleOvTransfer(host string, client *rpc.Client) *HandleOvTransfer {
	return &HandleOvTransfer{host, client}
}

func (r HandleOvTransfer) QueryTime(logger rpc.Logger, req ReqTimeQuery) (resp RespSimpleTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/transfer/ov/query/time?"+value.Encode())
	return
}

func (r HandleOvTransfer) QueryBuckets(logger rpc.Logger, req ReqTimeQuery) (resp RespBucketsQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/transfer/ov/query/buckets?"+value.Encode())
	return
}
