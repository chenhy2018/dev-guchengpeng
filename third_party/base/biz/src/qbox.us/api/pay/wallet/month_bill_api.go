package wallet

/*
MonthBillDetails doc: http://docs.qbox.me/bi-pay-public-api#toc13
*/
import (
	"launchpad.net/mgo/bson"
	Url "net/url"
	"qbox.us/rpc"
	"qbox.us/servend/account"
	"strconv"
	"time"
)

type MonthMoneyBill struct {
	Last    int64 `json:"last"`
	Add     int64 `json:"add"`
	Reduce  int64 `json:"reduce"`
	Current int64 `json:"current"`
}

type BillStatus int

const (
	BILL_STATUS_NEW             BillStatus = 0
	BILL_STATUS_INFORMED        BillStatus = 1
	BILL_STATUS_DEDUCTED        BillStatus = 2
	BILL_STATUS_COMPLAINED      BillStatus = 3
	BILL_STATUS_DEDUCT_INFORMED BillStatus = 4
	BILL_STATUS_PAID            BillStatus = 5
)

const (
	ACCUMU_BILL_STATUS_NOT_OUT   BillStatus = 6 //累计月帐单未出
	ACCUMU_BILL_STATUS_CHARGEOFF BillStatus = 7 //累计月帐单出帐
	ACCUMU_BILL_STATUS_PAID      BillStatus = 8 //累计月帐单已支付
)

var BillStatusStringArray = []string{
	"新建", "账单已通知", "已扣费", "申述",
	"已扣费通知", "已支付", "未出账", "已出账", "已支付",
}

func (b BillStatus) String() string {
	return BillStatusStringArray[int(b)]
}

const (
	V1 = "v1"
	V2 = "v2"
	V3 = "v3"
)

// month bill setter
type MonthBillSetter struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Uid     uint32        `json:"uid" bson:"uid"`
	Month   string        `json:"month" bson:"month"`
	Day     int           `json:"day" bson:"day"`
	Desc    string        `json:"desc" bson:"desc"`
	Money   int64         `json:"money" bson:"money"`
	Details string        `json:"details" bson:"details"` // BillDetail struct encoded
	Version string        `json:"version" bson:"version"`
}

type HundredNanoSecond int64

func (a HundredNanoSecond) Time() time.Time {
	return time.Unix(int64(a)/1e7, int64(a)%1e7*100)
}

// month bill getter
type MonthBill struct {
	Id            bson.ObjectId     `json:"id" bson:"_id"`
	Uid           uint32            `json:"uid" bson:"uid"`
	Month         string            `json:"month" bson:"month"`
	Day           int               `json:"day" bson:"day"`
	Created_at    HundredNanoSecond `json:"create_at" bson:"create_at"`
	Informed_at   HundredNanoSecond `json:"inform_at" bson:"inform_at"`
	Confirmed_at  HundredNanoSecond `json:"confirm_at" bson:"confirm_at"`
	Complained_at HundredNanoSecond `json:"complain_at" bson:"complain_at"`
	Finished_at   HundredNanoSecond `json:"finish_at" bson:"finish_at"`
	Paid_at       HundredNanoSecond `json:"pay_at" bson:"pay_at"`
	Updated_at    HundredNanoSecond `json:"update_at" bson:"update_at"`
	Chargeoff_at  HundredNanoSecond `json:"chargeoff_at" bson:"chargeoff_at"`
	Status        BillStatus        `json:"status" bson:"status"`
	Desc          string            `json:"desc" bson:"desc"`
	Money         int64             `json:"money" bson:"money"`
	Details       BillDetails       `json:"details" bson:"details"`
}

// month bill formated getter
type MonthBillFormated struct {
	Id      string              `json:"id" bson:"_id"`
	Uid     uint32              `json:"uid" bson:"uid"`
	Month   string              `json:"month" bson:"month"`
	Status  BillStatus          `json:"status" bson:"status"`
	Desc    string              `json:"desc" bson:"desc"`
	Money   int64               `json:"money" bson:"money"`
	Details FormatedBillDetails `json:"details" bson:"details"`
}

type BillStatusStr string

