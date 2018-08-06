package wallet_biz

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/rpc.v1"

	"qbox.us/api/pay/pay"
	W "qbox.us/api/pay/wallet.v3"
	"qbox.us/zone"
)

type Service struct {
	Host   string
	Client *rpc.Client
}

func NewService(host string, client *rpc.Client) *Service {
	return &Service{host, client}
}

type ReqGenBaseBill struct {
	Uid          uint32        `json:"uid"`
	Day          pay.Day       `json:"day"`           // format: 20060102
	Products     []pay.Product `json:"products"`      // products string, split by ",", like "storage,evm"
	Zones        string        `json:"zones"`         // zones string, split by ",", like "nb,bc"
	ZoneCodes    []zone.Zone   `json:"zone_codes"`    // zones codes, like "0,1"
	ExceptGroups []pay.Group   `json:"except_groups"` // items string, split by ",", like "common,mps"
	ExceptItems  []pay.Item    `json:"except_items"`  // items string, split by ",", like "space,transfer"
	IsDummy      bool          `json:"is_dummy"`      // indicate whether this operation is simulation. !!Only used for genbasebill
	IsForce      bool          `json:"is_force"`      // 即使当前时间已有账单，也会正常出账，只针对模拟出账有效
}

func (r Service) GenBaseBill(logger rpc.Logger, req ReqGenBaseBill) (bills []W.ModelBaseBill, err error) {
	params := url.Values{}
	params.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	params.Add("day", req.Day.ToString())

	products := make([]string, len(req.Products))
	for i, product := range req.Products {
		products[i] = product.ToString()
	}
	params.Add("products", strings.Join(products, ","))

	if len(req.ZoneCodes) > 0 {
		zoneCodes := make([]string, len(req.ZoneCodes))
		for i, z := range req.ZoneCodes {
			zoneCodes[i] = z.String()
		}
		params.Add("zone_codes", strings.Join(zoneCodes, ","))
	} else {
		params.Add("zones", req.Zones)
	}

	groups := make([]string, len(req.ExceptGroups))
	for i, group := range req.ExceptGroups {
		groups[i] = group.ToString()
	}
	params.Add("except_groups", strings.Join(groups, ","))

	items := make([]string, len(req.ExceptItems))
	for i, item := range req.ExceptItems {
		items[i] = item.ToString()
	}
	params.Add("except_items", strings.Join(items, ","))

	if req.IsDummy {
		params.Add("is_dummy", "true")
	}

	if req.IsForce {
		params.Add("is_force", "true")
	}

	err = r.Client.CallWithForm(logger, &bills, r.Host+"/gen/basebill", params)
	return
}

type ReqGenRtBill struct {
	Uid          uint32
	Start        pay.Second
	End          pay.Second
	Products     []pay.Product
	Zones        []zone.Zone
	ExceptGroups []pay.Group
	ExceptItems  []pay.Item
	IsDummy      bool
	IsForce      bool
}

func (r Service) GenRtBill(logger rpc.Logger, req ReqGenRtBill) (bills []W.ModelBaseBill, err error) {
	params := url.Values{}
	params.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	params.Add("start", req.Start.ToString())
	params.Add("end", req.End.ToString())

	products := make([]string, len(req.Products))
	for i, product := range req.Products {
		products[i] = product.ToString()
	}
	params.Add("products", strings.Join(products, ","))

	zoneCodes := make([]string, len(req.Zones))
	for i, z := range req.Zones {
		zoneCodes[i] = z.String()
	}
	params.Add("zone_codes", strings.Join(zoneCodes, ","))

	groups := make([]string, len(req.ExceptGroups))
	for i, group := range req.ExceptGroups {
		groups[i] = group.ToString()
	}
	params.Add("except_groups", strings.Join(groups, ","))

	items := make([]string, len(req.ExceptItems))
	for i, item := range req.ExceptItems {
		items[i] = item.ToString()
	}
	params.Add("except_items", strings.Join(items, ","))

	if req.IsDummy {
		params.Add("is_dummy", "true")
	}

	if req.IsForce {
		params.Add("is_force", "true")
	}

	err = r.Client.CallWithForm(logger, &bills, r.Host+"/gen/rtbill", params)
	return
}

type ReqGenBaseBillByRange struct {
	Uid          uint32
	Start        pay.Second
	End          pay.Second
	Products     []pay.Product
	Zones        []zone.Zone
	ExceptGroups []pay.Group
	ExceptItems  []pay.Item
	IsDummy      bool
	IsForce      bool
}

