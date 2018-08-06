package v3

import (
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v1"

	. "qbox.us/api/pay/pay"
)

type HandleRebate struct {
	Host   string
	Client *rpc.Client
}

func NewHandleRebate(host string, client *rpc.Client) *HandleRebate {
	return &HandleRebate{host, client}
}

type ModelRebate struct {
	KindInfo
	EffectTime Day   `json:"effect_time"` // 业务有效期，起始时间
	DeadTime   Day   `json:"dead_time"`   // 业务有效期，结束时间
	Days       int   `json:"days"`        // 绑定有效期最长天数
	Max        Money `json:"max"`         // 满max，max为0特例成免费额度
	Rebate     Money `json:"rebate"`      // 返rebate
	Scope      Scope `json:"scope"`       // 作用范围
}

func (r HandleRebate) Add(logger rpc.Logger, req ModelRebate) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/rebate/add", req)
	return
}

func (r HandleRebate) Get(logger rpc.Logger, req ReqID) (resp ModelRebate, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/rebate/get?"+value.Encode())
	return
}

func (r HandleRebate) List(logger rpc.Logger, req ReqList) (resp []ModelRebate, err error) {
	value := url.Values{}
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("type", req.Type)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/rebate/list?"+value.Encode())
	return
}
