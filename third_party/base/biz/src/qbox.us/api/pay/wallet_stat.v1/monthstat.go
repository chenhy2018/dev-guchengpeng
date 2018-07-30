package wallet_stat

import (
	"net/url"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleMonthStat struct {
	Host   string
	Client *rpc.Client
}

func NewHandleMonthStat(host string, client *rpc.Client) *HandleMonthStat {
	return &HandleMonthStat{host, client}
}

type ModelMonthStatAll struct {
	Month      Month   `json:"month"`
	Deduct     Money   `json:"deduct"`
	Recharge   Money   `json:"recharge"`
	DeductCash VMoneys `json:"deduct_cash"`
	Balance    VMoneys `json:"balance"`
	Owe        VMoneys `json:"owe"`
	CurrentOwe VMoneys `json:"current_owe"`
}

type ModelMonthStat struct {
	Month      Month `json:"month"`
	Deduct     Money `json:"deduct"`
	Recharge   Money `json:"recharge"`
	DeductCash Money `json:"deduct_cash"`
	Balance    Money `json:"balance"`
	Owe        Money `json:"owe"`
	CurrentOwe Money `json:"current_owe"`
}

type ReqMonthStat struct {
	Month Month `json:"month"`
}

func (r HandleMonthStat) Get(logger rpc.Logger, req ReqMonthStat) (resp ModelMonthStatAll, err error) {
	value := url.Values{}
	value.Add("month", req.Month.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v1/monthstat/get?"+value.Encode())
	return
}

type ReqMonthStatWhen struct {
	Month Month `json:"month"`
	Time  Day   `json:"time"`
}

func (r HandleMonthStat) GetBytime(logger rpc.Logger, req ReqMonthStatWhen) (resp ModelMonthStat, err error) {
	value := url.Values{}
	value.Add("month", req.Month.ToString())
	value.Add("time", req.Time.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v1/monthstat/get/bytime?"+value.Encode())
	return
}

func (r HandleMonthStat) Set(logger rpc.Logger, req ModelMonthStat) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v1/monthstat/set", req)
	return
}

type ReqMonthStatAddPay struct {
	Month      Month `json:"month"`
	Time       Day   `json:"time"`
	DeductCash Money `json:"deduct_cash"`
	Owe        Money `json:"owe"`
	Balance    Money `json:"balance"`
	CurrentOwe Money `json:"current_owe"`
}

func (r HandleMonthStat) AddPay(logger rpc.Logger, req ReqMonthStatAddPay) (err error) {
	value := url.Values{}
	value.Add("month", req.Month.ToString())
	value.Add("time", req.Time.ToString())
	value.Add("deduct_cash", req.DeductCash.ToString())
	value.Add("owe", req.Owe.ToString())
	value.Add("balance", req.Balance.ToString())
	value.Add("current_owe", req.CurrentOwe.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v1/monthstat/add/pay", map[string][]string(value))
	return
}
