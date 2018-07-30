package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
	P "qbox.us/api/pay/price/v3"
	. "qbox.us/zone"
)

type HandleBaseBillEvm struct {
	Host   string
	Client *rpc.Client
}

func NewHandleBaseBillEvm(host string, client *rpc.Client) *HandleBaseBillEvm {
	return &HandleBaseBillEvm{host, client}
}

type ModelBaseBillEvm struct {
	ID         string                 `json:"id"`
	Uid        uint32                 `json:"uid"`
	Zone       Zone                   `json:"zone"`
	ResourceId string                 `json:"resource_id"`
	Item       Item                   `json:"item"`
	Key        string                 `json:"key"`
	From       Second                 `json:"from"`
	To         Second                 `json:"to"`
	Desc       string                 `json:"desc"`
	CreateAt   HundredNanoSecond      `json:"create_at"`
	UpdateAt   HundredNanoSecond      `json:"update_at"`
	Money      Money                  `json:"money"`
	Detail     ModelBaseBillEvmDetail `json:"detail"`
	Version    string                 `json:"version"`
}

type ModelBaseBillEvmDetail struct {
	Value     int64                      `json:"value"`
	AccuValue int64                      `json:"accu_value"`
	Price     P.RespEvmPriceEntry        `json:"price"`
	All       ModelBaseBillDetailAll     `json:"all"`
	Group     ModelBaseBillDetailGroup   `json:"group"`
	Item      ModelBaseBillEvmDetailItem `json:"item"`
}

type ModelBaseBillEvmDetailItem struct {
	Base          ModelBaseBillBaseEvm          `json:"base"`
	UpgradeBase   ModelBaseBillBaseEvmUpgrade   `json:"upgrade_base"`
	DowngradeBase ModelBaseBillBaseEvmDowngrade `json:"downgrade_base"`
	RenewBase     ModelBaseBillBaseEvmRenew     `json:"renew_base"`
	Packages      []ModelBaseBillPackageEvm     `json:"packages"`
	Discounts     []ModelBaseBillDiscount       `json:"discounts"`
	Rebates       []ModelBaseBillRebate         `json:"rebates"`
	Type          EvmBasebillBaseType           `json:"type"` // evm basebill base type
}

type ModelBaseBillBaseEvm struct {
	Value    int64 `json:"value"`     // 实际产生费用部分
	Money    Money `json:"money"`     // 实际费用
	AllValue int64 `json:"all_value"` // 全额使用量，包含未产生最终费用的部分
	AllMoney Money `json:"all_money"` // 全额使用量对应的收入费用
	Update   bool  `json:"update"`    // 是否已经update了
}

type ModelBaseBillBaseEvmUpgrade struct {
	PreItem   Item   `json:"pre_item"`
	PreKey    string `json:"pre_key"`
	PrePrice  int64  `json:"pre_price"`
	LeftDays  int64  `json:"left_days"`
	Left      Money  `json:"left"`
	NewDeduct Money  `json:"new_deduct"`
	Deduct    Money  `json:"deduct"`
	At        Second `json:"at"`
}

type ModelBaseBillBaseEvmDowngrade struct {
	PreItem  Item   `json:"pre_item"`
	PreKey   string `json:"pre_key"`
	PrePrice int64  `json:"pre_price"`
	PreEnd   Day    `json:"pre_end"`
	NewEnd   Day    `json:"new_end"`
	At       Second `json:"at"`
}

type ModelBaseBillBaseEvmRenew struct {
	Money  Money  `json:"money"` // 续费费用
	PreEnd Day    `json:"pre_end"`
	NewEnd Day    `json:"new_end"`
	At     Second `json:"at"`
}

type ModelBaseBillPackageEvm struct {
	Price   P.RespBindingItemPackageEvm `json:"price"`
	Value   int64                       `json:"value"`
	Balance int64                       `json:"balance"`
	Reduce  Money                       `json:"reduce"`
	Overdue bool                        `json:"overdue"`
}

func (r HandleBaseBillEvm) Get(logger rpc.Logger, req ReqID) (resp ModelBaseBillEvm, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/basebill/evm/get?"+value.Encode())
	return
}

func (r HandleBaseBillEvm) Set(logger rpc.Logger, req ModelBaseBillEvm) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/basebill/evm/set", req)
	return
}

func (r HandleBaseBillEvm) DummySet(logger rpc.Logger, req ModelBaseBillEvm) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/basebill/evm/dummy/set", req)
	return
}

type ReqDummyFlushdbEvm struct{}

func (r HandleBaseBillEvm) DummyFlushdb(logger rpc.Logger, req ReqDummyFlushdbEvm) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/basebill/evm/dummy/flushdb", req)
	return
}

type ReqLastBaseBillEvmInMonth struct {
	Uid        uint32 `json:"uid"`
	ResourceId string `json:"resource_id"`
	Item       Item   `json:"item"`
	Key        string `json:"key"`
	Zone       Zone   `json:"zone"`
	Date       Day    `json:"date"`
}

func (r HandleBaseBillEvm) LastInMonth(logger rpc.Logger, req ReqLastBaseBillEvmInMonth) (resp ModelBaseBillEvm, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("resource_id", req.ResourceId)
	value.Add("item", req.Item.ToString())
	value.Add("key", req.Key)
	value.Add("zone", req.Zone.String())
	value.Add("date", req.Date.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/basebill/evm/last/in/month?"+value.Encode())
	return
}

func (r HandleBaseBillEvm) ListRange(logger rpc.Logger, req ReqUidAndRange) (resp []ModelBaseBillEvm, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("product", req.Product.ToString())
	if req.Zone != nil {
		value.Add("zone", (*req.Zone).String())
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/basebill/evm/list/range?"+value.Encode())
	return
}

func (r HandleBaseBillEvm) DummyListRange(logger rpc.Logger, req ReqUidAndRange) (resp []ModelBaseBillEvm, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("product", req.Product.ToString())
	if req.Zone != nil {
		value.Add("zone", (*req.Zone).String())
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/basebill/evm/dummy/list/range?"+value.Encode())
	return
}
