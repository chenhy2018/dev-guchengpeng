package wallet_stat

import (
	"net/url"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleMonthFinanceStat struct {
	Host   string
	Client *rpc.Client
}

func NewHandleMonthFinanceStat(host string, client *rpc.Client) *HandleMonthFinanceStat {
	return &HandleMonthFinanceStat{host, client}
}

type ModelMonthFinanceStat struct {
	Month          Month `json:"month"`
	Deduct         Money `json:"deduct"`
	Recharge       Money `json:"recharge"`
	Balance        Money `json:"balance"`
	Owe            Money `json:"owe"`
	CurrentOwe     Money `json:"current_owe"`
	CurrentBalance Money `json:"current_balance"`
}

type ReqMonthFinanceStat struct {
	Month Month `json:"month"`
}

func (r HandleMonthFinanceStat) Get(logger rpc.Logger, req ReqMonthFinanceStat) (resp ModelMonthFinanceStat, err error) {
	value := url.Values{}
	value.Add("month", req.Month.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v1/monthfinancestat/get?"+value.Encode())
	return
}

func (r HandleMonthFinanceStat) Set(logger rpc.Logger, req ModelMonthFinanceStat) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v1/monthfinancestat/set", req)
	return
}
