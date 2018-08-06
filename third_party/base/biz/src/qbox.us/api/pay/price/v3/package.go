package v3

import (
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v1"

	. "qbox.us/api/pay/pay"
)

type HandlePackage struct {
	Host   string
	Client *rpc.Client
}

func NewHandlePackage(host string, client *rpc.Client) *HandlePackage {
	return &HandlePackage{host, client}
}

type ModelPackage struct {
	KindInfo
	EffectTime Day                 `json:"effect_time"` // 业务有效期，起始时间
	DeadTime   Day                 `json:"dead_time"`   // 业务有效期，结束时间
	Days       int                 `json:"days"`        // 绑定有效期最长天数
	Price      Money               `json:"price"`       // 套餐价格，为0时特例成免费额度（即原先的reward)
	Quotas     map[Item]ModelQuota `json:"quotas"`
}

type ModelQuota struct {
	DataType ItemDataType `json:"data"`
	Quota    int64        `json:"quota"`
}

func (r HandlePackage) Add(logger rpc.Logger, req ModelPackage) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/package/add", req)
	return
}

func (r HandlePackage) Get(logger rpc.Logger, req ReqID) (resp ModelPackage, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/package/get?"+value.Encode())
	return
}

func (r HandlePackage) List(logger rpc.Logger, req ReqList) (resp []ModelPackage, err error) {
	value := url.Values{}
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("type", req.Type)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/package/list?"+value.Encode())
	return
}
