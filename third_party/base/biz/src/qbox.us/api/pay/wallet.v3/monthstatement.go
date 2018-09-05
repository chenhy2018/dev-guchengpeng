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

type ModelMonthStatementV4 struct {
	ID           string                              `json:"id"`
	Uid          uint32                              `json:"uid"`
	Month        Month                               `json:"month"`
	Desc         string                              `json:"desc"`
	CreateAt     HundredNanoSecond                   `json:"create_at"`
	UpdateAt     HundredNanoSecond                   `json:"update_at"`
	Status       int                                 `json:"status"`
	Money        Money                               `json:"money"`
	Cash         Money                               `json:"cash"`
	FM           Money                               `json:"fm"`
	Coupon       Money                               `json:"coupon"`
	Trans        []ModelTransactionInMonth           `json:"trans"`
	Detail       ModelMonthStatementDetailV4         `json:"detail"`
	EvmDetail    ModelMonthStatementDetailEvmV2      `json:"evm_detail"`
	CustomDetail ModelMonthStatementCustomBillDetail `json:"custom_detail"`
	Version      string                              `json:"version"`
}

type ModelMonthStatementDetailV4 struct {
	Rebates   []ModelBaseBillRebate                        `json:"rebates"`
	Discounts []ModelBaseBillDiscount                      `json:"discounts"`
	Products  map[Product]ModelMonthStatementProductDetail `json:"products"`
}

type ModelMonthStatementProductDetail struct {
	Money       Money                                `json:"money"` // before [all] rebates and discounts
	ActualMoney Money                                `json:"actual_money"`
	Rebates     []ModelBaseBillRebate                `json:"rebates"`
	Discounts   []ModelBaseBillDiscount              `json:"discounts"`
	Groups      map[Group]ModelMonthStatementGroupV4 `json:"groups"`
}

type ModelMonthStatementGroupV4 struct {
	Money       Money                              `json:"money"` // before [product, all] rebates and discounts
	ActualMoney Money                              `json:"actual_money"`
	Rebates     []ModelBaseBillRebate              `json:"rebates"`
	Discounts   []ModelBaseBillDiscount            `json:"discounts"`
	Items       map[Item][]ModelMonthStatementItem `json:"items"`
}

type ModelMonthStatementDetailEvmV2 struct {
	Rebates   []ModelBaseBillRebate                   `json:"rebates"`
	Discounts []ModelBaseBillDiscount                 `json:"discounts"`
	Groups    map[Group]ModelMonthStatementGroupEvmV2 `json:"groups"`
}

type ModelMonthStatementGroupEvmV2 struct {
	Money       Money                                                       `json:"money"` // before [all] rebates and discounts
	ActualMoney Money                                                       `json:"actual_money"`
	Rebates     []ModelBaseBillRebate                                       `json:"rebates"`
	Discounts   []ModelBaseBillDiscount                                     `json:"discounts"`
	Items       map[string]map[Item]map[string][]ModelMonthStatementItemEvm `json:"items"` // 1st key: resource_id, 2nd key: item, 3rd key: key
}

type ModelMonthStatementCustomBillDetail struct {
	Bills []ModelCustomBill `json:"bills"`
}

type HandleMonthStatementV4 struct {
	Host   string
	Client *rpc.Client
}

func NewHandleMonthStatementV4(host string, client *rpc.Client) *HandleMonthStatementV4 {
	return &HandleMonthStatementV4{host, client}
}

type HandleMonthStatement struct {
	Host   string
	Client *rpc.Client
}

func NewHandleMonthStatement(host string, client *rpc.Client) *HandleMonthStatement {
	return &HandleMonthStatement{host, client}
}

type ModelMonthStatement struct {
	ID        string                       `json:"id"`
	Uid       uint32                       `json:"uid"`
	Month     Month                        `json:"month"`
	Desc      string                       `json:"desc"`
	CreateAt  HundredNanoSecond            `json:"create_at"`
	UpdateAt  HundredNanoSecond            `json:"update_at"`
	Status    int                          `json:"status"`
	Money     Money                        `json:"money"`
	Cash      Money                        `json:"cash"`
	FM        Money                        `json:"fm"`
	Coupon    Money                        `json:"coupon"`
	Trans     []ModelTransactionInMonth    `json:"trans"`
	Detail    ModelMonthStatementDetail    `json:"detail"`
	EvmDetail ModelMonthStatementDetailEvm `json:"evm_detail"`
	Version   string                       `json:"version"`
}

type ModelTransactionInMonth struct {
	ID     string `json:"id"`
	Desc   string `json:"desc"`
	At     Second `json:"at"`
	Money  Money  `json:"money"`
	Cash   Money  `json:"cash"`
	FM     Money  `json:"fm"`
	Coupon Money  `json:"coupon"`
}