type ItemStat struct {
	Count      int   `json:"count"`      // total count
	ValidCount int   `json:"validcount"` // count for money >= 0.01 yuan
	Money      int64 `json:"money"`      // total money, unit of money is 0.0001 yuan
}

// get the bills' stat info for all users of month
type MonthBillStat struct {
	Month      string                     `json:"month"`
	Total      ItemStat                   `json:"total"`
	StatusStat map[BillStatusStr]ItemStat `json:"statusstat"`
}

type FormatedDetailUnit struct {
	Desc  string `json:"desc"`
	Value string `json:"value"`
	Money int64  `json:"money"`
}

type FormatedBillDetail struct {
	Desc  string               `json:"desc"`
	Value string               `json:"value"`
	Money int64                `json:"money"`
	Units []FormatedDetailUnit `json:"units"`
}

type FormatedBillDiscount struct {
	Desc  string `json:"desc"`
	Money int64  `json:"money"`
}

type FormatedBillDetails struct {
	Money     int64                         `json:"money"`
	Details   map[string]FormatedBillDetail `json:"details"` // space|transfer|api_get|api_put
	Discounts []FormatedBillDiscount        `json:"discounts"`
}

// ----------------------------------------------------------------------------------------
type BillDetailUnit struct {
	Id    string `json:"id" bson:"id"`
	Type  string `json:"type" bson:"type"`
	Desc  string `json:"desc" bson:"desc"`
	From  int64  `json:"from" bson:"from"`
	To    int64  `json:"to" bson:"to"`
	Price int64  `json:"price" bson:"price"`
	Value int64  `json:"value" bson:"value"`
	Money int64  `json:"money" bson:"money"`
}

type BillDiscount struct {
	Id      string `json:"id" bson:"id"`
	Type    string `json:"type" bson:"type"`
	Desc    string `json:"desc" bson:"desc"`
	Before  int64  `json:"before" bson:"before"`
	Change  int64  `json:"change" bson:"change"`
	Percent int64  `json:"percent" bson:"percent"`
	After   int64  `json:"after" bson:"after"`
}

type BillUnitsDetail struct {
	Type  string           `json:"type" bson:"type"`
	Desc  string           `json:"desc" bson:"desc"`
	Value int64            `json:"value" bson:"value"`
	Money int64            `json:"money" bson:"money"`
	Units []BillDetailUnit `json:"units" bson:"units"`
}

type BillDetails struct {
	Money     int64                      `json:"money" bson:"money"`
	Details   map[string]BillUnitsDetail `json:"details" bson:"details"` // space|transfer|api_get|api_put
	Discounts []BillDiscount             `json:"discounts" bson:"discounts"`
}

//----------------------------------------------------------------------------//

func GetMonthBillEx(c rpc.Client, host string, uid uint32, month, id string) (
	bill MonthBill, code int, err error) {
	url := host + "/month_bill/get"
	if id == "" {
		url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
		url += "&month=" + month
	} else {
		url += "?id=" + id
	}

	code, err = c.Call(&bill, url)
	return
}

func GetMonthBillFormated(c rpc.Client, host string, uid uint32, month, id string) (
	bill MonthBillFormated, code int, err error) {
	url := host + "/month_bill/getformated"
	if id == "" {
		url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
		url += "&month=" + month
	} else {
		url += "?id=" + id
	}

	code, err = c.Call(&bill, url)
	return
}

func GetMonthBillsEx(c rpc.Client, host string, args map[string]string) (
	bills []MonthBill, code int, err error) {
	values := make(Url.Values, len(args))
	for name, value := range args {
		values.Set(name, value)
	}
	url := host + "/month_bill/list"
	if len(args) > 0 {
		url += "?" + values.Encode()
	}
	code, err = c.Call(&bills, url)
	return
}

func GetMonthBillsFormated(c rpc.Client, host string, args map[string]string) (
	bills []MonthBillFormated, code int, err error) {
	values := make(Url.Values, len(args))
	for name, value := range args {
		values.Set(name, value)
	}
	url := host + "/month_bill/formated/list"
	if len(args) > 0 {
		url += "?" + values.Encode()
	}
	code, err = c.Call(&bills, url)
	return
}

