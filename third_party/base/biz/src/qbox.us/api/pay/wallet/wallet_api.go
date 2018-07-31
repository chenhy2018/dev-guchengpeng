package wallet

import (
	"net/http"
	"strconv"

	Url "net/url"
	"qbox.us/rpc"
	"qbox.us/servend/account"
	"qbox.us/servend/oauth"
)

//---------------------------------------------------------------------------//

func Recharge(c rpc.Client, host, excode, type_ string, uid uint32, money int64,
	desc string) (code int, err error) {

	code, err = c.CallWithForm(nil, host+"/recharge",
		map[string][]string{
			"excode": []string{excode},
			"type":   []string{type_},
			"uid":    []string{strconv.FormatUint(uint64(uid), 10)},
			"money":  []string{strconv.FormatInt(money, 10)},
			"desc":   []string{desc},
		})
	return
}

type Info struct {
	Amount       int64 `json:"amount"`
	Cash         int64 `json:"cash"`
	Coupon       int64 `json:"coupon"`
	VirtualMoney int64 `json:"virtual_money"`
}

func GetInfo(c rpc.Client, host string, uid uint32) (info Info, code int, err error) {
	code, err = c.Call(&info, host+"/info?uid="+strconv.FormatUint(uint64(uid), 10))
	return
}

func GetBills(c rpc.Client, host string, uid uint32, starttime, endtime int64,
	prefix, type_, expenses string) (bills []Bill, code int, err error) {
	url := host + "/get_bills"
	url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
	url += "&starttime=" + strconv.FormatInt(starttime, 10)
	url += "&endtime=" + strconv.FormatInt(endtime, 10)
	url += "&prefix=" + prefix + "&type=" + type_ + "&expenses=" + expenses
	code, err = c.Call(&bills, url)
	return
}

//type Change struct {
//	Time int64 `jons:"time"`
//	Change int64 `json:"change"`
//	Income int64 `json:"change"`
//	Expenses int64 `json:"expenses"`
//}
//
//func GetMonthChanges(c rpc.Client, host string, uid uint32, starttime, endtime int64) (
//	changes []Change, code int, err error) {
//	url := host + "/change/month"
//	url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
//	url += "&starttime=" + strconv.FormatInt(starttime, 10)
//	url += "&endtime=" + strconv.FormatInt(endtime, 10)
//	code, err = c.Call(&changes, url)
//	return
//}

//---------------------------------------------------------------------------//

// func Deposit(c rpc.Client, host, serial_num, id, type_ string, uid uint32,
// 	money, deadtime int64, desc string) (code int, err error) {

// 	code, err = c.CallWithForm(nil, host+"/deposit",
// 		map[string][]string{
// 			"serial_num": []string{serial_num},
// 			"id":         []string{id},
// 			"type":       []string{type_},
// 			"uid":        []string{strconv.FormatUint(uint64(uid), 10)},
// 			"money":      []string{strconv.FormatInt(money, 10)},
// 			"deadtime":   []string{strconv.FormatInt(deadtime, 10)},
// 			"desc":       []string{desc},
// 		})
// 	return
// }

// type MoneyInfo struct {
// 	Money int64
// }

// func GetDeposit(c rpc.Client, host string, uid uint32) (info MoneyInfo, code int, err error) {
// 	code, err = c.Call(&info,
// 		host+"/get_deposit?uid="+strconv.FormatUint(uint64(uid), 10))
// 	return
// }

// type DepositInfo struct {
// 	Money    int64  `json:"money"`
// 	Deadtime int64  `json:"deadtime"`
// 	Id       string `json:"id"`
// 	Type     string `json:"type"`
// 	Uid      uint32 `json:"uid"`
// }

// func TryCleanDeposit(c rpc.Client, host string, t int64) (
// 	deposits []DepositInfo, code int, err error) {
// 	code, err = c.Call(&deposits,
// 		host+"/try_clean_deposit?time="+strconv.FormatInt(t, 10))
// 	return
// }