type ModelMonthStatementDetail struct {
	Money       Money                              `json:"money"` // before [all] rebates and discounts
	ActualMoney Money                              `json:"actual_money"`
	Rebates     []ModelBaseBillRebate              `json:"rebates"`
	Discounts   []ModelBaseBillDiscount            `json:"discounts"`
	Groups      map[Group]ModelMonthStatementGroup `json:"groups"`
}

type ModelMonthStatementGroup struct {
	Money       Money                            `json:"money"` // before [product, all] rebates and discounts
	ActualMoney Money                            `json:"actual_money"`
	Rebates     []ModelBaseBillRebate            `json:"rebates"`
	Discounts   []ModelBaseBillDiscount          `json:"discounts"`
	Items       map[Item]ModelMonthStatementItem `json:"items"`
}

type ModelMonthStatementItem struct {
	Money       Money                     `json:"money"` // before [group, product, all] rebates and discounts
	ActualMoney Money                     `json:"actual_money"`
	Rebates     []ModelBaseBillRebate     `json:"rebates"`
	Discounts   []ModelBaseBillDiscount   `json:"discounts"`
	Units       []ModelMonthStatementUnit `json:"units"`
	Zone        Zone                      `json:"zone"`
}

type ModelMonthStatementUnit struct {
	From      Day                     `json:"from"`
	To        Day                     `json:"to"`
	Money     Money                   `json:"money"`
	Value     int64                   `json:"value"`
	Price     P.RespPriceEntry        `json:"price"`
	ResGrpIdx int                     `json:"res_grp_idx"`
	Extras    []map[string]string     `json:"extras,omitempty"`
	Base      ModelBaseBillBase       `json:"base"`
	Packages  []ModelBaseBillPackage  `json:"packages"`
	Discounts []ModelBaseBillDiscount `json:"discounts"`
	Rebates   []ModelBaseBillRebate   `json:"rebates"`
}

type ModelMonthStatementDetailEvm struct {
	Rebates   []ModelBaseBillRebate                 `json:"rebates"`
	Discounts []ModelBaseBillDiscount               `json:"discounts"`
	Groups    map[Group]ModelMonthStatementGroupEvm `json:"groups"`
}

type ModelMonthStatementGroupEvm struct {
	Money       Money                                                     `json:"money"` // before [all] rebates and discounts
	ActualMoney Money                                                     `json:"actual_money"`
	Rebates     []ModelBaseBillRebate                                     `json:"rebates"`
	Discounts   []ModelBaseBillDiscount                                   `json:"discounts"`
	Items       map[string]map[Item]map[string]ModelMonthStatementItemEvm `json:"items"` // 1st key: resource_id, 2nd key: item, 3rd key: key
}

type ModelMonthStatementItemEvm struct {
	Money        Money                               `json:"money"` // before [group, all] rebates and discounts
	ActualMoney  Money                               `json:"actual_money"`
	Rebates      []ModelBaseBillRebate               `json:"rebates"`
	Discounts    []ModelBaseBillDiscount             `json:"discounts"`
	Units        []ModelMonthStatementUnitEvm        `json:"units"`
	UpgradeUnits []ModelMonthStatementUpgradeUnitEvm `json:"upgrade_units"`
	DegradeUnits []ModelMonthStatementDegradeUnitEvm `json:"degrade_units"`
	RenewUnits   []ModelMonthStatementRenewUnitEvm   `json:"renew_units"`
	Zone         Zone                                `json:"zone"`
}

type ModelMonthStatementUnitEvm struct {
	From      Day                       `json:"from"`
	To        Day                       `json:"to"`
	Money     Money                     `json:"money"`
	Value     int64                     `json:"value"`
	Price     P.RespEvmPriceEntry       `json:"price"`
	Base      ModelBaseBillBaseEvm      `json:"base"`
	Packages  []ModelBaseBillPackageEvm `json:"packages"`
	Discounts []ModelBaseBillDiscount   `json:"discounts"`
	Rebates   []ModelBaseBillRebate     `json:"rebates"`
}

type ModelMonthStatementUpgradeUnitEvm struct {
	From      Day                         `json:"from"`
	To        Day                         `json:"to"`
	Money     Money                       `json:"money"`
	Value     int64                       `json:"value"`
	Price     P.RespEvmPriceEntry         `json:"price"`
	Upgrade   ModelBaseBillBaseEvmUpgrade `json:"upgrade"`
	Packages  []ModelBaseBillPackageEvm   `json:"packages"`
	Discounts []ModelBaseBillDiscount     `json:"discounts"`
	Rebates   []ModelBaseBillRebate       `json:"rebates"`
}