func GetRangeMonthBillsEx(c rpc.Client, host string, uid uint32, from, to string, offset, limit int) (
	bills []MonthBill, code int, err error) {
	values := Url.Values{}
	values.Add("uid", strconv.FormatUint(uint64(uid), 10))
	values.Add("from", from)
	values.Add("to", to)
	values.Add("offset", strconv.Itoa(offset))
	values.Add("limit", strconv.Itoa(limit))

	url := host + "/month_bill/rangelist"
	url += "?" + values.Encode()
	code, err = c.Call(&bills, url)
	return
}

func GetRangeMonthBillsFormated(c rpc.Client, host string, uid uint32, from, to string, offset, limit int) (
	bills []MonthBillFormated, code int, err error) {
	values := Url.Values{}
	values.Add("uid", strconv.FormatUint(uint64(uid), 10))
	values.Add("from", from)
	values.Add("to", to)
	values.Add("offset", strconv.Itoa(offset))
	values.Add("limit", strconv.Itoa(limit))

	url := host + "/month_bill/formated/rangelist"
	url += "?" + values.Encode()
	code, err = c.Call(&bills, url)
	return
}

func GetMonthBillMonthList(c rpc.Client, host string, uid uint32,
	lastMonth string, limit int) (months []int64, code int, err error) {
	url := host + "/month_bill/months"
	url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
	url += "&lastmonth=" + lastMonth
	url += "&limit=" + strconv.Itoa(limit)

	code, err = c.Call(&months, url)
	return
}

func GetMonthBillStat(c rpc.Client, host, month string) (
	stat MonthBillStat, code int, err error) {
	url := host + "/month_bill/stat?month=" + month
	code, err = c.Call(&stat, url)
	return
}

func AddMonthBillEx(c rpc.Client, host string, bill MonthBillSetter) (code int, err error) {
	code, err = c.CallWithJson(nil, host+"/month_bill/add", bill)
	return
}

func UpdateMonthBillEx(c rpc.Client, host string, bill MonthBillSetter) (code int, err error) {
	code, err = c.CallWithJson(nil, host+"/month_bill/update", bill)
	return
}

func formValues(uid uint32, month, id string) map[string][]string {
	values := Url.Values{}
	if id == "" {
		values.Add("uid", strconv.FormatUint(uint64(uid), 10))
		values.Add("month", month)
	} else {
		values.Add("id", id)
	}
	return values
}

func DoneMonthBillEx(c rpc.Client, host string, uid uint32, month, id string) (
	code int, err error) {
	values := formValues(uid, month, id)
	code, err = c.CallWithForm(nil, host+"/month_bill/done", values)

	return
}

func InformMonthBillEx(c rpc.Client, host string, uid uint32, month, id string) (
	code int, err error) {
	values := formValues(uid, month, id)
	code, err = c.CallWithForm(nil, host+"/month_bill/inform", values)

	return
}

func ConfirmMonthBillEx(c rpc.Client, host string, uid uint32, month, id string) (
	code int, err error) {
	values := formValues(uid, month, id)
	code, err = c.CallWithForm(nil, host+"/month_bill/confirm", values)

	return
}

func ChargeoffMonthBillEx(c rpc.Client, host string, uid uint32, month, id string) (
	code int, err error) {
	values := formValues(uid, month, id)
	code, err = c.CallWithForm(nil, host+"/month_bill/chargeoff", values)

	return
}

func ComplainMonthBillEx(c rpc.Client, host string, uid uint32, month, id string) (
	code int, err error) {
	values := formValues(uid, month, id)
	code, err = c.CallWithForm(nil, host+"/month_bill/complain", values)

	return
}

func DecomplainMonthBillEx(c rpc.Client, host string, uid uint32, month, id string) (
	code int, err error) {
	values := formValues(uid, month, id)
	code, err = c.CallWithForm(nil, host+"/month_bill/decomplain", values)
	return
}

func CheckMonthBillPaidEx(c rpc.Client, host string, uid uint32, month, id string) (
	code int, err error) {
	values := formValues(uid, month, id)
	code, err = c.CallWithForm(nil, host+"/month_bill/checkpaid", values)

	return
}

func DelMonthBillEx(c rpc.Client, host string, uid uint32, month, id string) (
	code int, err error) {
	values := formValues(uid, month, id)
	code, err = c.CallWithForm(nil, host+"/month_bill/del", values)

	return
}

