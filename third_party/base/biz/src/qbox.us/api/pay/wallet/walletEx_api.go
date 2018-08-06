package wallet

import (
	"net/http"
	Url "net/url"
	"strconv"

	"qbox.us/rpc"
	"qbox.us/servend/account"
	"qbox.us/servend/oauth"
)

//---------------------------------------------------------------------------//

func RechargeEx(c rpc.Client, host, excode string, type_ string, uid uint32, money int64,
	desc string) (code int, err error) {
	return Recharge(c, host, excode, type_, uid, money, desc)
}

func RechargeReward(c rpc.Client, host, excode string, type_ string, uid uint32, money int64,
	desc string, cost int64, customerRewardId string) (code int, err error) {
	code, err = c.CallWithForm(nil, host+"/recharge_reward",
		map[string][]string{
			"excode":             []string{excode},
			"type":               []string{type_},
			"uid":                []string{strconv.FormatUint(uint64(uid), 10)},
			"money":              []string{strconv.FormatInt(money, 10)},
			"desc":               []string{desc},
			"cost":               []string{strconv.FormatInt(cost, 10)},
			"customer_reward_id": []string{customerRewardId},
		})
	return
}

func RechargeRewardWithResp(c rpc.Client, host, excode string, type_ string, uid uint32, money int64,
	desc string, cost int64, customerRewardId string) (id string, code int, err error) {
	code, err = c.CallWithForm(&id, host+"/recharge/reward",
		map[string][]string{
			"excode":             []string{excode},
			"type":               []string{type_},
			"uid":                []string{strconv.FormatUint(uint64(uid), 10)},
			"money":              []string{strconv.FormatInt(money, 10)},
			"desc":               []string{desc},
			"cost":               []string{strconv.FormatInt(cost, 10)},
			"customer_reward_id": []string{customerRewardId},
		})
	return
}

//---------------------------------------------------------------------------//

type DeductInfo struct {
	Serial_num string `json:"excode"`
	Type       string `json:"type"`
	Uid        uint32 `json:"uid"`
	Money      int64  `json:"money"`
	Desc       string `json:"desc"`
	Details    string `json:"details"`
}

func DeductEx(c rpc.Client, host, excode, type_ string, uid uint32, money int64,
	desc, details string) (code int, err error) {

	code, err = c.CallWithJson(nil, host+"/deduct",
		DeductInfo{
			Serial_num: excode,
			Type:       type_,
			Uid:        uid,
			Money:      money,
			Desc:       desc,
			Details:    details,
		})
	return
}

func DeductCashEx(c rpc.Client, host, excode, type_ string, uid uint32, money int64,
	desc, details string) (code int, err error) {

	code, err = c.CallWithForm(nil, host+"/deduct_cash",
		map[string][]string{
			"excode":  []string{excode},
			"type":    []string{type_},
			"uid":     []string{strconv.FormatUint(uint64(uid), 10)},
			"money":   []string{strconv.FormatInt(money, 10)},
			"desc":    []string{desc},
			"details": []string{details},
		})
	return
}

func DeductDirect(c rpc.Client, host string, uid uint32, month string) (code int, err error) {
	code, err = c.CallWithForm(nil, host+"/deductdirect",
		map[string][]string{
			"uid":   []string{strconv.FormatUint(uint64(uid), 10)},
			"month": []string{month},
		})
	return
}

//---------------------------------------------------------------------------//

func GetInfoEx(c rpc.Client, host string, uid uint32) (info Info, code int, err error) {
	return GetInfo(c, host, uid)
}

//---------------------------------------------------------------------------//

func GetBillEx(c rpc.Client, host string, uid uint32, excode, prefix, type_ string) (
	bill BillEx, code int, err error) {

	values := Url.Values{}
	values.Add("uid", strconv.FormatUint(uint64(uid), 10))
	values.Add("excode", excode)
	values.Add("prefix", prefix)
	values.Add("type", type_)

	url := host + "/bill/get"
	url += "?" + values.Encode()
	code, err = c.Call(&bill, url)
	return
}

func GetBillBySN(c rpc.Client, host, serial_num string) (bill BillEx, code int, err error) {
	url := host + "/bill/getbysn"
	url += "?serial_num=" + serial_num
	code, err = c.Call(&bill, url)
	return
}

