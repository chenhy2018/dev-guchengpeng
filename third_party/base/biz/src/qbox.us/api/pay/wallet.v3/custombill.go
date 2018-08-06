package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleCustomBill struct {
	Host   string
	Client *rpc.Client
}

func NewHandleCustomBill(host string, client *rpc.Client) *HandleCustomBill {
	return &HandleCustomBill{host, client}
}

type ModelCustomBill struct {
	ID       string            `json:"id"`
	Uid      uint32            `json:"uid"`
	Product  Product           `json:"product"`
	DeductAt Second            `json:"deduct_at"`
	Money    Money             `json:"money"`
	Name     string            `json:"name"`
	Desc     string            `json:"desc"`
	Detail   string            `json:"detail"`
	CreateAt HundredNanoSecond `json:"create_at"`
	UpdateAt HundredNanoSecond `json:"update_at"`
}

func (r HandleCustomBill) Get(logger rpc.Logger, req ReqID) (resp ModelCustomBill, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/custombill/get?"+value.Encode())
	return
}

func (r HandleCustomBill) Set(logger rpc.Logger, req ModelCustomBill) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/custombill/set", req)
	return
}

func (r HandleCustomBill) ListRange(logger rpc.Logger, req ReqUidAndRangeSecond) (resp []ModelCustomBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("product", req.Product.ToString())
	if req.Zone != nil {
		value.Add("zone", (*req.Zone).String())
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/custombill/list/range?"+value.Encode())
	return
}