func SetBillStatusNotOut(c rpc.Client, host string, uid uint32, month string) (code int, err error) {
	values := formValues(uid, month, "")
	code, err = c.CallWithForm(nil, host+"/month_bill/set_status_not_out", values)
	return
}

//----------------------------------------------------------------------------//

func (r *ServiceInEx) GetMonthBill(user account.UserInfo, uid uint32, month, id string) (
	bill MonthBill, code int, err error) {
	return GetMonthBillEx(r.getClient(user), r.host, uid, month, id)
}

func (r *ServiceInEx) GetMonthBillFormated(user account.UserInfo, uid uint32, month, id string) (
	bill MonthBillFormated, code int, err error) {
	return GetMonthBillFormated(r.getClient(user), r.host, uid, month, id)
}

func (r *ServiceInEx) GetMonthBillStat(user account.UserInfo, month string) (
	stat MonthBillStat, code int, err error) {
	return GetMonthBillStat(r.getClient(user), r.host, month)
}

func (r *ServiceInEx) GetMonthBills(user account.UserInfo, args map[string]string) (
	bills []MonthBill, code int, err error) {
	return GetMonthBillsEx(r.getClient(user), r.host, args)
}

func (r *ServiceInEx) GetMonthBillsFormated(user account.UserInfo, args map[string]string) (
	bills []MonthBillFormated, code int, err error) {
	return GetMonthBillsFormated(r.getClient(user), r.host, args)
}

func (r *ServiceInEx) GetRangeMonthBills(user account.UserInfo, uid uint32, from, to string,
	offset, limit int) (bills []MonthBill, code int, err error) {
	return GetRangeMonthBillsEx(r.getClient(user), r.host, uid, from, to, offset, limit)
}

func (r *ServiceInEx) GetRangeMonthBillsFormated(user account.UserInfo, uid uint32, from, to string,
	offset, limit int) (bills []MonthBillFormated, code int, err error) {
	return GetRangeMonthBillsFormated(r.getClient(user), r.host, uid, from, to, offset, limit)
}

func (r *ServiceInEx) GetMonthBillMonthList(user account.UserInfo, uid uint32,
	lastMonth string, limit int) (months []int64, code int, err error) {
	return GetMonthBillMonthList(r.getClient(user), r.host, uid, lastMonth, limit)
}

func (r *ServiceInEx) AddMonthBill(user account.UserInfo, bill MonthBillSetter) (code int, err error) {
	return AddMonthBillEx(r.getClient(user), r.host, bill)
}

func (r *ServiceInEx) UpdateMonthBill(user account.UserInfo, bill MonthBillSetter) (code int, err error) {
	return UpdateMonthBillEx(r.getClient(user), r.host, bill)
}

func (r *ServiceInEx) DoneMonthBill(user account.UserInfo, uid uint32, month, id string) (
	code int, err error) {
	return DoneMonthBillEx(r.getClient(user), r.host, uid, month, id)
}

func (r *ServiceInEx) InformMonthBill(user account.UserInfo, uid uint32, month, id string) (
	code int, err error) {
	return InformMonthBillEx(r.getClient(user), r.host, uid, month, id)
}

func (r *ServiceInEx) ComplainMonthBill(user account.UserInfo, uid uint32, month, id string) (
	code int, err error) {
	return ComplainMonthBillEx(r.getClient(user), r.host, uid, month, id)
}

func (r *ServiceInEx) DecomplainMonthBill(user account.UserInfo, uid uint32, month, id string) (
	code int, err error) {
	return DecomplainMonthBillEx(r.getClient(user), r.host, uid, month, id)
}

func (r *ServiceInEx) ConfirmMonthBill(user account.UserInfo, uid uint32, month, id string) (
	code int, err error) {
	return ConfirmMonthBillEx(r.getClient(user), r.host, uid, month, id)
}

func (r *ServiceInEx) ChargeoffMonthBill(user account.UserInfo, uid uint32, month, id string) (
	code int, err error) {
	return ChargeoffMonthBillEx(r.getClient(user), r.host, uid, month, id)
}