//---------------------------------------------------------------------------//

func GetBillsEx(c rpc.Client, host string, uid uint32, starttime, endtime int64,
	prefix, type_, expenses string, isprocessed string, offset, limit int64) (bills []BillEx, code int, err error) {
	values := Url.Values{}
	values.Add("uid", strconv.FormatUint(uint64(uid), 10))
	values.Add("starttime", strconv.FormatInt(starttime, 10))
	values.Add("endtime", strconv.FormatInt(endtime, 10))
	values.Add("prefix", prefix)
	values.Add("type", type_)
	values.Add("expenses", expenses)
	values.Add("isprocessed", isprocessed)
	if offset >= 0 {
		values.Add("offset", strconv.FormatInt(offset, 10))
	}
	if limit >= 0 {
		values.Add("limit", strconv.FormatInt(limit, 10))
	}
	url := host + "/bill/list" + "?" + values.Encode()
	code, err = c.Call(&bills, url)
	return
}

func GetRechargeList(c rpc.Client, host string, args map[string]string) (
	bills []BillEx, code int, err error) {
	values := make(Url.Values, len(args))
	for name, value := range args {
		values.Set(name, value)
	}
	url := host + "/recharge/list"
	if len(args) > 0 {
		url += "?" + values.Encode()
	}
	code, err = c.Call(&bills, url)
	return
}

//---------------------------------------------------------------------------//

type Arrearage struct {
	Uid            uint32            `json:"uid"`
	ArrearageAt    HundredNanoSecond `json:"arrearage_at"`
	LastInformedAt HundredNanoSecond `json:"lastinformed_at"`
}

type Cash struct {
	Uid   uint32 `json:"uid"`
	Money int64  `bson:"money"`
}

func GetArrearage(c rpc.Client, host string, uid uint32) (
	arr Arrearage, code int, err error) {
	url := host + "/arrearage/get"
	url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
	code, err = c.Call(&arr, url)
	return
}

func SetArrearage(c rpc.Client, host string, arrearage Arrearage) (code int, err error) {
	code, err = c.CallWithJson(nil, host+"/arrearage/set", arrearage)
	return
}

func DeleteArrearage(c rpc.Client, host string, uid uint32) (code int, err error) {
	code, err = c.CallWithForm(nil, host+"/arrearage/delete",
		map[string][]string{
			"uid": []string{strconv.FormatUint(uint64(uid), 10)},
		})
	return
}

func UpdateArrearage(c rpc.Client, host string, uid uint32, lastInformedAt HundredNanoSecond) (code int, err error) {
	code, err = c.CallWithForm(nil, host+"/arrearage/update",
		map[string][]string{
			"uid":            []string{strconv.FormatUint(uint64(uid), 10)},
			"lastinformedat": []string{strconv.FormatInt(int64(lastInformedAt), 10)},
		})
	return
}

func GetArrearageList(c rpc.Client, host string, offset, limit int) (
	arrs []Arrearage, code int, err error) {
	url := host + "/arrearage/list"
	url += "?offset=" + strconv.Itoa(offset)
	url += "&limit=" + strconv.Itoa(limit)
	code, err = c.Call(&arrs, url)
	return
}

func GetCashList(c rpc.Client, host string, offset, limit int) (cashList []Cash, code int, err error) {
	url := host + "/cash/list"
	url += "?offset=" + strconv.Itoa(offset)
	url += "&limit=" + strconv.Itoa(limit)
	code, err = c.Call(&cashList, url)
	return
}

//----------------------------------------------------------------------------//

type CouponChange struct {
	Id     string `json:"id"`
	Change int64  `json:"change"` // 0.0001yuan
}

type Snapshot struct {
	Uid            uint32         `json:"uid"`
	Month          string         `json:"month"` // like: 2006-01
	Money          int64          `json:"money"` // 0.0001
	Cash           int64          `json:"cash"`  // 0.0001
	CashEx         int64          `json:"cash_ex"`
	CashIn         int64          `json:"cash_in"`
	Coupon         int64          `json:"coupon"`
	CouponEx       int64          `json:"coupon_ex"`
	CouponIn       int64          `json:"coupon_in"`
	CouponExDetail []CouponChange `json:"coupon_ex_detail"`
	CouponInDetail []CouponChange `json:"coupon_in_detail"`
}

