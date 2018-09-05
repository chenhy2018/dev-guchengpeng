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

type HandleBaseBill struct {
	Host   string
	Client *rpc.Client
}

func NewHandleBaseBill(host string, client *rpc.Client) *HandleBaseBill {
	return &HandleBaseBill{host, client}
}

type ModelBaseBill struct {
	ID        string              `json:"id"`
	Uid       uint32              `json:"uid"`
	Field     string              `json:"field"`
	Item      Item                `json:"item"`
	ResGrpIdx int                 `json:"res_grp_idx"`
	Zone      Zone                `json:"zone"`
	Product   Product             `json:"product"`
	FromSec   Second              `json:"from_sec"`
	ToSec     Second              `json:"to_sec"`
	Desc      string              `json:"desc"`
	CreateAt  HundredNanoSecond   `json:"create_at"`
	UpdateAt  HundredNanoSecond   `json:"update_at"`
	Money     Money               `json:"money"`
	Detail    ModelBaseBillDetail `json:"detail"`
	Version   string              `json:"version"`

	// Deprecated
	From Day `json:"from"`
	To   Day `json:"to"`
}

type ModelBaseBillDetail struct {
	Value     int64                    `json:"value"`
	AccuValue int64                    `json:"accu_value"`
	Extras    []map[string]string      `json:"extras,omitempty"`
	Price     P.RespPriceEntry         `json:"price"`
	All       ModelBaseBillDetailAll   `json:"all"`
	Product   ModelBaseBillDetailAll   `json:"product"`
	Group     ModelBaseBillDetailGroup `json:"group"`
	Item      ModelBaseBillDetailItem  `json:"item"`
}

type ModelBaseBillDetailAll struct {
	Rebates []ModelBaseBillRebate `json:"rebates"`
}

type ModelBaseBillDetailGroup struct {
	Discounts []ModelBaseBillDiscount `json:"discounts"`
	Rebates   []ModelBaseBillRebate   `json:"rebates"`
}

type ModelBaseBillDetailItem struct {
	Def       P.ModelItemDef          `json:"def"`
	Base      ModelBaseBillBase       `json:"base"`
	Packages  []ModelBaseBillPackage  `json:"packages"`
	Discounts []ModelBaseBillDiscount `json:"discounts"`
	Rebates   []ModelBaseBillRebate   `json:"rebates"`
	ResPack   ModelBaseBillResPack    `json:"respack"`
}

type ModelBaseBillBase struct {
	Units      []ModelBaseBillBaseUnit       `json:"units"`
	DailyUnits []ModelBaseBillBaseDailyUnits `json:"daily_units"`
}

type ModelBaseBillBaseUnit struct {
	From     int64 `json:"from"`
	To       int64 `json:"to"`
	Value    int64 `json:"value"`     // 实际产生费用部分
	Money    Money `json:"money"`     // 实际费用
	AllValue int64 `json:"all_value"` // 全额使用量，包含未产生最终费用的部分
	AllMoney Money `json:"all_money"` // 全额使用量对应的收入费用
}

type ModelBaseBillBaseDailyUnits struct {
	Day   string                  `json:"day"` // format: 2006-01-02
	Units []ModelBaseBillBaseUnit `json:"units"`
}

type ModelBaseBillPackage struct {
	Price   P.RespBindingItemPackage `json:"price"`
	Value   int64                    `json:"value"`
	Balance int64                    `json:"balance"`
	Reduce  Money                    `json:"reduce"`
	Units   []ModelRangeMoney        `json:"units"`
	Overdue bool                     `json:"overdue"`
}

type ModelRangeMoney struct {
	From   int64 `json:"from"`
	To     int64 `json:"to"`
	Value  int64 `json:"value"`
	Reduce Money `json:"reduce"` // 对应的实际费用
}

type ModelBaseBillDiscount struct {
	Price  P.RespBindingDiscount `json:"price"`
	Before Money                 `json:"before"`
	Change Money                 `json:"change"`
	After  Money                 `json:"after"`
}

type ModelBaseBillRebate struct {
	Price  P.RespBindingRebate `json:"price"`
	Before Money               `json:"before"`
	Change Money               `json:"change"`
	After  Money               `json:"after"`
}

type ModelBaseBillResPack struct {
	NewQuotas    []ModelResPackQuota       `json:"new_quotas"`   // 本次出账新增的额度
	UsedQuotas   []ModelResPackQuota       `json:"used_quotas"`  // 本次出账占用的额度
	Transactions []ModelResPackTransaction `json:"transactions"` // 本次出账关联的额度流水
}

func (r HandleBaseBill) Get(logger rpc.Logger, req ReqID) (resp ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/basebill/get?"+value.Encode())
	return
}

func (r HandleBaseBill) Set(logger rpc.Logger, req ModelBaseBill) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/basebill/set", req)
	return
}

func (r HandleBaseBill) DummySet(logger rpc.Logger, req ModelBaseBill) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/basebill/dummy/set", req)
	return
}

type ReqDummyFlushdb struct{}

func (r HandleBaseBill) DummyFlushdb(logger rpc.Logger, req ReqDummyFlushdb) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/basebill/dummy/flushdb", req)
	return
}

type ReqLastBaseBillInRange struct {
	Uid   uint32 `json:"uid"`
	Item  Item   `json:"item"`
	Zone  Zone   `json:"zone"`
	Start Second `json:"start"`
	End   Second `json:"end"`
}

func (r HandleBaseBill) LastInRange(logger rpc.Logger, req ReqLastBaseBillInRange) (resp ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("item", req.Item.ToString())
	value.Add("zone", req.Zone.String())
	value.Add("start", req.Start.ToString())
	value.Add("end", req.End.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/basebill/last/in/range?"+value.Encode())
	return
}

func (r HandleBaseBill) ListRange(logger rpc.Logger, req ReqUidAndRangeSecond) (resp []ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("product", req.Product.ToString())
	if req.Zone != nil {
		value.Add("zone", (*req.Zone).String())
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/basebill/list/range?"+value.Encode())
	return
}

func (r HandleBaseBill) DummyListRange(logger rpc.Logger, req ReqUidAndRangeSecond) (resp []ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("product", req.Product.ToString())
	if req.Zone != nil {
		value.Add("zone", (*req.Zone).String())
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/basebill/dummy/list/range?"+value.Encode())
	return
}

func (r HandleBaseBill) ListMonthBillCost(logger rpc.Logger, req ReqListMonthBillCost) (data RespUserCost, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("zone", req.Zone.String())
	value.Add("month", req.Month)
	err = r.Client.Call(logger, &data, r.Host+"/v3/basebill/list/month/bill/cost?"+value.Encode())
	return
}