// func DiscardDeposit(c rpc.Client, host string, serial_num string, uid uint32,
// 	id, type_, desc string) (code int, err error) {
// 	code, err = c.CallWithForm(nil, host+"/discard_deposit",
// 		map[string][]string{
// 			"serial_num": []string{serial_num},
// 			"uid":        []string{strconv.FormatUint(uint64(uid), 10)},
// 			"id":         []string{id},
// 			"type":       []string{type_},
// 			"desc":       []string{desc},
// 		})
// 	return
// }

//---------------------------------------------------------------------------//

func NewCoupon(c rpc.Client, host, excode, type_ string, quota int64,
	day int, deadtime int64, desc string) (id string, code int, err error) {

	var m map[string]string = make(map[string]string)
	code, err = c.CallWithForm(&m, host+"/new_coupon",
		map[string][]string{
			"excode":   []string{excode},
			"type":     []string{type_},
			"quota":    []string{strconv.FormatInt(quota, 10)},
			"day":      []string{strconv.FormatInt(int64(day), 10)},
			"deadtime": []string{strconv.FormatInt(deadtime, 10)},
			"desc":     []string{desc},
		})
	if err == nil {
		id = m["id"]
	}
	return
}

func NewCouponNew(c rpc.Client, host, title, type_ string, quota int64,
	day int, deadtime int64, desc string) (id string, code int, err error) {

	code, err = c.CallWithForm(&id, host+"/coupon/new",
		map[string][]string{
			"title":    []string{title},
			"type":     []string{type_},
			"quota":    []string{strconv.FormatInt(quota, 10)},
			"day":      []string{strconv.FormatInt(int64(day), 10)},
			"deadtime": []string{strconv.FormatInt(deadtime, 10)},
			"desc":     []string{desc},
		})
	return
}

func ActiveCoupon(c rpc.Client, host, excode string, uid uint32, id string,
	desc string) (coupon Coupon, code int, err error) {

	code, err = c.CallWithForm(&coupon, host+"/active_coupon",
		map[string][]string{
			"excode": []string{excode},
			"uid":    []string{strconv.FormatUint(uint64(uid), 10)},
			"id":     []string{id},
			"desc":   []string{desc},
		})
	return
}

func TryCleanCoupon(c rpc.Client, host string, t int64) (coupons []Coupon, code int, err error) {
	code, err = c.Call(&coupons, host+"/try_clean_coupon?time="+strconv.FormatInt(t, 10))
	return
}

func DiscardCoupon(c rpc.Client, host, excode, id string, uid uint32, desc string) (
	code int, err error) {
	code, err = c.CallWithForm(nil, host+"/discard_coupon",
		map[string][]string{
			"excode": []string{excode},
			"id":     []string{id},
			"uid":    []string{strconv.FormatUint(uint64(uid), 10)},
			"desc":   []string{desc},
		})
	return
}

type Coupon struct {
	Quota      int64        `json:"quota"`
	Balance    int64        `json:"balance"`
	Effecttime int64        `json:"effecttime"`
	Deadtime   int64        `json:"deadtime"`
	Uid        uint32       `json:"uid"`
	Day        int          `json:"day"`
	Id         string       `json:"id"`
	CreateAt   int64        `json:"create_at"`
	UpdateAt   int64        `json:"update_at"`
	Type       CouponType   `json:"type"`
	Status     CouponStatus `json:"status"`
	Title      string       `json:"title"`
	Desc       string       `json:"desc"`
}

type CouponStatus int

const (
	COUPON_STATUS_IGNORE = iota
	COUPON_STATUS_NEW
	COUPON_STATUS_ACTIVE
)

type CouponType string

var COUPON_TYPE_NEWUSER CouponType = "NEWUSER"
var COUPON_TYPE_RECHARGE CouponType = "RECHARGE"
var COUPON_TYPE_INVALID CouponType = ""

func GetCoupons(c rpc.Client, host string, uid uint32) (
	coupons []Coupon, code int, err error) {
	url := host + "/get_coupons"
	url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
	code, err = c.Call(&coupons, url)
	return
}

func GetCoupon(c rpc.Client, host string, id string) (
	coupon Coupon, code int, err error) {
	url := host + "/coupon/get"
	url += "?id=" + id
	code, err = c.Call(&coupon, url)
	return
}

func GetInactiveCoupons(c rpc.Client, host string) (coupons []Coupon, code int, err error) {
	code, err = c.Call(coupons, host+"/get_inactive_coupons")
	return
}