func GetMonthSnapshot(c rpc.Client, host string, uid uint32, month string) (
	snapshot Snapshot, code int, err error) {
	url := host + "/month_snapshot/get"
	url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
	url += "&month=" + month
	code, err = c.Call(&snapshot, url)
	return
}

func UpdateMonthSnapshot(c rpc.Client, host string, uid uint32, month string) (
	code int, err error) {
	code, err = c.CallWithForm(nil, host+"/month_snapshot/update",
		map[string][]string{
			"uid":   []string{strconv.FormatUint(uint64(uid), 10)},
			"month": []string{month},
		})
	return
}

func SetMonthSnapshot(c rpc.Client, host string, uid uint32, month string) (
	code int, err error) {
	code, err = c.CallWithForm(nil, host+"/month_snapshot/set",
		map[string][]string{
			"uid":   []string{strconv.FormatUint(uint64(uid), 10)},
			"month": []string{month},
		})
	return
}

//---------------------------------------------------------------------------//

// type DepositExInfo struct {
// 	Serial_num string `json:"serial_num"`
// 	Id         string `json:"id"`
// 	Type       string `json:"type"`
// 	Uid        uint32 `json:"uid"`
// 	Money      int64  `json:"money"`
// 	Deadtime   int64  `json:"deadtime"`
// 	Desc       string `json:"desc"`
// 	Details    string `json:"details"`
// }

// func DepositEx(c rpc.Client, host, serial_num, id, type_ string, uid uint32, money int64,
// 	deadtime int64, desc, details string) (code int, err error) {

// 	code, err = c.CallWithJson(nil, host+"/deposit",
// 		DepositExInfo{
// 			Serial_num: serial_num,
// 			Id:         id,
// 			Type:       type_,
// 			Uid:        uid,
// 			Money:      money,
// 			Deadtime:   deadtime,
// 			Desc:       desc,
// 			Details:    details,
// 		})
// 	return
// }

// //---------------------------------------------------------------------------//

// func GetDepositEx(c rpc.Client, host string, uid uint32) (info MoneyInfo, code int, err error) {
// 	code, err = c.Call(&info, host+"/get_deposit?uid="+strconv.FormatUint(uint64(uid), 10))
// 	return
// }

//---------------------------------------------------------------------------//

func NewCouponEx(c rpc.Client, host, excode, type_ string, quota int64,
	day int, deadtime int64, desc string) (id string, code int, err error) {
	return NewCoupon(c, host, excode, type_, quota, day, deadtime, desc)
}

//--------------------------------------------------------------------------//

func ActiveCouponEx(c rpc.Client, host, excode string, uid uint32, id string,
	desc string) (coupon Coupon, code int, err error) {
	return ActiveCoupon(c, host, excode, uid, id, desc)
}

//---------------------------------------------------------------------------//

func GetCouponsEx(c rpc.Client, host string, uid uint32) (
	coupons []Coupon, code int, err error) {
	return GetCoupons(c, host, uid)
}

//---------------------------------------------------------------------------//

func GetInactiveCouponsEx(c rpc.Client, host string) (
	coupons []Coupon, code int, err error) {
	return GetInactiveCoupons(c, host)
}

//---------------------------------------------------------------------------//
type UserInfoExt struct {
	Uid        uint32 `json:"uid" bson:"uid"`
	Email      string `json:"email" bson:"email"`
	Type       int    `json:"type" bson:"type"`
	Tag        int    `json:"tag" bson:"tag"`
	DeductType int    `json:"deduct_type" bson:"deduct_type"`
	Company    string `json:"company" bson:"company"`
	Remark     string `json:"remark"  bson:"remark"`
	NickName   string `json:"nickname" bson:"nickname"`
}

//---------------------------------------------------------------------------//
func GetUserInfoList(c rpc.Client, host string, args map[string]string) (
	userlist []UserInfoExt, err error) {
	params := make(Url.Values, len(args))

	for key, value := range args {
		params.Set(key, value)
	}

	url := host + "/user/list"
	if len(args) > 0 {
		url += "?" + params.Encode()
	}

	_, err = c.Call(&userlist, url)
	return
}

func GetUserInfo(c rpc.Client, host, uid string) (userinfo UserInfoExt, err error) {
	_, err = c.Call(&userinfo, host+"/user/get?uid="+uid)
	return
}

