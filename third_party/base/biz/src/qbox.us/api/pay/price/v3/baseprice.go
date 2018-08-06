package v3

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v1"

	. "qbox.us/api/pay/pay"
)

type HandleBase struct {
	Host   string
	Client *rpc.Client
}

func NewHandleBase(host string, client *rpc.Client) *HandleBase {
	return &HandleBase{host, client}
}

type ModelBase struct {
	KindInfo
	Items map[Item]ModelItemBasePrice `json:"items"`
}

type ModelResourceGroupList struct {
	Id     string            `json:"id"`
	Type   int               `json:"type"`
	Groups []json.RawMessage `json:"groups"`
}

type ModelResourceGroupListForPortal struct {
	Id     string                        `json:"id"`
	Type   int                           `json:"type"`
	Groups []ModelResourceGroupForPortal `json:"groups"`
}

type ModelResourceGroupForPortal struct {
	Name   string                      `json:"name"`
	Desc   string                      `json:"desc"`
	Price  ModelItemBasePriceForPortal `json:"price"`
	Detail json.RawMessage             `json:"detail"`
}

type ModelItemBasePrice struct {
	Type              ItemBasePriceType      `json:"type"`
	DataType          ItemDataType           `json:"data"`
	CountType         ItemCountType          `json:"count"`
	Unit              int64                  `json:"unit"`
	Price             ModelRangePrice        `json:"price"`
	CumulativeCycle   CumulativeType         `json:"cumulative_cycle"`
	BillPeriodType    BillPeriodType         `json:"bill_period_type"`
	ResourceGroupList ModelResourceGroupList `json:"resource_groups"`
}

type ModelRangePrice struct {
	Type   RangePriceType    `json:"type"`
	Ranges []ModelPriceRange `json:"ranges"`
}

type ModelPriceRange struct {
	Range int64 `json:"range"`
	Price Money `json:"price"`
}

func (r HandleBase) Add(logger rpc.Logger, req ModelBase) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/base/add", req)
	return
}

func (r HandleBase) Update(logger rpc.Logger, req ModelBase) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/base/update", req)
	return
}

// 获取基础价格
func (r HandleBase) Get(logger rpc.Logger, req ReqID) (resp ModelBase, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/base/get?"+value.Encode())
	return
}

func (r HandleBase) List(logger rpc.Logger, req ReqList) (resp []ModelBase, err error) {
	value := url.Values{}
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("type", req.Type)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/base/list?"+value.Encode())
	return
}

func (r HandleBase) ListPrice(logger rpc.Logger, req ReqWithZones) (resp []BaseWithZone, err error) {
	value := url.Values{}
	value.Add("products", req.Products)
	value.Add("zones", req.Zones)
	if req.When != nil {
		value.Add("when", (*req.When).ToString())
	}
	err = r.Client.Call(logger, &resp, r.Host+"/v3/base/list/price?"+value.Encode())
	return
}
