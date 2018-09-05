package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleReviseBill struct {
	Host   string
	Client *rpc.Client
}

func NewHandleReviseBill(host string, client *rpc.Client) *HandleReviseBill {
	return &HandleReviseBill{host, client}
}

func (r HandleReviseBill) Get(logger rpc.Logger, req ReqID) (resp ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/revisebill/get?"+value.Encode())
	return
}

func (r HandleReviseBill) Set(logger rpc.Logger, req ModelBaseBill) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/revisebill/set", req)
	return
}

func (r HandleReviseBill) LastInRange(logger rpc.Logger, req ReqLastBaseBillInRange) (resp ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("item", req.Item.ToString())
	value.Add("zone", req.Zone.String())
	value.Add("start", req.Start.ToString())
	value.Add("end", req.End.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/revisebill/last/in/range?"+value.Encode())
	return
}

func (r HandleReviseBill) ListRange(logger rpc.Logger, req ReqUidAndRangeSecond) (resp []ModelBaseBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("product", req.Product.ToString())
	if req.Zone != nil {
		value.Add("zone", (*req.Zone).String())
	}
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/revisebill/list/range?"+value.Encode())
	return
}

func (r HandleReviseBill) ListMonthBillCost(logger rpc.Logger, req ReqListMonthBillCost) (data RespUserCost, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("zone", req.Zone.String())
	value.Add("month", req.Month)
	err = r.Client.Call(logger, &data, r.Host+"/v3/revisebill/list/month/bill/cost?"+value.Encode())
	return
}