func UpdateUserInfo(c rpc.Client, host string, userInfoExt UserInfoExt) (code int, err error) {
	code, err = c.CallWithForm(nil, host+"/user/update",
		map[string][]string{
			"uid":         []string{strconv.FormatUint(uint64(userInfoExt.Uid), 10)},
			"company":     []string{userInfoExt.Company},
			"type":        []string{strconv.FormatInt(int64(userInfoExt.Type), 10)},
			"tag":         []string{strconv.FormatInt(int64(userInfoExt.Tag), 10)},
			"deduct_type": []string{strconv.FormatInt(int64(userInfoExt.DeductType), 10)},
			"remark":      []string{userInfoExt.Remark},
			"nickname":    []string{userInfoExt.NickName},
		})
	return
}

//---------------------------------------------------------------------------//
type WalletEnum struct {
	Value int64  `json:"value" bson:"value"`
	Desc  string `json:"desc"  bson:"desc"`
	Type  int64  `json:"type"  bson:"type"`
}

const (
	USERTAG = iota
)

func GetWalletEnumList(c rpc.Client, host string, etype int64) (enumlist []WalletEnum, err error) {
	_, err = c.Call(&enumlist, host+"/enum/list?type="+strconv.FormatInt(etype, 10))
	return
}

func AddWalletEnum(c rpc.Client, host string, value int64, desc string, etype int64) (err error) {
	_, err = c.CallWithForm(nil, host+"/enum/add",
		map[string][]string{
			"value": []string{strconv.FormatInt(value, 10)},
			"desc":  []string{desc},
			"type":  []string{strconv.FormatInt(etype, 10)},
		})
	return
}

func UpdateWalletEnum(c rpc.Client, host string, value int64, desc string, etype int64) (err error) {
	_, err = c.CallWithForm(nil, host+"/enum/update",
		map[string][]string{
			"value": []string{strconv.FormatInt(value, 10)},
			"desc":  []string{desc},
			"type":  []string{strconv.FormatInt(etype, 10)},
		})
	return
}

///////////////////////////////////////////////////////////////////////////////

type ServiceInEx struct {
	host string
	acc  account.InterfaceEx
}

func NewServiceInEx(host string, acc account.InterfaceEx) *ServiceInEx {
	return &ServiceInEx{host: host, acc: acc}
}

func (r *ServiceInEx) getClient(user account.UserInfo) rpc.Client {
	token := r.acc.MakeAccessToken(user)
	return rpc.Client{oauth.NewClient(token, nil)}
}

func (r *ServiceInEx) Recharge(user account.UserInfo, excode, type_ string, uid uint32,
	money int64, desc string) (code int, err error) {
	return RechargeEx(r.getClient(user), r.host, excode, type_, uid, money, desc)
}

func (r *ServiceInEx) RechargeReward(user account.UserInfo, excode, type_ string, uid uint32,
	money int64, desc string, cost int64, customerRewardId string) (code int, err error) {
	return RechargeReward(r.getClient(user), r.host, excode, type_, uid, money, desc, cost, customerRewardId)
}

func (r *ServiceInEx) Deduct(user account.UserInfo, excode, type_ string, uid uint32,
	money int64, desc, details string) (code int, err error) {
	return DeductEx(r.getClient(user), r.host, excode, type_, uid, money, desc, details)
}

func (r *ServiceInEx) DeductCash(user account.UserInfo, excode, type_ string, uid uint32,
	money int64, desc, details string) (code int, err error) {
	return DeductCashEx(r.getClient(user), r.host, excode, type_, uid, money, desc, details)
}

func (r *ServiceInEx) DeductDirect(user account.UserInfo, uid uint32, month string) (code int, err error) {
	return DeductDirect(r.getClient(user), r.host, uid, month)
}

func (r *ServiceInEx) GetInfo(user account.UserInfo, uid uint32) (info Info, code int, err error) {
	return GetInfoEx(r.getClient(user), r.host, uid)
}

func (r *ServiceInEx) GetBill(user account.UserInfo, uid uint32, excode, prefix, type_ string) (
	bill BillEx, code int, err error) {
	return GetBillEx(r.getClient(user), r.host, uid, excode, prefix, type_)
}

