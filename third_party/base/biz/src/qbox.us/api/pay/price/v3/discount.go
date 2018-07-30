package v3

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleDiscount struct {
	Host   string
	Client *rpc.Client
}

func NewHandleDiscount(host string, client *rpc.Client) *HandleDiscount {
	return &HandleDiscount{host, client}
}

type ModelDiscount struct {
	KindInfo
	Percent    int   `json:"percent"`     // like 90=90%, 70=70%...
	EffectTime Day   `json:"effect_time"` // 业务有效期，起始时间
	DeadTime   Day   `json:"dead_time"`   // 业务有效期，结束时间
	Days       int   `json:"days"`        // 绑定有效期最长天数
	Scope      Scope `json:"scope"`       // 作用范围
}

func (r HandleDiscount) Add(logger rpc.Logger, req ModelDiscount) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/discount/add", req)
	return
}

func (r HandleDiscount) Get(logger rpc.Logger, req ReqID) (resp ModelDiscount, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/discount/get?"+value.Encode())
	return
}

func (r HandleDiscount) List(logger rpc.Logger, req ReqList) (resp []ModelDiscount, err error) {
	value := url.Values{}
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("type", req.Type)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/discount/list?"+value.Encode())
	return
}
