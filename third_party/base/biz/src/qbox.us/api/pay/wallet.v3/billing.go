package wallet

import (
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

type HandleBilling struct {
	Host   string
	Client *rpc.Client
}

type DailyGetter struct {
	Uid          uint32 `json:"uid"`
	Day          string `json:"day"`           // format: 20060102
	Products     string `json:"products"`      // products string, split by ",", like "storage,evm"
	Zones        string `json:"zones"`         // zones string, split by ",", like "nb,bc"
	ExceptGroups string `json:"except_groups"` // items string, split by ",", like "common,mps"
	ExceptItems  string `json:"except_items"`  // items string, split by ",", like "space,transfer"
	IsDummy      bool   `json:"is_dummy"`      // indicate whether this operation is simulation. Only for billing.
}

func (h *HandleBilling) GenBaseBill(logger rpc.Logger, req DailyGetter) (resp []ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("day", req.Day)
	value.Add("products", req.Products)
	value.Add("zones", req.Zones)
	value.Add("except_groups", req.ExceptGroups)
	value.Add("except_items", req.ExceptItems)
	value.Add("is_dummy", strconv.FormatBool(req.IsDummy))
	err = h.Client.Call(logger, &resp, h.Host+"/gen/basebill?"+value.Encode())
	return

}

func (h *HandleBilling) GenEvmBaseBill(logger rpc.Logger, req DailyGetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("day", req.Day)
	value.Add("products", req.Products)
	value.Add("zones", req.Zones)
	value.Add("except_groups", req.ExceptGroups)
	value.Add("except_items", req.ExceptItems)
	value.Add("is_dummy", strconv.FormatBool(req.IsDummy))
	err = h.Client.Call(logger, nil, h.Host+"/gen/basebill/evm?"+value.Encode())
	return

}
