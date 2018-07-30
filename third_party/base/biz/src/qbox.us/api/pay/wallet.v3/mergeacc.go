package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleMergeAcc struct {
	Host   string
	Client *rpc.Client
}

func NewHandleMergeAcc(host string, client *rpc.Client) *HandleMergeAcc {
	return &HandleMergeAcc{host, client}
}

type ReqUidAndChildren struct {
	Uid      uint32 `json:"uid"`
	Children string `json:"children"` // split by ",", only available for parent uid
	IsChild  bool   `json:"is_child"` // child or parent uid
}

func (r HandleMergeAcc) InfoWithchildren(logger rpc.Logger, req ReqUidAndChildren) (info ModelInfo, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("children", req.Children)
	value.Add("is_child", strconv.FormatBool(req.IsChild))
	err = r.Client.Call(logger, &info, r.Host+"/v3/mergeacc/info/withchildren?"+value.Encode())
	return
}

type ReqUidTransWithchildren struct {
	Uid         uint32            `json:"uid"`
	Children    string            `json:"children"` // split by ",", only availale for parent uid
	IsChild     bool              `json:"is_child"`
	StartTime   HundredNanoSecond `json:"starttime"`
	EndTime     HundredNanoSecond `json:"endtime"`
	Prefix      string            `json:"prefix"`
	Type        string            `json:"type"`
	Expenses    string            `json:"expenses"`
	IsProcessed *string           `json:"isprocessed"`
	IsHide      *bool             `json:"ishide"`
	Offset      *int64            `json:"offset"`
	Limit       *int64            `json:"limit"`
}

func (r HandleMergeAcc) BillListWithchildren(logger rpc.Logger, req ReqUidTransWithchildren) (resp []ModelTransaction, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("children", req.Children)
	value.Add("is_child", strconv.FormatBool(req.IsChild))
	value.Add("starttime", req.StartTime.ToString())
	value.Add("endtime", req.EndTime.ToString())
	value.Add("prefix", req.Prefix)
	value.Add("type", req.Type)
	value.Add("expenses", req.Expenses)
	if req.IsProcessed != nil {
		value.Add("isprocessed", *req.IsProcessed)
	}
	if req.IsHide != nil {
		value.Add("ishide", strconv.FormatBool(*req.IsHide))
	}
	if req.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*req.Offset), 10))
	}
	if req.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*req.Limit), 10))
	}
	err = r.Client.Call(logger, &resp, r.Host+"/v3/mergeacc/bill/list/withchildren?"+value.Encode())
	return
}

type ReqUidMonthWithChildren struct {
	Uid      uint32 `json:"uid"`
	Month    Month  `json:"month"`
	Limit    int    `json:"limit"`
	Children string `json:"children"` // split by ","
}

type RespUidMonthWithChildren struct {
	Uid   uint32 `json:"uid"`
	Month Second `json:"month"`
}

func (r HandleMergeAcc) MonthsWithchildren(logger rpc.Logger, req ReqUidMonthWithChildren) (resp []RespUidMonthWithChildren, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("children", req.Children)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/mergeacc/months/withchildren?"+value.Encode())
	return
}

type RespMixMonthWithChildren struct {
	Uid         uint32 `json:"uid"`
	Month       Second `json:"month"`
	IsMonthBill bool   `json:"is_monthbill"` // true:monthbill, false:monthstatement
}

func (r HandleMergeAcc) MonthsMixWithchildren(logger rpc.Logger, req ReqUidMonthWithChildren) (resp []RespMixMonthWithChildren, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("children", req.Children)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/mergeacc/months/mix/withchildren?"+value.Encode())
	return
}

type ModelMergeAccMonthStatement struct {
	Id        string                            `json:"id"`
	Uid       uint32                            `json:"uid"`
	Month     Month                             `json:"month"`
	Desc      string                            `json:"desc"`
	Status    int                               `json:"status"`
	Money     Money                             `json:"money"`
	Bills     []ModelMergeAccMonthStatementBill `json:"bills"`
	CreatedAt HundredNanoSecond                 `json:"create_at"`
	UpdateAt  HundredNanoSecond                 `json:"update_at"`
	Version   string                            `json:"version"`
}

type ModelMergeAccMonthStatementBill struct {
	Uid   uint32 `json:"uid"`
	Id    string `json:"id"`
	Money Money  `json:"money"`
}

type ReqMergeaccMonthStatementLister struct {
	Uid    uint32 `json:"uid"` // 0: invalid value
	From   *Month `json:"from"`
	To     *Month `json:"to"`
	Status int    `json:"status"` // 0: invalid value
	Offset int    `json:"offset"` // 0: invalid value
	Limit  int    `json:"limit"`  // 0: invalid value
}

func (r HandleMergeAcc) MonthstatementList(logger rpc.Logger, req ReqMergeaccMonthStatementLister) (resp []ModelMergeAccMonthStatement, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.From != nil {
		value.Add("from", (*req.From).ToString())
	}
	if req.To != nil {
		value.Add("to", (*req.To).ToString())
	}
	value.Add("status", strconv.FormatInt(int64(req.Status), 10))
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/mergeacc/monthstatement/list?"+value.Encode())
	return
}

func (r HandleMergeAcc) MonthstatementGet(logger rpc.Logger, req ReqIDOrMonth) (resp ModelMergeAccMonthStatement, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/mergeacc/monthstatement/get?"+value.Encode())
	return
}

func (r HandleMergeAcc) MonthstatementSet(logger rpc.Logger, req ModelMergeAccMonthStatement) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/mergeacc/monthstatement/set", req)
	return
}
