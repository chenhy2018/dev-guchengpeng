package wallet

import (
	"net/url"
	"strconv"

	"qbox.us/rpc"
	"qbox.us/servend/account"
)

type MonthInSecond int64
type Second int64

type MonthInfo struct {
	Uid       uint32        `json:"uid"`
	Month     MonthInSecond `json:"month"`
	Biz       MonthBizInfo  `json:"biz"`
	Money     ChangeInfo    `json:"money"`
	Coupon    ChangeInfo    `json:"coupon"`
	StartTime Second        `json:"starttime"`
}

type MonthBizStatus int

const (
	MONTH_BIZ_STATUS_NONE   MonthBizStatus = 0
	MONTH_BIZ_STATUS_NORMAL MonthBizStatus = 1
	MONTH_BIZ_STATUS_DONE   MonthBizStatus = 2
	MONTH_BIZ_STATUS_PAID   MonthBizStatus = 3
)

type MonthBizInfo struct {
	Status MonthBizStatus `json:"status"`
	Deduct BizDeductInfo  `json:"deduct"`
}

type BizDeductInfo struct {
	All    int64 `json:"all"`    // Money + Coupon
	Money  int64 `json:"money"`  // 虚拟币
	Coupon int64 `json:"coupon"` // 优惠劵
}

type ChangeInfo struct {
	Before   int64 `json:"before"`
	Deduct   int64 `json:"deduct"`
	Recharge int64 `json:"recharge"`
	After    int64 `json:"after"`
}

//----------------------------------------------------------------------------//

func GetMonthBizInfo(c rpc.Client, host string, uid uint32, month string) (
	info MonthBizInfo, code int, err error) {
	args := url.Values{}
	args.Add("uid", strconv.FormatUint(uint64(uid), 10))
	args.Add("month", month) // 200601
	code, err = c.Call(&info, host+"/month/biz/get?"+args.Encode())
	return
}

func GetMonthInfo(c rpc.Client, host string, uid uint32, month string) (
	info MonthInfo, code int, err error) {
	args := url.Values{}
	args.Add("uid", strconv.FormatUint(uint64(uid), 10))
	args.Add("month", month) // 200601
	code, err = c.Call(&info, host+"/month/get?"+args.Encode())
	return
}

//	Args:
//	  uid 可指定多个用户id或不指定
//	  month 指定月份，格式200601，可选
//	  biz.status 0（不出）|1（可出）|2（已出）|3（已付），可选
//	  biz.rebate true|false，可选
func ListMonthInfo(c rpc.Client, host string, args map[string][]string) (
	infos []MonthInfo, code int, err error) {
	values := url.Values{}
	for key, vs := range args {
		for _, v := range vs {
			values.Add(key, v)
		}
	}
	code, err = c.Call(&infos, host+"/month/list?"+values.Encode())
	return
}

func setMonthBizStatus(c rpc.Client, host string, uid uint32, month string, status MonthBizStatus) (
	code int, err error) {
	args := url.Values{}
	args.Add("uid", strconv.FormatUint(uint64(uid), 10))
	args.Add("month", month) // 200601
	args.Add("biz.status", strconv.Itoa(int(status)))
	code, err = c.CallWithForm(nil, host+"/month/biz/status", args)
	return
}

func GetDayDeductUids(c rpc.Client, host string, day string, offset, limit int) (
	uids []uint32, code int, err error) {
	args := url.Values{}
	args.Add("day", day) // "20060102"
	args.Add("offset", strconv.Itoa(offset))
	args.Add("limit", strconv.Itoa(limit))
	code, err = c.Call(&uids, host+"/month/get_daydeduct_uidlist?"+args.Encode())
	return
}

func GetMonthDeductUids(c rpc.Client, host string, month string, offset, limit int) (
	uids []uint32, code int, err error) {
	args := url.Values{}
	args.Add("month", month) // "200601"
	args.Add("offset", strconv.Itoa(offset))
	args.Add("limit", strconv.Itoa(limit))

	code, err = c.Call(&uids, host+"/month/get_monthdeduct_uidlist?"+args.Encode())
	return
}

func UpdateAllMonthBizs(c rpc.Client, host string) (code int, err error) {
	code, err = c.CallWithForm(nil, host+"/month/biz/update", url.Values{})
	return
}

