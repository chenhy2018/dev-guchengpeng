package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleRtBill struct {
	Host   string
	Client *rpc.Client
}

func NewHandleRtBill(host string, client *rpc.Client) *HandleRtBill {
	return &HandleRtBill{host, client}
}

func (r HandleRtBill) Get(logger rpc.Logger, req ReqID) (resp ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/rtbill/get?"+value.Encode())
	return
}

func (r HandleRtBill) Set(logger rpc.Logger, req ModelBaseBill) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/rtbill/set", req)
	return
}

func (r HandleRtBill) LastInRange(logger rpc.Logger, req ReqLastBaseBillInRange) (resp ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("item", req.Item.ToString())
	value.Add("zone", req.Zone.String())
	value.Add("start", req.Start.ToString())
	value.Add("end", req.End.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/rtbill/last/in/range?"+value.Encode())
	return
}

func (r HandleRtBill) ListRange(logger rpc.Logger, req ReqUidAndRangeSecond) (resp []ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("product", req.Product.ToString())
	if req.Zone != nil {
		value.Add("zone", (*req.Zone).String())
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/rtbill/list/range?"+value.Encode())
	return
}
