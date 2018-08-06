package wallet_stat

import (
	"net/url"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleMonthbillItem struct {
	Host   string
	Client *rpc.Client
}

func NewHandleMonthbillItem(host string, client *rpc.Client) *HandleMonthbillItem {
	return &HandleMonthbillItem{host, client}
}

type ModelMonthbillItem struct {
	Month Month                      `json:"month"`
	Money Money                      `json:"money"`
	Items map[Group]map[string]Money `json:"items"`
}

type ReqMonthbillItem struct {
	Month Month `json:"month"`
}

func (r HandleMonthbillItem) Get(logger rpc.Logger, req ReqMonthbillItem) (resp ModelMonthbillItem, err error) {
	value := url.Values{}
	value.Add("month", req.Month.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v1/monthbillitem/get?"+value.Encode())
	return
}

func (r HandleMonthbillItem) Set(logger rpc.Logger, req ModelMonthbillItem) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v1/monthbillitem/set", req)
	return
}
