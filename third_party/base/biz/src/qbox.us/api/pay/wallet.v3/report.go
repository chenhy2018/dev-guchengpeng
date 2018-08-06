package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	"qbox.us/api/pay/pay"
)

type HandleReport struct {
	Host   string
	Client *rpc.Client
}

func NewHandleReport(host string, client *rpc.Client) *HandleReport {
	return &HandleReport{host, client}
}

type ReqItemBillReport struct {
	Item   pay.Item `json:"item"`
	From   int64    `json:"from"`
	To     int64    `json:"to"`
	Uid    uint32   `json:"uid"` // optional
	Marker string   `json:"marker"`
}

type RespItemBillReportEntry struct {
	Uid           uint32    `json:"uid"`
	BillId        string    `json:"bill_id"`
	BillMoney     pay.Money `json:"bill_money"`
	TransactionId string    `json:"transaction_id"`
	UnPaidMoney   pay.Money `json:"unpaid_money"`
}

type RespItemBillReport struct {
	Entries []RespItemBillReportEntry `json:"entries"`
	Marker  string                    `json:"marker"`
}

// WsItemBills query bills of specific item in given time range, along with their transactions
func (r HandleReport) ItemBills(logger rpc.Logger, req ReqItemBillReport) (resp RespItemBillReport, err error) {
	value := url.Values{}

	value.Add("from", strconv.FormatInt(int64(req.From), 10))
	value.Add("to", strconv.FormatInt(int64(req.To), 10))
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("marker", req.Marker)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/report/item/bills?"+value.Encode())
	return
}

type ReqTopMonthstatements struct {
	Month pay.Month `json:"month"`
	Limit int       `json:"limit"`
}

// WsMonthstatementsTopMoney query month statements, sorted by money
func (r HandleReport) MonthstatementsTopMoney(logger rpc.Logger, req ReqTopMonthstatements) (resp []ModelMonthStatementV4, err error) {
	value := url.Values{}

	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("month", req.Month.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/report/monthstatements/top/money?"+value.Encode())
	return
}
