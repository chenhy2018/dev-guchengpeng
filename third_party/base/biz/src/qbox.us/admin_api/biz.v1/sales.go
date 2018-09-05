package biz

import (
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

type SalesCustomers struct {
	SaleName     string   `json:"sale_name"`
	CustomerUids []uint32 `json:"customer_uids"`
}

type Sales struct {
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
	Name   string `json:"name"`
}

// 获取销售负责的客户
// offset,limit
// 例如：limit=2, 每次获取2个销售的客户数据
func (s *BizService) GetCustomersBySales(l rpc.Logger, offset, limit int) (salesCustomes []SalesCustomers, err error) {

	offsetStr := strconv.Itoa(offset)
	limitStr := strconv.Itoa(limit)

	param := url.Values{
		"offset": {offsetStr},
		"limit":  {limitStr},
	}

	err = s.rpc.CallWithForm(l, &salesCustomes, s.host+"/sales/customers", param)

	return
}

// GetSales get sales info by sales email or customer's info
// salesEmail, customerEmail and customerUid only required one the them
func (s *BizService) GetSales(l rpc.Logger, salesEmail, customerEmail string, customerUid uint32) (sales Sales, err error) {
	params := url.Values{}
	if salesEmail != "" {
		params["sale_email"] = []string{salesEmail}
	}

	if customerEmail != "" {
		params["customer_email"] = []string{customerEmail}
	}

	if customerUid > 0 {
		params["uid"] = []string{strconv.Itoa(int(customerUid))}
	}

	err = s.rpc.GetCallWithForm(l, &sales, s.host+"/sales/get", params)
	return
}
