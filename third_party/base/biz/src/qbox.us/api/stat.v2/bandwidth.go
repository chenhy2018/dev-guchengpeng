package stat

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleBandwidth struct {
	Host   string
	Client *rpc.Client
}

func NewHandleBandwidth(host string, client *rpc.Client) *HandleBandwidth {
	return &HandleBandwidth{host, client}
}

func (r HandleBandwidth) QueryTime(logger rpc.Logger, req ReqTimeQuery) (resp RespRtTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/bandwidth/query/time?"+value.Encode())
	return
}

func (r HandleBandwidth) QueryTimeAdjustment(logger rpc.Logger, req ReqBandwidthAdjustment) (resp RespRtTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v2/bandwidth/query/time/adjustment?"+value.Encode())
	return
}

func (r HandleBandwidth) QueryBuckets(logger rpc.Logger, req ReqTimeQuery) (resp RespBucketsQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/bandwidth/query/buckets?"+value.Encode())
	return
}

func (r HandleBandwidth) DomainQueryTime(logger rpc.Logger, req ReqTimeQuery) (resp RespRtTimeQuery, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("p", req.P.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/bandwidth/domain/query/time?"+value.Encode())
	return
}