func (r *ServiceInEx) GetBillBySN(user account.UserInfo, serial_num string) (
	bill BillEx, code int, err error) {
	return GetBillBySN(r.getClient(user), r.host, serial_num)
}

func (r *ServiceInEx) GetBills(user account.UserInfo, uid uint32, starttime, endtime int64,
	prefix, type_, expenses, isprocessed string, offset, limit int64) (bills []BillEx, code int, err error) {
	return GetBillsEx(r.getClient(user), r.host, uid, starttime, endtime,
		prefix, type_, expenses, isprocessed, offset, limit)
}

//func (r *ServiceInEx) GetRechargeList(user account.UserInfo, args map[string]string) (
//bills []BillEx, code int, err error) {
//return GetRechargeList(r.getClient(user), r.host, args)
//}

func (r *ServiceInEx) GetArrearage(user account.UserInfo, uid uint32) (
	arr Arrearage, code int, err error) {
	return GetArrearage(r.getClient(user), r.host, uid)
}

func (r *ServiceInEx) SetArrearage(user account.UserInfo, arrearage Arrearage) (
	code int, err error) {
	return SetArrearage(r.getClient(user), r.host, arrearage)
}

func (r *ServiceInEx) UpdateArrearage(user account.UserInfo, uid uint32, lastInformedAt HundredNanoSecond) (
	code int, err error) {
	return UpdateArrearage(r.getClient(user), r.host, uid, lastInformedAt)
}

func (r *ServiceInEx) DeleteArrearage(user account.UserInfo, uid uint32) (
	code int, err error) {
	return DeleteArrearage(r.getClient(user), r.host, uid)
}

func (r *ServiceInEx) GetArrearageList(user account.UserInfo, offset, limit int) (
	arrs []Arrearage, code int, err error) {
	return GetArrearageList(r.getClient(user), r.host, offset, limit)
}

func (r *ServiceInEx) GetCashList(user account.UserInfo, offset, limit int) (
	cashs []Cash, code int, err error) {
	return GetCashList(r.getClient(user), r.host, offset, limit)
}

func (r *ServiceInEx) GetMonthSnapshot(user account.UserInfo, uid uint32, month string) (
	snapshot Snapshot, code int, err error) {
	return GetMonthSnapshot(r.getClient(user), r.host, uid, month)
}

func (r *ServiceInEx) SetMonthSnapshot(user account.UserInfo, uid uint32, month string) (
	code int, err error) {
	return SetMonthSnapshot(r.getClient(user), r.host, uid, month)
}

func (r *ServiceInEx) UpdateMonthSnapshot(user account.UserInfo, uid uint32, month string) (
	code int, err error) {
	return UpdateMonthSnapshot(r.getClient(user), r.host, uid, month)
}

// func (r *ServiceInEx) Deposit(user account.UserInfo, serial_num, id, type_ string, uid uint32,
// 	money int64, deadtime int64, desc, details string) (code int, err error) {
// 	return DepositEx(r.getClient(user), r.host, serial_num,
// 		id, type_, uid, money, deadtime, desc, details)
// }

// func (r *ServiceInEx) GetDeposit(user account.UserInfo, uid uint32) (
// 	info MoneyInfo, code int, err error) {
// 	return GetDepositEx(r.getClient(user), r.host, uid)
// }

func (r *ServiceInEx) NewCoupon(user account.UserInfo, excode, type_ string, quota int64,
	day int, deadtime int64, desc string) (id string, code int, err error) {
	return NewCouponEx(r.getClient(user), r.host, excode, type_, quota, day, deadtime, desc)
}

func (r *ServiceInEx) NewCouponNew(user account.UserInfo, title, type_ string, quota int64,
	day int, deadtime int64, desc string) (id string, code int, err error) {
	return NewCouponNew(r.getClient(user), r.host, title, type_, quota, day, deadtime, desc)
}

func (r *ServiceInEx) ActiveCoupon(user account.UserInfo, excode string, uid uint32,
	id, desc string) (coupon Coupon, code int, err error) {
	return ActiveCouponEx(r.getClient(user), r.host, excode, uid, id, desc)
}

func (r *ServiceInEx) GetCoupons(user account.UserInfo, uid uint32) (
	coupons []Coupon, code int, err error) {
	return GetCouponsEx(r.getClient(user), r.host, uid)
}