func (r *ServiceInEx) CheckMonthBillPaid(user account.UserInfo, uid uint32, month, id string) (
	code int, err error) {
	return CheckMonthBillPaidEx(r.getClient(user), r.host, uid, month, id)
}

func (r *ServiceInEx) DelMonthBill(user account.UserInfo, uid uint32, month, id string) (
	code int, err error) {
	return DelMonthBillEx(r.getClient(user), r.host, uid, month, id)
}

//----------------------------------------------------------------------------//

func (r ServiceEx) GetMonthBills(host string, args map[string]string) (
	bills []MonthBill, code int, err error) {
	return GetMonthBillsEx(r.Conn, host, args)
}

func (r ServiceEx) GetMonthBillsFormated(host string, args map[string]string) (
	bills []MonthBillFormated, code int, err error) {
	return GetMonthBillsFormated(r.Conn, host, args)
}

func (r ServiceEx) GetRangeMonthBills(host string, uid uint32, from, to string, offset, limit int) (
	bills []MonthBill, code int, err error) {
	return GetRangeMonthBillsEx(r.Conn, host, uid, from, to, offset, limit)
}

func (r ServiceEx) GetRangeMonthBillsFormated(host string, uid uint32, from, to string, offset, limit int) (
	bills []MonthBillFormated, code int, err error) {
	return GetRangeMonthBillsFormated(r.Conn, host, uid, from, to, offset, limit)
}

func (r ServiceEx) GetMonthBillMonthList(host string, uid uint32,
	lastMonth string, limit int) (months []int64, code int, err error) {
	return GetMonthBillMonthList(r.Conn, host, uid, lastMonth, limit)
}

func (r ServiceEx) GetMonthBill(host string, uid uint32, month, id string) (
	bill MonthBill, code int, err error) {
	return GetMonthBillEx(r.Conn, host, uid, month, id)
}

func (r ServiceEx) GetMonthBillFormated(host string, uid uint32, month, id string) (
	bill MonthBillFormated, code int, err error) {
	return GetMonthBillFormated(r.Conn, host, uid, month, id)
}

func (r ServiceEx) GetMonthBillStat(host, month string) (
	stat MonthBillStat, code int, err error) {
	return GetMonthBillStat(r.Conn, host, month)
}

func (r ServiceEx) AddMonthBill(host string, bill MonthBillSetter) (code int, err error) {
	return AddMonthBillEx(r.Conn, host, bill)
}

func (r ServiceEx) UpdateMonthBill(host string, bill MonthBillSetter) (code int, err error) {
	return UpdateMonthBillEx(r.Conn, host, bill)
}

func (r ServiceEx) DoneMonthBill(host string, uid uint32, month, id string) (code int, err error) {
	return DoneMonthBillEx(r.Conn, host, uid, month, id)
}

func (r ServiceEx) InformMonthBill(host string, uid uint32, month, id string) (code int, err error) {
	return InformMonthBillEx(r.Conn, host, uid, month, id)
}

func (r ServiceEx) ComplainMonthBill(host string, uid uint32, month, id string) (code int, err error) {
	return ComplainMonthBillEx(r.Conn, host, uid, month, id)
}

func (r ServiceEx) DecomplainMonthBill(host string, uid uint32, month, id string) (code int, err error) {
	return DecomplainMonthBillEx(r.Conn, host, uid, month, id)
}

func (r ServiceEx) ConfirmMonthBill(host string, uid uint32, month, id string) (code int, err error) {
	return ConfirmMonthBillEx(r.Conn, host, uid, month, id)
}

func (r ServiceEx) CheckMonthBillPaid(host string, uid uint32, month, id string) (
	code int, err error) {
	return CheckMonthBillPaidEx(r.Conn, host, uid, month, id)
}

func (r ServiceEx) DelMonthBill(host string, uid uint32, month, id string) (
	code int, err error) {
	return DelMonthBillEx(r.Conn, host, uid, month, id)
}

func (r *ServiceEx) ChargeoffMonthBill(host string, uid uint32, month, id string) (
	code int, err error) {
	return ChargeoffMonthBillEx(r.Conn, host, uid, month, id)
}

func (r *ServiceEx) SetBillStatusNotOut(host string, uid uint32, month string) (code int, err error) {
	return SetBillStatusNotOut(r.Conn, host, uid, month)
}
