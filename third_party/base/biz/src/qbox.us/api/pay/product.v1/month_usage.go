package product

import (
	"net/url"
	"strconv"
)

import (
	"time"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/pay/pay"
	. "qbox.us/zone"
)

type HandleMonthUsage struct {
	Host   string
	Client *rpc.Client
}

func NewHandleMonthUsage(host string, client *rpc.Client) *HandleMonthUsage {
	return &HandleMonthUsage{host, client}
}

type ReqMonthUsageGet struct {
	Uid   uint32 `json:"uid"`
	Zone  Zone   `json:"zone"`
	Month int    `json:"month"`
}

type RespMonthUsage struct {
	ID      string `json:"id"`
	Uid     uint32 `json:"uid"`
	Profile struct {
		UType         uint32 `json:"utype"`
		Email         string `json:"email"`
		SalesEmail    string `json:"sales_email"`
		CustomerGroup int    `json:"customer_group"`
		Disabled      bool   `json:"disabled"`
		IsEnterprise  bool   `json:"is_enterprise"` // stduser2
		InternalType  int    `json:"internal_type"`
	} `json:"profile"`
	Zone         Zone                              `json:"zone"`
	Month        int                               `json:"month"` // format: 201603
	Rank         uint                              `json:"rank"`
	Usage        map[pay.Item]RespModelUsage       `json:"usage"`
	Quotas       map[pay.Item]RespModelQuota       `json:"quotas"`
	Consumptions map[pay.Item]RespModelConsumption `json:"consumptions"`
	Money        struct {
		Sum     pay.Money `json:"sum"`
		Fee     pay.Money `json:"fee"`
		Balance pay.Money `json:"balance"`
	} `json:"money"`
	Overflow      bool      `json:"overflow"`
	Freeze        bool      `json:"freeze"`
	NotifiedTimes int       `json:"notified_times"` // 通知次数
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type RespModelUsage struct {
	DataType string  `json:"data"`
	Usage    float64 `json:"usage"`
}

type RespModelQuota struct {
	DataType string `json:"data"`
	Quota    int64  `json:"quota"`
}

type RespModelConsumption struct {
	Product     string    `json:"product"`
	ItemDisplay string    `json:"item_display"`
	Cost        pay.Money `json:"cost"`
}

type ReqMonthUsageSet struct {
	Data string `json:"data"`
}

type ReqMonthUsageDistinctSales struct {
	Zone  Zone `json:"zone"`
	Month int  `json:"month"`
}

type ReqMonthUsageList struct {
	Email         string  `json:"email"`
	Uid           *uint32 `json:"uid"`
	SalesEmail    string  `json:"sales_email"`
	CustomerGroup *int    `json:"customer_group"`
	Disabled      *bool   `json:"disabled"`
	IsEnterprise  *bool   `json:"is_enterprise"`
	InternalType  *int    `json:"internal_type"`
	Zone          *Zone   `json:"zone"`
	Month         int     `json:"month"`
	Overflow      *bool   `json:"overflow"`
	Freeze        *bool   `json:"freeze"`
	Balance       string  `json:"balance"`
	PageSize      int     `json:"page_size"`
	Page          int     `json:"page"`
}

func (r HandleMonthUsage) Get(logger rpc.Logger, params ReqMonthUsageGet) (data RespMonthUsage, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(params.Uid), 10))
	value.Add("zone", params.Zone.String())
	value.Add("month", strconv.FormatInt(int64(params.Month), 10))
	err = r.Client.Call(logger, &data, r.Host+"/v1/month_usage/get?"+value.Encode())
	return
}

func (r HandleMonthUsage) Set(logger rpc.Logger, p ReqMonthUsageSet) (err error) {
	value := url.Values{}
	value.Add("data", p.Data)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v1/month_usage/set", map[string][]string(value))
	return
}

func (r HandleMonthUsage) List(logger rpc.Logger, params ReqMonthUsageList) (data []RespMonthUsage, err error) {
	value := url.Values{}
	value.Add("email", params.Email)
	if params.Uid != nil {
		value.Add("uid", strconv.FormatUint(uint64(*params.Uid), 10))
	}
	value.Add("sales_email", params.SalesEmail)
	if params.CustomerGroup != nil {
		value.Add("customer_group", strconv.FormatInt(int64(*params.CustomerGroup), 10))
	}
	if params.Disabled != nil {
		value.Add("disabled", strconv.FormatBool(*params.Disabled))
	}
	if params.IsEnterprise != nil {
		value.Add("is_enterprise", strconv.FormatBool(*params.IsEnterprise))
	}
	if params.InternalType != nil {
		value.Add("internal_type", strconv.FormatInt(int64(*params.InternalType), 10))
	}
	if params.Zone != nil {
		value.Add("zone", (*params.Zone).String())
	}
	value.Add("month", strconv.FormatInt(int64(params.Month), 10))
	if params.Overflow != nil {
		value.Add("overflow", strconv.FormatBool(*params.Overflow))
	}
	if params.Freeze != nil {
		value.Add("freeze", strconv.FormatBool(*params.Freeze))
	}
	value.Add("balance", params.Balance)
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	value.Add("page", strconv.FormatInt(int64(params.Page), 10))
	err = r.Client.Call(logger, &data, r.Host+"/v1/month_usage/list?"+value.Encode())
	return
}

func (r HandleMonthUsage) DistinctSales(logger rpc.Logger, params ReqMonthUsageDistinctSales) (data []string, err error) {
	value := url.Values{}
	value.Add("zone", params.Zone.String())
	value.Add("month", strconv.FormatInt(int64(params.Month), 10))
	err = r.Client.Call(logger, &data, r.Host+"/v1/month_usage/distinct/sales?"+value.Encode())
	return
}