func (r *ServiceInEx) GetCoupon(user account.UserInfo, id string) (coupon Coupon, code int, err error) {
	return GetCoupon(r.getClient(user), r.host, id)
}

func (r *ServiceInEx) GetInactiveCoupons(user account.UserInfo) (
	coupons []Coupon, code int, err error) {
	return GetInactiveCouponsEx(r.getClient(user), r.host)
}

func (r *ServiceInEx) GetAdminCouponList(user account.UserInfo, uid, title string, type_ CouponType,
	status CouponStatus, offset, limit string) (coupons []Coupon, code int, err error) {
	return GetAdminCouponList(r.getClient(user), r.host, uid, title, type_, status, offset, limit)
}

func (r *ServiceInEx) GetAdminCouponCount(user account.UserInfo, uid, title string, type_ CouponType,
	status CouponStatus) (count, code int, err error) {
	return GetAdminCouponCount(r.getClient(user), r.host, uid, title, type_, status)
}

func (r *ServiceInEx) GetUserInfo(user account.UserInfo, uid string) (userinfo UserInfoExt, err error) {
	return GetUserInfo(r.getClient(user), r.host, uid)
}

//---------------------------------------------------------------------------//

type ServiceEx struct {
	Conn rpc.Client
}

func NewEx(t http.RoundTripper) ServiceEx {
	client := &http.Client{Transport: t}
	return ServiceEx{rpc.Client{client}}
}

func (r ServiceEx) Recharge(host string, excode, type_ string, uid uint32, money int64,
	desc string) (code int, err error) {
	return RechargeEx(r.Conn, host, excode, type_, uid, money, desc)
}

func (r ServiceEx) RechargeReward(host string, excode, type_ string, uid uint32, money int64,
	desc string, cost int64, customerRewardId string) (code int, err error) {
	return RechargeReward(r.Conn, host, excode, type_, uid, money, desc, cost, customerRewardId)
}

func (r ServiceEx) AddReward(host string, excode, serial_num, type_ string, uid uint32, money int64,
	desc string) (code int, err error) {
	return AddReward(r.Conn, host, excode, serial_num, type_, uid, money, desc)
}

//func (r ServiceEx) GetRewards(host string, money int64, offset, limit int) (rewardInfo []RewardInfo, code int, err error) {

//return GetRewards(r.Conn, host, money, offset, limit)
//}

func (r ServiceEx) GetRewardBySerialNum(host, serialNum string) (rewardInfo RewardInfo, code int, err error) {
	return GetRewardBySerialNum(r.Conn, host, serialNum)
}

func (r ServiceEx) Deduct(host string, excode, type_ string, uid uint32, money int64,
	desc, details string) (code int, err error) {
	return DeductEx(r.Conn, host, excode, type_, uid, money, desc, details)
}

func (r ServiceEx) DeductCash(host string, excode, type_ string, uid uint32, money int64,
	desc, details string) (code int, err error) {
	return DeductCashEx(r.Conn, host, excode, type_, uid, money, desc, details)
}

func (r ServiceEx) DeductDirect(host string, uid uint32, month string) (code int, err error) {
	return DeductDirect(r.Conn, host, uid, month)
}

func (r ServiceEx) GetInfo(host string, uid uint32) (info Info, code int, err error) {
	return GetInfoEx(r.Conn, host, uid)
}

func (r ServiceEx) GetBill(host string, uid uint32, excode, prefix, type_ string) (
	bill BillEx, code int, err error) {
	return GetBillEx(r.Conn, host, uid, excode, prefix, type_)
}

func (r ServiceEx) GetBillBySN(host, serial_num string) (
	bill BillEx, code int, err error) {
	return GetBillBySN(r.Conn, host, serial_num)
}

func (r ServiceEx) GetBills(host string, uid uint32, starttime, endtime int64,
	prefix, type_, expenses string, isprocessed string, offset, limit int64) (bills []BillEx, code int, err error) {
	return GetBillsEx(r.Conn, host, uid, starttime, endtime,
		prefix, type_, expenses, isprocessed, offset, limit)
}

func (r ServiceEx) GetRechargeList(host string, args map[string]string) (
	bills []BillEx, code int, err error) {
	return GetRechargeList(r.Conn, host, args)
}

