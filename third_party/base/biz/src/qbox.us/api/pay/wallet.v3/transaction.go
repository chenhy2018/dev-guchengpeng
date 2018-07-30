package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleTransaction struct {
	Host   string
	Client *rpc.Client
}

func NewHandleTransaction(host string, client *rpc.Client) *HandleTransaction {
	return &HandleTransaction{host, client}
}

type ReqTrGet struct {
	Uid    uint32 `json:"uid"`
	Excode string `json:"excode"`
	Prefix string `json:"prefix"`
	Type   string `json:"type"`
}

func (r HandleTransaction) Get(logger rpc.Logger, req ReqTrGet) (resp ModelTransaction, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("excode", req.Excode)
	value.Add("prefix", req.Prefix)
	value.Add("type", req.Type)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/transaction/get?"+value.Encode())
	return
}

type ReqTrGetbysn struct {
	SerialNum string `json:"serial_num"`
}

func (r HandleTransaction) GetBysn(logger rpc.Logger, req ReqTrGetbysn) (resp ModelTransaction, err error) {
	value := url.Values{}
	value.Add("serial_num", req.SerialNum)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/transaction/get/bysn?"+value.Encode())
	return
}

type ReqTrList struct {
	Uid         uint32            `json:"uid"`       // 0: all users
	StartTime   HundredNanoSecond `json:"starttime"` // 0: ignore
	EndTime     HundredNanoSecond `json:"endtime"`   // 0: ignore
	Prefix      string            `json:"prefix"`
	Type        string            `json:"type"`
	Expenses    *bool             `json:"expenses"`
	IsProcessed *bool             `json:"isprocessed"`
	IsHide      *bool             `json:"ishide"`
	Offset      int64             `json:"offset"`
	Limit       int64             `json:"limit"`
}

func (r HandleTransaction) List(logger rpc.Logger, req ReqTrList) (resp []ModelTransaction, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("starttime", req.StartTime.ToString())
	value.Add("endtime", req.EndTime.ToString())
	value.Add("prefix", req.Prefix)
	value.Add("type", req.Type)
	if req.Expenses != nil {
		value.Add("expenses", strconv.FormatBool(*req.Expenses))
	}
	if req.IsProcessed != nil {
		value.Add("isprocessed", strconv.FormatBool(*req.IsProcessed))
	}
	if req.IsHide != nil {
		value.Add("ishide", strconv.FormatBool(*req.IsHide))
	}
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/transaction/list?"+value.Encode())
	return
}

type ReqRechargeList struct {
	Uid       uint32            `json:"uid"`       // 0: all users
	StartTime HundredNanoSecond `json:"starttime"` // 0: ignore
	EndTime   HundredNanoSecond `json:"endtime"`   // 0: ignore
	Type      string            `json:"type"`
	IsHide    *bool             `json:"ishide"`
	Offset    int64             `json:"offset"`
	Limit     int64             `json:"limit"`
}

func (r HandleTransaction) ListRecharge(logger rpc.Logger, req ReqRechargeList) (resp []ModelTransaction, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("starttime", req.StartTime.ToString())
	value.Add("endtime", req.EndTime.ToString())
	value.Add("type", req.Type)
	if req.IsHide != nil {
		value.Add("ishide", strconv.FormatBool(*req.IsHide))
	}
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/transaction/list/recharge?"+value.Encode())
	return
}

type HandleTransactionV4 struct {
	Host   string
	Client *rpc.Client
}

func NewHandleTransactionV4(host string, client *rpc.Client) *HandleTransactionV4 {
	return &HandleTransactionV4{host, client}
}

type ReqTrListUids struct {
	StartTime HundredNanoSecond `json:"starttime"`
	EndTime   HundredNanoSecond `json:"endtime"`
}

func (r HandleTransactionV4) List(logger rpc.Logger, req ReqTrList) (resp []ModelTransactionV4, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("starttime", req.StartTime.ToString())
	value.Add("endtime", req.EndTime.ToString())
	value.Add("prefix", req.Prefix)
	value.Add("type", req.Type)

	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v4/transaction/list?"+value.Encode())
	return
}

func (r HandleTransactionV4) ListUids(logger rpc.Logger, req ReqTrListUids) (uids []uint32, err error) {
	value := url.Values{}
	value.Add("starttime", req.StartTime.ToString())
	value.Add("endtime", req.EndTime.ToString())
	err = r.Client.Call(logger, &uids, r.Host+"/v4/transaction/list/uids?"+value.Encode())
	return
}

type ReqListByExcodes struct {
	Excodes    []string `json:"excode"`
	GotDetails bool     `json:"got_details"`
}

func (r HandleTransactionV4) ListByExcodes(logger rpc.Logger, req ReqListByExcodes) (resp []ModelTransactionV4, err error) {
	value := url.Values{}
	for _, v := range req.Excodes {
		value.Add("excode", v)
	}
	value.Add("got_details", strconv.FormatBool(req.GotDetails))
	err = r.Client.Call(logger, &resp, r.Host+"/v4/transaction/list/by/excodes?"+value.Encode())
	return
}

type ReqTrListbySN struct {
	SerialNums []string
	GotDetails bool
}

func (r HandleTransactionV4) ListBySerialNums(logger rpc.Logger, req ReqTrListbySN) (resp []ModelTransactionV4, err error) {
	value := url.Values{}
	for _, n := range req.SerialNums {
		value.Add("serial_nums", n)
	}
	value.Add("got_details", strconv.FormatBool(req.GotDetails))
	err = r.Client.GetCall(logger, &resp, r.Host+"/v4/transaction/list/bysn?"+value.Encode())
	return
}