func UpdateAllMonthInfos(c rpc.Client, host, month string) (code int, err error) {
	args := url.Values{}
	args.Add("month", month) // 200601
	code, err = c.CallWithForm(nil, host+"/month/update", args)
	return
}

//----------------------------------------------------------------------------//

func (r *ServiceInEx) GetMonthBizInfo(user account.UserInfo, uid uint32, month string) (
	info MonthBizInfo, code int, err error) {
	return GetMonthBizInfo(r.getClient(user), r.host, uid, month)
}

func (r *ServiceInEx) GetMonthInfo(user account.UserInfo, uid uint32, month string) (
	info MonthInfo, code int, err error) {
	return GetMonthInfo(r.getClient(user), r.host, uid, month)
}

func (r *ServiceInEx) ListMonthInfo(user account.UserInfo, args map[string][]string) (
	infos []MonthInfo, code int, err error) {
	return ListMonthInfo(r.getClient(user), r.host, args)
}

func (r *ServiceInEx) SetMonthBizStatusNone(user account.UserInfo, uid uint32, month string) (
	code int, err error) {
	return setMonthBizStatus(r.getClient(user), r.host, uid, month, MONTH_BIZ_STATUS_NONE)
}

func (r *ServiceInEx) SetMonthBizStatusNormal(user account.UserInfo, uid uint32, month string) (
	code int, err error) {
	return setMonthBizStatus(r.getClient(user), r.host, uid, month, MONTH_BIZ_STATUS_NORMAL)
}

func (r *ServiceInEx) GetDayDeductUids(user account.UserInfo, day string, offset, limit int) (
	uids []uint32, code int, err error) {
	return GetDayDeductUids(r.getClient(user), r.host, day, offset, limit)
}

func (r *ServiceInEx) GetMonthDeductUids(user account.UserInfo, month string, offset, limit int) (
	uids []uint32, code int, err error) {
	return GetMonthDeductUids(r.getClient(user), r.host, month, offset, limit)
}

func (r *ServiceInEx) UpdateAllMonthBizs(user account.UserInfo) (code int, err error) {
	return UpdateAllMonthBizs(r.getClient(user), r.host)
}

func (r *ServiceInEx) UpdateAllMonthInfos(user account.UserInfo, month string) (
	code int, err error) {
	return UpdateAllMonthInfos(r.getClient(user), r.host, month)
}

//----------------------------------------------------------------------------//

func (r ServiceEx) GetMonthBizInfo(host string, uid uint32, month string) (
	info MonthBizInfo, code int, err error) {
	return GetMonthBizInfo(r.Conn, host, uid, month)
}

func (r ServiceEx) GetMonthInfo(host string, uid uint32, month string) (
	info MonthInfo, code int, err error) {
	return GetMonthInfo(r.Conn, host, uid, month)
}

func (r ServiceEx) ListMonthInfo(host string, args map[string][]string) (
	infos []MonthInfo, code int, err error) {
	return ListMonthInfo(r.Conn, host, args)
}

func (r ServiceEx) SetMonthBizStatusNone(host string, uid uint32, month string) (
	code int, err error) {
	return setMonthBizStatus(r.Conn, host, uid, month, MONTH_BIZ_STATUS_NONE)
}

func (r ServiceEx) SetMonthBizStatusNormal(host string, uid uint32, month string) (
	code int, err error) {
	return setMonthBizStatus(r.Conn, host, uid, month, MONTH_BIZ_STATUS_NORMAL)
}

func (r ServiceEx) GetDayDeductUids(host, day string, offset, limit int) (
	uids []uint32, code int, err error) {
	return GetDayDeductUids(r.Conn, host, day, offset, limit)
}

func (r ServiceEx) GetMonthDeductUids(host, month string, offset, limit int) (
	uids []uint32, code int, err error) {
	return GetMonthDeductUids(r.Conn, host, month, offset, limit)
}

func (r ServiceEx) UpdateAllMonthBizs(host string) (code int, err error) {
	return UpdateAllMonthBizs(r.Conn, host)
}

func (r ServiceEx) UpdateAllMonthInfos(host, month string) (code int, err error) {
	return UpdateAllMonthInfos(r.Conn, host, month)
}