type ModelMonthStatementDegradeUnitEvm struct {
	From      Day                           `json:"from"`
	To        Day                           `json:"to"`
	Money     Money                         `json:"money"`
	Value     int64                         `json:"value"`
	Price     P.RespEvmPriceEntry           `json:"price"`
	Degrade   ModelBaseBillBaseEvmDowngrade `json:"degrade"`
	Packages  []ModelBaseBillPackageEvm     `json:"packages"`
	Discounts []ModelBaseBillDiscount       `json:"discounts"`
	Rebates   []ModelBaseBillRebate         `json:"rebates"`
}

type ModelMonthStatementRenewUnitEvm struct {
	From      Day                       `json:"from"`
	To        Day                       `json:"to"`
	Money     Money                     `json:"money"`
	Value     int64                     `json:"value"`
	Price     P.RespEvmPriceEntry       `json:"price"`
	Renew     ModelBaseBillBaseEvmRenew `json:"renew"`
	Packages  []ModelBaseBillPackageEvm `json:"packages"`
	Discounts []ModelBaseBillDiscount   `json:"discounts"`
	Rebates   []ModelBaseBillRebate     `json:"rebates"`
}

type ReqIDOrMonth struct {
	ID    string `json:"id"`
	Uid   uint32 `json:"uid"`
	Month Month  `json:"month"`
}

func (r HandleMonthStatementV4) Get(logger rpc.Logger, req ReqIDOrMonth) (resp ModelMonthStatementV4, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v4/monthstatement/get?"+value.Encode())
	return
}

func (r HandleMonthStatementV4) Set(logger rpc.Logger, req ModelMonthStatementV4) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v4/monthstatement/set", req)
	return
}

type ReqMonthStatementLister struct {
	Uid    *uint32 `json:"uid"`
	From   *Month  `json:"from"`
	To     *Month  `json:"to"`
	Status *int    `json:"status"`
	Offset *int    `json:"offset"`
	Limit  *int    `json:"limit"`
}

func (r HandleMonthStatementV4) List(logger rpc.Logger, req ReqMonthStatementLister) (resp []ModelMonthStatementV4, err error) {
	value := url.Values{}
	if req.Uid != nil {
		value.Add("uid", strconv.FormatUint(uint64(*req.Uid), 10))
	}
	if req.From != nil {
		value.Add("from", (*req.From).ToString())
	}
	if req.To != nil {
		value.Add("to", (*req.To).ToString())
	}
	if req.Status != nil {
		value.Add("status", strconv.FormatInt(int64(*req.Status), 10))
	}
	if req.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*req.Offset), 10))
	}
	if req.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*req.Limit), 10))
	}
	err = r.Client.Call(logger, &resp, r.Host+"/v4/monthstatement/list?"+value.Encode())
	return
}

type ReqUidAndMonth struct {
	Uid   uint32 `json:"uid"`
	Month Month  `json:"month"`
	Limit int    `json:"limit"`
}

func (r HandleMonthStatementV4) Months(logger rpc.Logger, req ReqUidAndMonth) (resp []Second, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v4/monthstatement/months?"+value.Encode())
	return
}

type RespMixMonth struct {
	Month       Second `json:"month"`
	IsMonthBill bool   `json:"is_monthbill"` // true:monthbill, false:monthstatement
}

func (r HandleMonthStatementV4) MonthsMix(logger rpc.Logger, req ReqUidAndMonth) (resp []RespMixMonth, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v4/monthstatement/months/mix?"+value.Encode())
	return
}

func (r HandleMonthStatement) Get(logger rpc.Logger, req ReqIDOrMonth) (resp ModelMonthStatement, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/monthstatement/get?"+value.Encode())
	return
}

func (r HandleMonthStatement) Set(logger rpc.Logger, req ModelMonthStatement) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/monthstatement/set", req)
	return
}

func (r HandleMonthStatement) List(logger rpc.Logger, req ReqMonthStatementLister) (resp []ModelMonthStatement, err error) {
	value := url.Values{}
	if req.Uid != nil {
		value.Add("uid", strconv.FormatUint(uint64(*req.Uid), 10))
	}
	if req.From != nil {
		value.Add("from", (*req.From).ToString())
	}
	if req.To != nil {
		value.Add("to", (*req.To).ToString())
	}
	if req.Status != nil {
		value.Add("status", strconv.FormatInt(int64(*req.Status), 10))
	}
	if req.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*req.Offset), 10))
	}
	if req.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*req.Limit), 10))
	}
	err = r.Client.Call(logger, &resp, r.Host+"/v3/monthstatement/list?"+value.Encode())
	return
}

func (r HandleMonthStatement) Months(logger rpc.Logger, req ReqUidAndMonth) (resp []Second, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/monthstatement/months?"+value.Encode())
	return
}

func (r HandleMonthStatement) MonthsMix(logger rpc.Logger, req ReqUidAndMonth) (resp []RespMixMonth, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/monthstatement/months/mix?"+value.Encode())
	return
}