func GetAdminCouponList(c rpc.Client, host string, uid, title string, type_ CouponType, status CouponStatus,
	offset, limit string) (coupons []Coupon, code int, err error) {
	url := host + "/coupon/admin/list"

	values := genCouponListArgs(uid, title, type_, status, offset, limit)

	if len(values) > 0 {
		url += "?" + values.Encode()
	}

	code, err = c.Call(&coupons, url)
	return
}

func GetAdminCouponCount(c rpc.Client, host, uid, title string, type_ CouponType, status CouponStatus) (
	count, code int, err error) {
	url := host + "/coupon/admin/count"

	values := genCouponCountArgs(uid, title, type_, status)

	if len(values) > 0 {
		url += "?" + values.Encode()
	}

	code, err = c.Call(&count, url)
	return
}

func genCouponCountArgs(uid, title string, type_ CouponType, status CouponStatus) Url.Values {
	return genCouponListArgs(uid, title, type_, status, "", "")
}

func genCouponListArgs(uid, title string, type_ CouponType, status CouponStatus, offset, limit string) Url.Values {
	values := Url.Values{}
	add := func(key, value, invalid string) {
		if value != invalid {
			values.Add(key, value)
		}
	}
	add("uid", uid, "")
	add("title", title, "")
	add("type", string(type_), string(COUPON_TYPE_INVALID))
	add("status", strconv.Itoa(int(status)), strconv.Itoa(int(COUPON_STATUS_IGNORE)))
	add("offset", offset, "")
	add("limit", limit, "")
	return values
}

///////////////////////////////////////////////////////////////////////////////

type ServiceIn struct {
	host string
	acc  account.InterfaceEx
}

func NewServiceIn(host string, acc account.InterfaceEx) *ServiceIn {
	return &ServiceIn{host: host, acc: acc}
}

func (r *ServiceIn) getClient(user account.UserInfo) rpc.Client {
	token := r.acc.MakeAccessToken(user)
	return rpc.Client{oauth.NewClient(token, nil)}
}

func (r *ServiceIn) Recharge(user account.UserInfo, excode, type_ string, uid uint32,
	money int64, desc string) (code int, err error) {
	return Recharge(r.getClient(user), r.host, excode, type_, uid, money, desc)
}

func (r *ServiceIn) GetInfo(user account.UserInfo, uid uint32) (info Info, code int, err error) {
	return GetInfo(r.getClient(user), r.host, uid)
}

func (r *ServiceIn) GetBills(user account.UserInfo, uid uint32, starttime, endtime int64,
	prefix, type_, expenses string) (bills []Bill, code int, err error) {
	return GetBills(r.getClient(user), r.host, uid, starttime, endtime, prefix, type_, expenses)
}

//func (r *ServiceIn) GetMonthChanges(user account.UserInfo, uid uint32, starttime, endtime int64) (
//	changes []Change, code int, err error) {
//	return GetMonthChanges(r.getClient(user), r.host, uid, starttime, endtime)
//}

// func (r *ServiceIn) Deposit(user account.UserInfo, serial_num, id, type_ string, uid uint32,
// 	money, deadtime int64, desc string) (code int, err error) {
// 	return Deposit(r.getClient(user), r.host, serial_num, id, type_, uid, money, deadtime, desc)
// }

// func (r *ServiceIn) GetDeposit(user account.UserInfo, uid uint32) (
// 	info MoneyInfo, code int, err error) {
// 	return GetDeposit(r.getClient(user), r.host, uid)
// }

// func (r *ServiceIn) TryCleanDeposit(user account.UserInfo, t int64) (
// 	deposits []DepositInfo, code int, err error) {
// 	return TryCleanDeposit(r.getClient(user), r.host, t)
// }

// func (r *ServiceIn) DiscardDeposit(user account.UserInfo, excode string, uid uint32,
// 	id, type_, desc string) (code int, err error) {
// 	return DiscardDeposit(r.getClient(user), r.host, excode, uid, id, type_, desc)
// }

func (r *ServiceIn) NewCoupon(user account.UserInfo, excode, type_ string, quota int64,
	day int, deadtime int64, desc string) (id string, code int, err error) {
	return NewCoupon(r.getClient(user), r.host, excode, type_, quota, day, deadtime, desc)
}

