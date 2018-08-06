package stat

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type RespStatInfo struct {
	Space       int64 `json:"space"`
	Space_avg   int64 `json:"space_avg"`
	Bandwidth   int64 `json:"bandwidth"`
	Apicall_get int64 `json:"apicall_get"`
	Apicall_put int64 `json:"apicall_put"`
	Transfer    int64 `json:"transfer"`
	OvTransfer  int64 `json:"ov_transfer"`
}

type RespStatInfoDaily struct {
	Space             int64 `json:"space"`
	Apicall_get       int64 `json:"apicall_get"`
	Apicall_get_Month int64 `json:"apicall_get_month"`
	Apicall_put       int64 `json:"apicall_put"`
	Apicall_put_Month int64 `json:"apicall_put_month"`
	Transfer          int64 `json:"transfer"`
	Transfer_Month    int64 `json:"transfer_month"`
}

type HandleInfo struct {
	Host   string
	Client *rpc.Client
}

func NewHandleInfo(host string, client *rpc.Client) *HandleInfo {
	return &HandleInfo{host, client}
}

func (r HandleInfo) Month(logger rpc.Logger, req ReqMonthInfo) (resp RespStatInfo, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("month", req.Month.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/info/month?"+value.Encode())
	return
}

func (r HandleInfo) Day(logger rpc.Logger, req ReqDayInfo) (resp RespStatInfoDaily, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Bucket != nil {
		value.Add("bucket", *req.Bucket)
	}
	value.Add("day", req.Day.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v2/info/day?"+value.Encode())
	return
}