func (r ServiceEx) GetArrearage(host string, uid uint32) (
	arr Arrearage, code int, err error) {
	return GetArrearage(r.Conn, host, uid)
}

func (r ServiceEx) SetArrearage(host string, arrearage Arrearage) (code int, err error) {
	return SetArrearage(r.Conn, host, arrearage)
}

func (r ServiceEx) UpdateArrearage(host string, uid uint32, lastInformedAt HundredNanoSecond) (code int, err error) {
	return UpdateArrearage(r.Conn, host, uid, lastInformedAt)
}

func (r ServiceEx) DeleteArrearage(host string, uid uint32) (code int, err error) {
	return DeleteArrearage(r.Conn, host, uid)
}

func (r ServiceEx) GetArrearageList(host string, offset, limit int) (
	arrs []Arrearage, code int, err error) {
	return GetArrearageList(r.Conn, host, offset, limit)
}

func (r ServiceEx) GetCashList(host string, offset, limit int) (
	cashs []Cash, code int, err error) {
	return GetCashList(r.Conn, host, offset, limit)
}

func (r ServiceEx) GetMonthSnapshot(host string, uid uint32, month string) (
	snapshot Snapshot, code int, err error) {
	return GetMonthSnapshot(r.Conn, host, uid, month)
}

func (r ServiceEx) SetMonthSnapshot(host string, uid uint32, month string) (
	code int, err error) {
	return SetMonthSnapshot(r.Conn, host, uid, month)
}

func (r ServiceEx) UpdateMonthSnapshot(host string, uid uint32, month string) (
	code int, err error) {
	return UpdateMonthSnapshot(r.Conn, host, uid, month)
}

// func (r ServiceEx) Deposit(host string, serial_num, id, type_ string, uid uint32, money int64,
// 	deadtime int64, desc, details string) (code int, err error) {
// 	return DepositEx(r.Conn, host, serial_num, id, type_, uid, money, deadtime, desc, details)
// }

// func (r ServiceEx) GetDeposit(host string, uid uint32) (info MoneyInfo, code int, err error) {
// 	return GetDeposit(r.Conn, host, uid)
// }

func (r ServiceEx) NewCoupon(host string, excode, type_ string, quota int64,
	day int, deadtime int64, desc string) (id string, code int, err error) {
	return NewCouponEx(r.Conn, host, excode, type_, quota, day, deadtime, desc)
}

func (r ServiceEx) NewCouponNew(host string, title, type_ string, quota int64,
	day int, deadtime int64, desc string) (id string, code int, err error) {
	return NewCouponNew(r.Conn, host, title, type_, quota, day, deadtime, desc)
}

func (r ServiceEx) ActiveCoupon(host string, excode string, uid uint32, id string,
	desc string) (coupon Coupon, code int, err error) {
	return ActiveCouponEx(r.Conn, host, excode, uid, id, desc)
}

func (r ServiceEx) GetCoupons(host string, uid uint32) (
	coupons []Coupon, code int, err error) {
	return GetCouponsEx(r.Conn, host, uid)
}

func (r ServiceEx) GetCoupon(host string, id string) (coupon Coupon, code int, err error) {
	return GetCoupon(r.Conn, host, id)
}

func (r ServiceEx) GetInactiveCoupons(host string) (
	coupons []Coupon, code int, err error) {
	return GetInactiveCoupons(r.Conn, host)
}

func (r ServiceEx) GetAdminCouponList(host string, uid, title string, type_ CouponType, status CouponStatus,
	offset, limit string) (coupons []Coupon, code int, err error) {
	return GetAdminCouponList(r.Conn, host, uid, title, type_, status, offset, limit)
}

func (r ServiceEx) GetAdminCouponCount(host, uid, title string, type_ CouponType, status CouponStatus) (
	count, code int, err error) {
	return GetAdminCouponCount(r.Conn, host, uid, title, type_, status)
}

func (r ServiceEx) GetUserInfoList(host string, args map[string]string) (
	userlist []UserInfoExt, err error) {
	return GetUserInfoList(r.Conn, host, args)
}

func (r ServiceEx) GetUserInfo(host, uid string) (userinfo UserInfoExt, err error) {
	return GetUserInfo(r.Conn, host, uid)
}
