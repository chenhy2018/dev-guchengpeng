package v3

import (
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v1"

	. "qbox.us/api/pay/pay"
)

type HandlePackageEvm struct {
	Host   string
	Client *rpc.Client
}

func NewHandlePackageEvm(host string, client *rpc.Client) *HandlePackageEvm {
	return &HandlePackageEvm{host, client}
}

type ModelPackageEvm struct {
	KindInfo                          // 如果是排他性(同时只能有一个item起作用)的package, type请设置为 EVM_PKG_TYPE_UNIQ
	EffectTime Day                    `json:"effect_time"` // 业务有效期，起始时间
	DeadTime   Day                    `json:"dead_time"`   // 业务有效期，结束时间
	Days       int                    `json:"days"`        // 绑定有效期最长天数
	Price      Money                  `json:"price"`       // 套餐价格，为0时特例成免费额度（即原先的reward)
	Quotas     map[Item]ModelQuotaEvm `json:"quotas"`      // key: evm resource id
}

type ModelQuotaEvm struct {
	RecId string `json:"rec_id"`
	Quota int64  `json:"quota"`
}

func (r HandlePackageEvm) Add(logger rpc.Logger, req ModelPackageEvm) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/package/evm/add", req)
	return
}

func (r HandlePackageEvm) Get(logger rpc.Logger, req ReqID) (resp ModelPackageEvm, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/package/evm/get?"+value.Encode())
	return
}

func (r HandlePackageEvm) List(logger rpc.Logger, req ReqList) (resp []ModelPackageEvm, err error) {
	value := url.Values{}
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("type", req.Type)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/package/evm/list?"+value.Encode())
	return
}
