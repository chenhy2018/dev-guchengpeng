package wallet

import (
	Url "net/url"
	"qbox.us/rpc"
	"qbox.us/servend/account"
	"strconv"
)

type DayBill struct {
	Id      string `json:"id" bson:"_id"`
	Uid     uint32 `json:"uid" bson:"uid"`
	Day     string `json:"day" bson:"day"` // "2006-01-02"
	Desc    string `json:"desc" bson:"desc"`
	Money   int64  `json:"money" bson:"money"`
	Details string `json:"details" bson:"details"` // BillDetails struct encoded
}

//----------------------------------------------------------------------------//

func GetDayBill(c rpc.Client, host string, uid uint32, day string) (
	bill DayBill, code int, err error) {
	url := host + "/day_bill/get"
	url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
	url += "&day=" + day // "20060102"

	code, err = c.Call(&bill, url)
	return
}

func AddDayBill(c rpc.Client, host string, bill DayBill) (code int, err error) {
	code, err = c.CallWithJson(nil, host+"/day_bill/add", bill)
	return
}

func DelDayBill(c rpc.Client, host string, uid uint32, day string) (code int, err error) {
	values := Url.Values{}
	values.Add("uid", strconv.FormatUint(uint64(uid), 10))
	values.Add("day", day)

	code, err = c.CallWithForm(nil, host+"/day_bill/del", values)
	return
}

//----------------------------------------------------------------------------//

func (r *ServiceInEx) GetDayBill(user account.UserInfo, uid uint32, day string) (
	bill DayBill, code int, err error) {
	return GetDayBill(r.getClient(user), r.host, uid, day)
}

func (r *ServiceInEx) AddDayBill(user account.UserInfo, bill DayBill) (code int, err error) {
	return AddDayBill(r.getClient(user), r.host, bill)
}

func (r *ServiceInEx) DelDayBill(user account.UserInfo, uid uint32, day string) (
	code int, err error) {
	return DelDayBill(r.getClient(user), r.host, uid, day)
}

//----------------------------------------------------------------------------//

func (r ServiceEx) GetDayBill(host string, uid uint32, day string) (
	bill DayBill, code int, err error) {
	return GetDayBill(r.Conn, host, uid, day)
}

func (r ServiceEx) AddDayBill(host string, bill DayBill) (code int, err error) {
	return AddDayBill(r.Conn, host, bill)
}

func (r ServiceEx) DelDayBill(host string, uid uint32, day string) (
	code int, err error) {
	return DelDayBill(r.Conn, host, uid, day)
}
