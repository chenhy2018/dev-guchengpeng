package stat

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleRank struct {
	Host   string
	Client *rpc.Client
}

func NewHandleRank(host string, client *rpc.Client) *HandleRank {
	return &HandleRank{host, client}
}

type ReqBasicArg struct {
	Month Month `json:"month"`
	Skip  uint  `json:"skip,default"`
	Limit uint  `json:"limit,default"`
}

type ReqApicallArg struct {
	Month Month  `json:"month"`
	Skip  uint   `json:"skip,default"`
	Limit uint   `json:"limit,default"`
	Type  string `json:"type"`
}

type RespEntry struct {
	Uid   []uint32 `json:"uid"`
	Data  []int64  `json:"data"`
	Start uint     `json:"start"`
	End   uint     `json:"end"`
}

func (r HandleRank) Traffic(logger rpc.Logger, args ReqBasicArg) (ret RespEntry, err error) {
	value := url.Values{}
	value.Add("month", args.Month.ToString())
	value.Add("skip,default", strconv.FormatUint(uint64(args.Skip), 10))
	value.Add("limit,default", strconv.FormatUint(uint64(args.Limit), 10))
	err = r.Client.Call(logger, &ret, r.Host+"/v2/rank/traffic?"+value.Encode())
	return
}

func (r HandleRank) Space(logger rpc.Logger, args ReqBasicArg) (ret RespEntry, err error) {
	value := url.Values{}
	value.Add("month", args.Month.ToString())
	value.Add("skip,default", strconv.FormatUint(uint64(args.Skip), 10))
	value.Add("limit,default", strconv.FormatUint(uint64(args.Limit), 10))
	err = r.Client.Call(logger, &ret, r.Host+"/v2/rank/space?"+value.Encode())
	return
}

// 带宽数据的月排名从流量的月排名计算出来
// 取月平均带宽, 单位 bit/s
func (r HandleRank) Bandwidth(logger rpc.Logger, args ReqBasicArg) (ret RespEntry, err error) {
	value := url.Values{}
	value.Add("month", args.Month.ToString())
	value.Add("skip,default", strconv.FormatUint(uint64(args.Skip), 10))
	value.Add("limit,default", strconv.FormatUint(uint64(args.Limit), 10))
	err = r.Client.Call(logger, &ret, r.Host+"/v2/rank/bandwidth?"+value.Encode())
	return
}

func (r HandleRank) Apicall(logger rpc.Logger, args ReqApicallArg) (ret RespEntry, err error) {
	value := url.Values{}
	value.Add("month", args.Month.ToString())
	value.Add("skip,default", strconv.FormatUint(uint64(args.Skip), 10))
	value.Add("limit,default", strconv.FormatUint(uint64(args.Limit), 10))
	value.Add("type", args.Type)
	err = r.Client.Call(logger, &ret, r.Host+"/v2/rank/apicall?"+value.Encode())
	return
}