func (r *ServiceIn) ActiveCoupon(user account.UserInfo, excode string, uid uint32, id string,
	desc string) (coupon Coupon, code int, err error) {
	return ActiveCoupon(r.getClient(user), r.host, excode, uid, id, desc)
}

func (r *ServiceIn) TryCleanCoupon(user account.UserInfo, t int64) (
	coupons []Coupon, code int, err error) {
	return TryCleanCoupon(r.getClient(user), r.host, t)
}

func (r *ServiceIn) DiscardCoupon(user account.UserInfo, excode, id string, uid uint32,
	desc string) (code int, err error) {
	return DiscardCoupon(r.getClient(user), r.host, excode, id, uid, desc)
}

func (r *ServiceIn) GetCoupons(user account.UserInfo, uid uint32) (
	coupons []Coupon, code int, err error) {
	return GetCoupons(r.getClient(user), r.host, uid)
}

func (r *ServiceIn) GetInactiveCoupons(user account.UserInfo) (
	coupons []Coupon, code int, err error) {
	return GetInactiveCoupons(r.getClient(user), r.host)
}

//---------------------------------------------------------------------------//

type Service struct {
	Conn rpc.Client
}

func New(t http.RoundTripper) Service {
	client := &http.Client{Transport: t}
	return Service{rpc.Client{client}}
}

func (r Service) Recharge(host string, excode, type_ string, uid uint32, money int64,
	desc string) (code int, err error) {
	return Recharge(r.Conn, host, excode, type_, uid, money, desc)
}

func (r Service) GetInfo(host string, uid uint32) (info Info, code int, err error) {
	return GetInfo(r.Conn, host, uid)
}

func (r Service) GetBills(host string, uid uint32, starttime, endtime int64,
	prefix, type_, expenses string) (bills []Bill, code int, err error) {
	return GetBills(r.Conn, host, uid, starttime, endtime, prefix, type_, expenses)
}

//func (r Service) GetMonthChanges(host string, uid uint32, starttime, endtime int64) (
//	changes []Change, code int, err error) {
//	return GetMonthChanges(r.Conn, host, uid, starttime, endtime)
//}

// func (r Service) Deposit(host string, serial_num, id, type_ string, uid uint32,
// 	money, deadtime int64, desc string) (code int, err error) {
// 	return Deposit(r.Conn, host, serial_num, id, type_, uid, money, deadtime, desc)
// }

// func (r Service) GetDeposit(host string, uid uint32) (info MoneyInfo, code int, err error) {
// 	return GetDeposit(r.Conn, host, uid)
// }

// func (r Service) TryCleanDeposit(host string, t int64) (
// 	deposits []DepositInfo, code int, err error) {
// 	return TryCleanDeposit(r.Conn, host, t)
// }

// func (r Service) DiscardDeposit(host string, serial_num string, uid uint32,
// 	id, type_, desc string) (code int, err error) {
// 	return DiscardDeposit(r.Conn, host, serial_num, uid, id, type_, desc)
// }

func (r Service) NewCoupon(host string, excode, type_ string, quota int64,
	day int, deadtime int64, desc string) (id string, code int, err error) {
	return NewCoupon(r.Conn, host, excode, type_, quota, day, deadtime, desc)
}

func (r Service) ActiveCoupon(host string, excode string, uid uint32, id string,
	desc string) (coupon Coupon, code int, err error) {
	return ActiveCoupon(r.Conn, host, excode, uid, id, desc)
}

func (r Service) TryCleanCoupon(host string, t int64) (coupons []Coupon, code int, err error) {
	return TryCleanCoupon(r.Conn, host, t)
}

func (r Service) DiscardCoupon(host string, excode, id string, uid uint32,
	desc string) (code int, err error) {
	return DiscardCoupon(r.Conn, host, excode, id, uid, desc)
}

func (r Service) GetCoupons(host string, uid uint32) (
	coupons []Coupon, code int, err error) {
	return GetCoupons(r.Conn, host, uid)
}

func (r Service) GetInactiveCoupons(host string) (
	coupons []Coupon, code int, err error) {
	return GetInactiveCoupons(r.Conn, host)
}
