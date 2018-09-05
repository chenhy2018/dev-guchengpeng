package wallet_stat

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleUserMonthMoney struct {
	Host   string
	Client *rpc.Client
}

func NewHandleUserMonthMoney(host string, client *rpc.Client) *HandleUserMonthMoney {
	return &HandleUserMonthMoney{host, client}
}

type ModelUserMonthAllMoney struct {
	Uid        uint32  `json:"uid"`
	Month      Month   `json:"month"`
	Deduct     VMoneys `json:"deduct"`
	Recharge   Money   `json:"recharge"`
	FNRecharge Money   `json:"freenb_recharge"`
	Pay        VMoneys `json:"pay"`
	UnPay      VMoneys `json:"unpay"`
	FNPay      VMoneys `json:"freenb_pay"`
	Balance    VMoneys `json:"balance"`
	FNBalance  VMoneys `json:"freenb_balance"`
}

type ModelUserMonthMoney struct {
	Uid        uint32 `json:"uid"`
	Month      Month  `json:"month"`
	Deduct     Money  `json:"deduct"`
	Recharge   Money  `json:"recharge"`
	FNRecharge Money  `json:"freenb_recharge"`
	Pay        Money  `json:"pay"`
	UnPay      Money  `json:"unpay"`
	FNPay      Money  `json:"freenb_pay"`
	Balance    Money  `json:"balance"`
	FNBalance  Money  `json:"freenb_balance"`
}

type ReqUserMonthMoney struct {
	Uid   uint32 `json:"uid"`
	Month Month  `json:"month"`
}

func (r HandleUserMonthMoney) Get(logger rpc.Logger, req ReqUserMonthMoney) (resp ModelUserMonthAllMoney, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v1/usermoney/get?"+value.Encode())
	return
}

type ReqUserMonthMoneyWhen struct {
	Uid   uint32 `json:"uid"`
	Month Month  `json:"month"`
	Time  Day    `json:"time"`
}

func (r HandleUserMonthMoney) GetBytime(logger rpc.Logger, req ReqUserMonthMoneyWhen) (resp ModelUserMonthMoney, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	value.Add("time", req.Time.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v1/usermoney/get/bytime?"+value.Encode())
	return
}

func (r HandleUserMonthMoney) Set(logger rpc.Logger, req ModelUserMonthMoney) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v1/usermoney/set", req)
	return
}

type ReqUserAddPay struct {
	Uid     uint32 `json:"uid"`
	Month   Month  `json:"month"`
	Time    Day    `json:"time"`
	Deduct  Money  `json:"deduct"`
	Pay     Money  `json:"pay"`
	UnPay   Money  `json:"unpay"`
	Balance Money  `json:"balance"`
}

func (r HandleUserMonthMoney) AddPay(logger rpc.Logger, req ReqUserAddPay) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	value.Add("time", req.Time.ToString())
	value.Add("deduct", req.Deduct.ToString())
	value.Add("pay", req.Pay.ToString())
	value.Add("unpay", req.UnPay.ToString())
	value.Add("balance", req.Balance.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v1/usermoney/add/pay", map[string][]string(value))
	return
}

type ReqUserAddFNPay struct {
	Uid       uint32 `json:"uid"`
	Month     Month  `json:"month"`
	Time      Day    `json:"time"`
	FNPay     Money  `json:"freenb_pay"`
	FNBalance Money  `json:"freenb_balance"`
}

func (r HandleUserMonthMoney) AddFNpay(logger rpc.Logger, req ReqUserAddFNPay) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	value.Add("time", req.Time.ToString())
	value.Add("freenb_pay", req.FNPay.ToString())
	value.Add("freenb_balance", req.FNBalance.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v1/usermoney/add/f/npay", map[string][]string(value))
	return
}

func (r HandleUserMonthMoney) List(logger rpc.Logger, req ReqMonth) (resp []ModelUserMonthMoney, err error) {
	value := url.Values{}
	value.Add("month", req.Month.ToString())
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v1/usermoney/list?"+value.Encode())
	return
}