func (r Service) GenBasebillByRange(logger rpc.Logger, req ReqGenBaseBillByRange) (bills []W.ModelBaseBill, err error) {
	params := url.Values{}
	params.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	params.Add("start", req.Start.ToString())
	params.Add("end", req.End.ToString())

	products := make([]string, len(req.Products))
	for i, product := range req.Products {
		products[i] = product.ToString()
	}
	params.Add("products", strings.Join(products, ","))

	zoneCodes := make([]string, len(req.Zones))
	for i, z := range req.Zones {
		zoneCodes[i] = z.String()
	}
	params.Add("zone_codes", strings.Join(zoneCodes, ","))

	groups := make([]string, len(req.ExceptGroups))
	for i, group := range req.ExceptGroups {
		groups[i] = group.ToString()
	}
	params.Add("except_groups", strings.Join(groups, ","))

	items := make([]string, len(req.ExceptItems))
	for i, item := range req.ExceptItems {
		items[i] = item.ToString()
	}
	params.Add("except_items", strings.Join(items, ","))

	if req.IsDummy {
		params.Add("is_dummy", "true")
	}

	if req.IsForce {
		params.Add("is_force", "true")
	}

	err = r.Client.CallWithForm(logger, &bills, r.Host+"/gen/basebill/by/range", params)
	return
}

type ReqGenMonthstatement struct {
	Uid   uint32    `json:"uid"`
	Month pay.Month `json:"month"`
}

func (r Service) GenMonthstatement(logger rpc.Logger, req ReqGenMonthstatement) (err error) {
	params := url.Values{}
	params.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	params.Add("month", req.Month.ToString())

	err = r.Client.CallWithForm(logger, nil, r.Host+"/monthstatement/set", params)
	return
}

func (r Service) GenMergeMonthstatement(logger rpc.Logger, req ReqGenMonthstatement) (err error) {
	params := url.Values{}
	params.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	params.Add("month", req.Month.ToString())

	err = r.Client.CallWithForm(logger, nil, r.Host+"/mergemonthstatement/set", params)
	return
}

type ReqDeductProducts struct {
	Uid          uint32        `json:"uid"`
	Day          pay.Day       `json:"day"`           // format: 20060102
	Products     []pay.Product `json:"products"`      // products string, split by ",", like "storage,evm"
	ExceptGroups []pay.Group   `json:"except_groups"` // items string, split by ",", like "common,mps"
	ExceptItems  []pay.Item    `json:"except_items"`  // items string, split by ",", like "space,transfer"
}

func (r Service) DeductProducts(logger rpc.Logger, req ReqDeductProducts) (err error) {
	params := url.Values{}
	params.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	params.Add("day", req.Day.ToString())

	products := make([]string, len(req.Products))
	for i, product := range req.Products {
		products[i] = product.ToString()
	}
	params.Add("products", strings.Join(products, ","))

	groups := make([]string, len(req.ExceptGroups))
	for i, group := range req.ExceptGroups {
		groups[i] = group.ToString()
	}
	params.Add("except_groups", strings.Join(groups, ","))

	items := make([]string, len(req.ExceptItems))
	for i, item := range req.ExceptItems {
		items[i] = item.ToString()
	}
	params.Add("except_items", strings.Join(items, ","))

	err = r.Client.CallWithForm(logger, nil, r.Host+"/deduct/products", params)
	return
}

type ReqJob struct {
	Name   string            `json:"name"`
	Args   map[string]string `json:"args"`
	Vendor string            `json:"vendor"`
}

func (r Service) JobStart(logger rpc.Logger, req ReqJob) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/job/start", req)
	return
}

type RespJobStatus struct {
	Name    string `json:"name"`
	All     int64  `json:"all"`
	Success int64  `json:"success"`
	Failed  int64  `json:"failed"`
	Process int64  `json:"process"`
	Remain  int64  `json:"remain"`
}

func (r Service) JobStatus(logger rpc.Logger, req ReqJob) (resp RespJobStatus, err error) {
	err = r.Client.CallWithJson(logger, &resp, r.Host+"/job/status", req)
	return
}

func (r Service) JobList(logger rpc.Logger) (resp []RespJobStatus, err error) {
	err = r.Client.CallWithJson(logger, &resp, r.Host+"/job/list", nil)
	return
}
