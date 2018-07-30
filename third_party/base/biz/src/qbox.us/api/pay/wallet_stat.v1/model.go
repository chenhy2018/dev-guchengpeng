package wallet_stat

import (
	. "qbox.us/api/pay/pay"
)

type MonthBillStatus int

const (
	MONTH_BILL_STATUS_NEW             MonthBillStatus = 0 // 新建
	MONTH_BILL_STATUS_INFORMED        MonthBillStatus = 1 // 账单通知
	MONTH_BILL_STATUS_DEDUCTED        MonthBillStatus = 2 // 扣费
	MONTH_BILL_STATUS_COMPLAINED      MonthBillStatus = 3 // 申诉
	MONTH_BILL_STATUS_DEDUCT_INFORMED MonthBillStatus = 4 // 扣费通知
	MONTH_BILL_STATUS_PAID            MonthBillStatus = 5 // 已支付
)

const (
	ACCUMU_MONTH_BILL_STATUS_NOT_OUT   MonthBillStatus = 6 //累计月帐单未出账
	ACCUMU_MONTH_BILL_STATUS_CHARGEOFF MonthBillStatus = 7 //累计月帐单出帐
	ACCUMU_MONTH_BILL_STATUS_PAID      MonthBillStatus = 8 //累计月帐单已支付
)

type MonthStatementStatus int

const (
	MONTH_STATEMENT_STATUS_NOT_OUT   MonthStatementStatus = 1
	MONTH_STATEMENT_STATUS_CHARGEOFF MonthStatementStatus = 2 //月对账单出帐
	MONTH_STATEMENT_STATUS_PAID      MonthStatementStatus = 3 //月对账单已支付
)

type DataType string

const (
	DATA_TYPE_SPACE     DataType = "space"
	DATA_TYPE_TRANSFER  DataType = "transfer"
	DATA_TYPE_BANDWIDTH DataType = "bandwidth"
	DATA_TYPE_APIGET    DataType = "api_get"
	DATA_TYPE_APIPUT    DataType = "api_put"
	DATA_TYPE_SERVICE   DataType = "service"
)

type ReqMonth struct {
	Month  Month `json:"month"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
}

type Job string

const (
	MONTHBILL   Job = "monthbill"
	PAY         Job = "pay"
	TRANSACTION Job = "transaction"
)

func (j Job) ToString() string {
	return string(j)
}

type VMoney struct {
	Time  Second `json:"time"`
	Money Money  `json:"money"`
}

type VMoneys []VMoney

func (v VMoneys) Get() Money {
	if len(v) == 0 {
		return 0
	}
	return v[0].Money
}

func (v VMoneys) GetByTime(t Second) Money {
	for _, i := range v {
		if i.Time <= t {
			return i.Money
		}
	}
	return 0
}

func (v VMoneys) PushFront(m VMoney) VMoneys {
	if v == nil {
		v = VMoneys([]VMoney{m})
	} else {
		v = append(VMoneys{m}, v...)
	}
	return v
}

func (v VMoneys) PopFront() VMoneys {
	if len(v) > 0 {
		v = v[1:]
	}
	return v
}
