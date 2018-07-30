package v2

import (
	"strconv"
	"time"
)

import (
	"labix.org/v2/mgo/bson"
	P "qbox.us/api/pay/price/v2"
)

//----------------------------------------------------------------------------//
type ApplicablePeopleType int

func (a ApplicablePeopleType) ToString() string {
	return strconv.FormatInt(int64(a), 10)
}

type BillType int

const (
	BILLTYPE_BASEBILL        BillType = 1
	BILLTYPE_BASEBILL_FORMAT BillType = 2
)

var BillTypeM = map[BillType]string{
	BILLTYPE_BASEBILL:        "基础账单",
	BILLTYPE_BASEBILL_FORMAT: "基础格式化",
}

func (bt BillType) ToString() string {
	return strconv.Itoa(int(bt))
}

func (bt BillType) Desc() string {
	return BillTypeM[bt]
}

type CouponType string

func (t CouponType) ToString() string {
	return string(t)
}

type CouponStatus int

func (c CouponStatus) ToString() string {
	return strconv.FormatInt(int64(c), 10)
}

type Second int64

var ZERO_SECOND Second = 0

func NewSecond(t time.Time) Second {
	return Second(t.Unix())
}

func (a Second) Time() time.Time {
	return time.Unix(int64(a), 0)
}

func (a Second) ToString() string {
	return strconv.FormatInt(int64(a), 10)
}

func (r *Second) Value() int64 {
	if r == nil {
		return int64(0)
	}
	return int64(*r)
}

type HundredNanoSecond int64

func NewHundredNanoSecond(t time.Time) HundredNanoSecond {
	return HundredNanoSecond(t.UnixNano() / 100)
}

func (a HundredNanoSecond) Time() time.Time {
	return time.Unix(int64(a)/1e7, int64(a)%1e7*100)
}

func (a HundredNanoSecond) ToString() string {
	return strconv.FormatInt(int64(a), 10)
}

//----------------------------------------------------------------------------//

type Money int64 //精确到0.01分

func (m Money) ToString() string {
	return strconv.FormatInt(int64(m), 10)
}

func (m Money) AsYuan() float64 {
	return float64(m) / 10000
}

////////////////////////////////////////////////////////////////////////////////

type DataType string

const (
	DATA_TYPE_SPACE     DataType = "space"
	DATA_TYPE_TRANSFER  DataType = "transfer"
	DATA_TYPE_BANDWIDTH DataType = "bandwidth"
	DATA_TYPE_APIGET    DataType = "api_get"
	DATA_TYPE_APIPUT    DataType = "api_put"
	DATA_TYPE_SERVICE   DataType = "service"
)

func (r DataType) ToString() string {
	return string(r)
}

//--------------------------//
const (
	DEDUCT_BY_DAY_AVERAYGE string = "day_average"
	DEDUCT_BY_ACCUMULATE   string = "accumulate"
	DEDUCT_BY_MONTH        string = "month"
)

//----------------------------------------------------------------------------//

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

var MonthBillStatusStringArray = []string{
	"新建", "账单已通知", "已扣费", "申述", "已扣费通知", "已支付",
	"未出账", "已出账", "已支付",
}

func (s MonthBillStatus) ToString() string {
	return strconv.FormatInt(int64(s), 10)
}

func (s MonthBillStatus) String() string {
	return MonthBillStatusStringArray[int(s)]
}

// 月对账单
type MonthStatementStatus int

const (
	MONTH_STATEMENT_STATUS_NOT_OUT   MonthStatementStatus = 1
	MONTH_STATEMENT_STATUS_CHARGEOFF MonthStatementStatus = 2 //月对账单出帐
	MONTH_STATEMENT_STATUS_PAID      MonthStatementStatus = 3 //月对账单已支付
)

var MonthStatementStatusM = map[MonthStatementStatus]string{
	MONTH_STATEMENT_STATUS_NOT_OUT:   "未出账",
	MONTH_STATEMENT_STATUS_CHARGEOFF: "已出账",
	MONTH_STATEMENT_STATUS_PAID:      "已支付",
}

func (s MonthStatementStatus) ToString() string {
	return strconv.FormatInt(int64(s), 10)
}

func (ms MonthStatementStatus) Desc() string {
	return MonthStatementStatusM[ms]
}

type UserType int

func (s UserType) ToString() string {
	return strconv.FormatInt(int64(s), 10)
}

type UserTag int64

func (s UserTag) ToString() string {
	return strconv.FormatInt(int64(s), 10)
}

type UserDeductType int

func (s UserDeductType) ToString() string {
	return strconv.FormatInt(int64(s), 10)
}

//----------------------------------------------------------------------------//

type FreezeType int

const (
	FREEZE_TYPE_IGNORE FreezeType = 0
)

func (r FreezeType) ToString() string {
	return strconv.Itoa(int(r))
}

type FreezeStep int

const (
	FREEZE_STEP_IGNORE FreezeStep = 0
	FREEZE_STEP_NEW    FreezeStep = 1
	FREEZE_STEP_INFORM FreezeStep = 2
	FREEZE_STEP_DONE   FreezeStep = 3
	FREEZE_STEP_DROP   FreezeStep = 4
)

func (r FreezeStep) ToString() string {
	return strconv.Itoa(int(r))
}

//----------------------------------------------------------------------------//

type Version string

const (
	V1 Version = "v1"
	V2 Version = "v2"
	V3 Version = "v3"
	V4 Version = "v4"
	V5 Version = "v5"
)

func (r Version) ToString() string {
	return string(r)
}

//----------------------------------------------------------------------------//
// 流水的类型
type BILL_TYPE string

const (
	BILL_RECHARGE_BANK     BILL_TYPE = "BANK"     // 银行充值
	BILL_RECHARGE_ALIPAY   BILL_TYPE = "alipay"   // 支付宝充值
	BILL_RECHARGE_PRESENT  BILL_TYPE = "PRESENT"  // 充值赠送
	BILL_RECHARGE_WRITEOFF BILL_TYPE = "writeoff" // 充值销账

	BILL_DEDUCT_DAILY   BILL_TYPE = "storage-day"   // 每日扣费
	BILL_DEDUCT_MONTH   BILL_TYPE = "storage-month" // 每月扣费
	BILL_DEDUCT_EXPRESS BILL_TYPE = "发票快递扣费"        // 发票快递扣费 TODO: 这中文。。。

	BILL_COUPON_ACTIVE  = "active"  // 激活优惠劵
	BILL_COUPON_DISCARD = "discard" // 优惠劵过期

	// 历史残留
	BILL_RECHARGE_ALIPAY_OLD   = "ALIPAY_RECHARGE" // 支付宝充值
	BILL_RECHARGE_ALIPAY_OLD2  = "recharge"        // 支付宝充值
	BILL_RECHARGE_BANK_OLD     = "bank"            // 银行充值
	BILL_RECHARGE_WRITEOFF_OLD = "writeoff"        // 充值销账
	BILL_RECHARGE_TEST         = "test"            // 测试用流水
	BILL_RECHARGE_TEST2        = "TEST"            // 测试用流水
	BILL_DEDUCT_COUPON         = "COUPON"          // 购买优惠劵，为了提前开发票
	BILL_DEDUCT_PACKAGE        = "package"         // 曾经的套餐扣费
)

////////////////////////////////////////////////////////////////////////////////
// batch info uid max len
const (
	MaxUidLen = 200
)

//激活优惠券
type ActiveCouponIn struct {
	Excode string `json:"excode"`
	Uid    uint32 `json:"uid"`
	Id     string `json:"id"`
	Desc   string `json:"desc"`
}

//充值优惠券
type AddRewardIn struct {
	Excode    string  `json:"excode"`
	Desc      string  `json:"desc"`
	Type      string  `json:"type"`
	Uid       uint32  `json:"uid"`
	Money     Money   `json:"money"`
	At        *Second `json:"at"` // 充值操作时间
	SerialNum string  `json:"serial_num"`
}

//----------------------------------------------------------------------------//

// 欠费信息
type ArrearageInfo struct {
	Arrear      Money             `json:"arrear"`
	Time        HundredNanoSecond `json:"time"`        // 开始欠款时间
	IsRecharged bool              `json:"isrecharged"` // 是否付费过
}

// 欠费用户信息
type ArrearageUserInfo struct {
	Uid           uint32            `json:"uid"`
	Email         string            `json:"email"`         // 便于信息展示
	CustomerGroup int               `json:"customergroup"` // 用户类别
	Arrear        Money             `json:"arrear"`
	Time          HundredNanoSecond `json:"time"`        // 开始欠款时间
	IsRecharged   bool              `json:"isrecharged"` // 是否付费过
}

// 列取欠费用户信息
type ArrearageUserLister struct {
	CustomerGroup *int               `json:"customergroup"` // 用户类别
	From          *HundredNanoSecond `json:"from"`          // 欠款时间区间开始，0表示忽略此参数
	To            *HundredNanoSecond `json:"to"`            // 欠款时间区间结束，0表示忽略此参数
	Max           *Money             `json:"max"`           // 欠费额最大值，注意欠费额为负值，0表示忽略此参数
	Min           *Money             `json:"min"`           // 欠费额最小值，注意欠费额为负值，0表示忽略此参数
	Offset        *int64             `json:"offset"`
	Limit         *int64             `json:"limit"`
}

// 统一的账单格式
type BaseBillSet struct {
	Uid     uint32         `json:"uid"`
	Field   DataType       `json:"field"`
	From    string         `json:"from"` // 2006-01-02
	To      string         `json:"to"`   // 2006-01-02
	Money   Money          `json:"money"`
	Desc    string         `json:"desc"`
	Details BaseBillDetail `json:"details"`
	Version Version        `json:"version"`
}

// format 接口
type BaseBillFormat struct {
	Money   Money                  `json:"money"`
	Details []BaseBillDetailFormat `json:"details"`
}

type BaseBillDetailFormat struct {
	Start     string               `json:"start"`      // 2006-01-02
	End       string               `json:"end"`        // 2006-01-02
	ValueType string               `json:"value_type"` // 计费模式，空间:日均，流量/请求：累积,带宽：top95...
	Money     Money                `josn:"money"`
	Value     int64                `json:"value"`
	Units     BaseBillDetailUnits  `json:"units"`
	Rewards   []BaseBillRewardUnit `json:"rewards"`
	Discounts []BaseBillDiscount   `json:"discounts"`
}

type BaseBillGet struct {
	Id        string            `json:"id"`
	Uid       uint32            `json:"uid"`
	Field     DataType          `json:"field"`
	From      string            `json:"from"` // 2006-01-02
	To        string            `json:"to"`   // 2006-01-02
	Money     Money             `json:"money"`
	Desc      string            `json:"desc"`
	Details   BaseBillDetail    `json:"details"`
	CreatedAt HundredNanoSecond `json:"create_at"`
	UpdatedAt HundredNanoSecond `json:"update_at"`
	Version   Version           `json:"version"`
}

type BaseBillDetail struct {
	Units     BaseBillDetailUnits            `json:"units"`
	Rewards   []BaseBillRewardUnit           `json:"rewards"`
	Discounts []BaseBillDiscount             `json:"discounts"`
	Price     P.UserPriceInFieldAndTimeRange `json:"price"`
	Value     int64                          `json:"value"`
	AccuValue int64                          `json:"accu_value"`
}

type BaseBillDiscard struct {
	Uid uint32 `json:"uid"`
	Day string `json:"day"` // format: 2006-01-02
}

type BaseBillDiscount MonthBillDiscount

type BaseBillDetailUnits struct {
	Type       string                     `json:"type"` // UNITPRICE, FIRST_BUYOUT, EACH_BUYOUT ...
	Units      []BaseBillDetailUnit       `json:"units"`
	DailyUnits []BaseBillDetailDailyUnits `json:"daily_units"`
}

type BaseBillDetailUnit struct {
	From     int64 `json:"from"`
	To       int64 `json:"to"`
	Price    Money `json:"price"`
	Value    int64 `json:"value"`     // 实际产生费用部分
	Money    Money `json:"money"`     // 实际费用
	AllValue int64 `json:"all_value"` // 全额使用量，包含未产生最终费用的部分
	AllMoney Money `json:"all_money"` // 全额使用量对应的收入费用
}

type BaseBillDetailDailyUnits struct {
	Day   string               `json:"day"` // format: 2006-01-02
	Units []BaseBillDetailUnit `json:"units"`
}

type PriceId string

type BaseBillRewardUnit struct {
	Id      string       `json:"id"`
	OpId    string       `json:"opid"`
	Type    string       `json:"type"`
	Desc    string       `json:"desc"`
	Quota   int64        `json:"quota"`
	Value   int64        `json:"value"`
	Balance int64        `json:"balance"`
	Reduce  Money        `json:"reduce"`
	Details []RangeMoney `json:"details"` //阶梯使用详情
	Overdue bool         `json:"overdue"`
}

type RangeMoney struct {
	From   int64 `json:"from"`
	To     int64 `json:"to"`
	Value  int64 `json:"value"`
	Reduce Money `json:"money"` // 实际费用
}

type Refund struct {
	Uid       uint32  `json:"uid"`
	Money     Money   `json:"money"`      //用户充值付的钱
	FreeMoney Money   `json:"free_money"` //用户充值赠送的钱
	Excode    string  `json:"excode"`
	At        *Second `json:"at"` // 业务相关时间，退款实际发生时间 ?? 还是充值时间，待确认
	Desc      string  `json:"desc"`
	Detail    string  `json:"detail"`
}

// 获取指定日期之前的最后一份账单的参数，如果不指定日期，代表当前时间
type BaseBillLastGetter struct {
	Uid   uint32   `json:"uid"`
	Field DataType `json:"field"`
	Date  *string  `json:"date"` // 2006-01-02
}

type BaseBillGetter struct {
	Id string `json:"id"`
}

type BillDiscountGetIn struct {
	SerialNum string `json:"serial_num"`
}

type BillDiscountListIn struct {
	RewardId    *string `json:"reward_id"`
	PartnerName *string `json:"partner_name"`
	From        *string `json:"from"` //20060102
	To          *string `json:"to"`   //20060102
	Offset      *int    `json:"offset"`
	Limit       *int    `json:"limit"`
}

type BillDiscountSumIn struct {
	RewardId    string `json:"reward_id"`
	PartnerName string `json:"partner_name"`
	From        string `json:"from"` //20060102
	To          string `json:"to"`   //20060102
}

type BillGetbysn struct {
	SerialNum string `json:"serial_num"`
}

type BillInfo struct {
	Details interface{} `json:"details"` // BASEBILL: map[DataType][]BaseBill, BASEBILL_FORMAT:map[DataType]BaseBillFormat
	Type    BillType    `json:"type"`    // BASEBILL, BASEBILL_FORMAT
}

type CashDetail struct {
	Before int64 `json:"before" bson:"before"`
	Change int64 `json:"change" bson:"change"`
	After  int64 `json:"after" bson:"after"`
}

type CashListIn struct {
	Offset *int `json:"offset"`
	Limit  *int `json:"limit"`
}

type CashListOut struct {
	Uid   uint32 `json:"uid"`
	Money Money  `json:"money"`
}

type Coupon struct {
	Quota      Money             `json:"quota"`
	Balance    int64             `json:"balance"`
	CreateAt   HundredNanoSecond `json:"create_at"`
	UpdateAt   HundredNanoSecond `json:"update_at"`
	EffectTime HundredNanoSecond `json:"effecttime"`
	DeadTime   HundredNanoSecond `json:"deadtime"`
	Uid        uint32            `json:"uid"`
	Day        int               `json:"day"`
	Id         string            `json:"id"`
	Title      string            `json:"title"`
	Desc       string            `json:"desc"`
	Type       CouponType        `json:"type"`
	Status     CouponStatus      `json:"status"`
}

//激活优惠券
type CouponActiveIn struct {
	Excode string `json:"excode"`
	Uid    uint32 `json:"uid"`
	Id     string `json:"id"`
	Desc   string `json:"desc"`
}

//根据查询条件获取优惠券
type CouponAdminListIn struct {
	Offset *int    `json:"offset"`
	Limit  *int    `json:"limit"`
	Uid    int64   `json:"uid"`
	Type   string  `json:"type"`
	Status int     `json:"status"`
	Title  *string `json:"title"`
}

//根据查询条件获取优惠券
type CouponAdminCountIn struct {
	Uid    int64  `json:"uid"`
	Type   string `json:"type"`
	Status int    `json:"status"`
	Title  string `json:"title"`
}

type CouponChange struct {
	Id     string `json:"id"`
	Change Money  `json:"change"`
}

type CouponDetail struct {
	Id     string `json:"id" bson:"id"`
	Before int64  `json:"before" bson:"before"`
	Change int64  `json:"change" bson:"change"`
	After  int64  `json:"after" bson:"after"`
}

//获取优惠券
type CouponGetIn struct {
	Id string `json:"id"`
}

type CouponHistory struct {
	Uid    uint32 `json:"uid"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

//新建优惠券
type CouponNewIn struct {
	Quota    Money  `json:"quota"`
	Day      int    `json:"day"`
	DeadTime int64  `json:"deadtime"`
	Type     string `json:"type"`
	Desc     string `json:"desc"`
	Title    string `json:"title"`
}

//月清单对应的帐单列表
type BaseBillsForMonth struct {
	Uid   uint32 `json:"uid"`
	Month string `json:"month"` //2006-01
}

// 获取指定月份之前的月账单和月对账单月份列表
type MonthBillMonthStatementMonthsIn struct {
	Uid       uint32 `json:"uid"`
	LastMonth string `json:"lastmonth"` //format: 2006-01
	Limit     int    `json:"limit"`
}

type MonthOut struct {
	Month       HundredNanoSecond `json:"month"`
	IsMonthBill bool              `json:"is_monthbill"` // true:monthbill, false:monthstatement
}

type MonthBillMonthStatementMonthsOut struct {
	Months []MonthOut `json:"months"`
}

type MonthStatementBills struct {
	IsPaid bool          `json:"ispaid"`
	List   []BaseBillGet `json:"list"`
}

type MonthStatement struct {
	Id        string     `json:"id" bson:"_id"`
	Uid       uint32     `json:"uid" bson:"uid"`
	Month     string     `json:"month" bson:"month"` // 2006-01
	Desc      string     `json:"desc" bson:"desc"`
	Status    int        `json:"status" bson:"status"`
	Money     int64      `json:"money" bson:"money"`
	Bills     []BillInfo `json:"bills" bson:"bills"`
	CreatedAt Second     `json:"create_at"`
	Version   string     `json:"version" bson:"version"`
}

// 获取月对账单，其中(Uid+Month)和Id二选一
type MonthStatementGetter struct {
	Uid   uint32 `json:"uid"`
	Month string `json:"month"` // format: 2006-01, default: ""
	Id    string `json:"id"`    // default: ""
}

// 批量获取用户的月对账单
type MonthStatementGetterList struct {
	Uids  string `json:"uids"`  // split by ","
	Month string `json:"month"` // format: 2006-01
}

// 获取月对账单列表，其中字段皆为可选
type MonthStatementLister struct {
	Uid    *uint32               `json:"uid"`
	From   *string               `json:"from"` // format: 2006-01
	To     *string               `json:"to"`   // format: 2006-01
	Status *MonthStatementStatus `json:"status"`
	Offset *int                  `json:"offset"`
	Limit  *int                  `json:"limit"`
}

// 获取指定月份之前的月对账单月份列表
type MonthStatementMonthsIn struct {
	Uid       uint32 `json:"uid"`
	LastMonth string `json:"lastmonth"` //format: 2006-01
	Limit     int    `json:"limit"`
}

//CustomerReward interface 对应的类型 ===
type Range struct {
	StartPrice int64 `json:"start_price" bson:"start_price"`
	EndPrice   int64 `json:"end_price" bson:"end_price"`
	Reward     int64 `json:"reward" bson:"reward"`
}

type DiscountReward struct {
	Discount int64 `json:"discount" bson:"discount"`
}

type SpecificPriceReward struct {
	SpecificPrice Range `json:"specific_price" bson:"specific_price"`
}

type RangePriceReward struct {
	RangePrice []Range `json:"range_price" bson:"range_price"`
}

//==========

type CustomerReward struct {
	Id                 string               `json:"id"`
	Title              string               `json:"title"`
	Desc               string               `json:"desc"`
	PartnerName        string               `json:"partner_name"`
	PartnerRewardId    string               `json:"partner_reward"`
	CustomerRewardType string               `json:"client_reward_type"`
	CustomerReward     interface{}          `json:"client_reward"`
	EffectTime         int64                `json:"effect_time"`
	Deadline           int64                `json:"deadline"`
	CreateAt           int64                `json:"create_at"`
	UpdateAt           int64                `json:"update_at"`
	ApplicablePeople   ApplicablePeopleType `json:"applicable_people"`
	IsAvaliable        bool                 `json:"is_avaliable"`
	Version            string               `json:"version"`
}

type CustomerRewardAvailableIn struct {
	Id        string `json:"id"`
	Available bool   `json:"available"`
}

type CustomerRewardCalculateIn struct {
	Uid  uint32 `json:"uid"`
	Id   string `json:"id"`
	Cost Money  `json:"cost"`
}

type CustomerRewardGetIn struct {
	Id string `json:"id"`
}

type CustomerRewardListIn struct {
	Name   *string `json:"name"`
	Offset *int    `json:"offset"`
	Limit  *int    `json:"limit"`
}

type CustomerRewardUpdatedescIn struct {
	Id   string `json:"id"`
	Desc string `json:"desc"`
}

// 日账单
type DailyBill struct {
	Id      bson.ObjectId    `json:"id"`
	Uid     uint32           `json:"uid"`
	Day     string           `json:"day"` // format: 2006-01-02
	Desc    string           `json:"desc"`
	Money   Money            `json:"money"`
	Details MonthBillDetails `json:"details"`
	Version Version          `json:"version"`
}

type DailyBillDiscard struct {
	Uid  uint32 `json:"uid"`
	Date string `json:"date"` // format: 2006-01-02
}

// 日账单获取
type DailyBillGetter struct {
	Uid uint32 `json:"uid"`
	Day string `json:"day"` // format: 20060102
}

// 日账单设置
type DailyBillSetter struct {
	Uid     uint32  `json:"uid"`
	Day     string  `json:"day"` // format: 2006-01-02
	Desc    string  `json:"desc"`
	Money   Money   `json:"money"`
	Details string  `json:"details"` // MonthBillDetail struct json encoded
	Version Version `json:"version"`
}

// 日账单列表
type DailyBillGetterRange struct {
	Uid  uint32 `json:"uid"`
	From string `json:"from"` // format: 2006-01-02
	To   string `json:"to"`
}

type DiscountTransaction struct {
	Serial_num  string `json:"serial_num"`
	Time        int64  `json:"time"`
	PartnerName string `json:"partner_name"`
	CustomerID  string `json:"customer_id"`
	PartnerID   string `json:"partner_id"`
	Money       int64  `json:"money"`
	Pay         int64  `json:"pay"`
	Cost        int64  `json:"cost"`
}

type DiscountTransactionSummary struct {
	Count int   `json:"count"`
	Money int64 `json:"money"`
	Pay   int64 `json:"pay"`
	Cost  int64 `json:"cost"`
}

type DeductCashIn struct {
	Excode  string  `json:"excode"`
	Type    string  `json:"type"`
	Uid     uint32  `json:"uid"`
	Money   Money   `json:"money"`
	At      *Second `json:"at"` // 业务扣费时间
	Desc    string  `json:"desc"`
	Details string  `json:"details"`
}

//扣费
type DeductIn struct {
	Excode  string  `json:"excode"`
	Type    string  `json:"type"`
	Uid     uint32  `json:"uid"`
	Money   int64   `json:"money"`
	At      *Second `json:"at"` // 业务扣费时间
	Desc    string  `json:"desc"`
	Details string  `json:"details"`
}

//获取帐单
type BillGetIn struct {
	Uid    *uint32 `json:"uid"`
	Excode string  `json:"excode"`
	Prefix string  `json:"prefix"`
	Type   string  `json:"type"`
}

// 冻结用户信息
type FreezeUserInfo struct {
	Uid           uint32            `json:"uid"`
	Email         string            `json:"email"`
	CustomerGroup int               `json:"customergroup"`
	Arrear        Money             `json:"arrear"`
	Time          HundredNanoSecond `json:"time"` // 开始欠款时间
	Step          FreezeStep        `json:"step"`
	Type          FreezeType        `json:"type"`
	RemindAt      HundredNanoSecond `json:"remind_at"`  // 提醒日期
	DisableAt     HundredNanoSecond `json:"disable_at"` // 冻结日期
	UpdateAt      HundredNanoSecond `json:"update_at"`  // 最后修改时间,设置时指定无效
}

// 列取冻结用户信息
type FreezeUserLister struct {
	Step   FreezeStep `json:"step"`
	Offset int64      `json:"offset"` // 0为无效
	Limit  int64      `json:"limit"`  // 0为无效
}

//获取帐单列表
type Get_billsIn struct {
	Uid         *uint32           `json:"uid"`
	StartTime   HundredNanoSecond `json:"starttime"`
	EndTime     HundredNanoSecond `json:"endtime"`
	Prefix      string            `json:"prefix"`
	Type        string            `json:"type"`
	Expenses    string            `json:"expenses"`
	IsProcessed *string           `json:"isprocessed"`
	IsHide      *bool             `json:"ishide"`
	Offset      *int64            `json:"offset"`
	Limit       *int64            `json:"limit"`
}

//隐藏帐单
type HideBillIn struct {
	Id     string `json:"id"`
	IsHide bool   `json:"ishide"`
}

type InfoOut struct {
	Amount    Money `json:"amount"`     //coupon + cash
	Cash      Money `json:"cash"`       //cash = CostMoney + FreeMoney
	Coupon    Money `json:"coupon"`     //优惠券
	CostMoney Money `json:"cost_money"` //用户充值的钱
	FreeMoney Money `json:"free_money"` //用户赠送的NB
}

type InfoOuts struct {
	Infos []InfoOut `json:"infos"`
}

// 月账单
type MonthBill struct {
	Id               bson.ObjectId     `json:"id"`
	Uid              uint32            `json:"uid"`
	Month            string            `json:"month"` // format: 2006-01
	Day              int               `json:"day"`   // 月内日期  1,2,3...
	CreatedAt        HundredNanoSecond `json:"create_at"`
	InformedAt       HundredNanoSecond `json:"inform_at"`
	DeductedAt       HundredNanoSecond `json:"deduct_at"`
	ComplainedAt     HundredNanoSecond `json:"complain_at"`
	InformDeductedAt HundredNanoSecond `json:"inform_deduct_at"`
	PaidAt           HundredNanoSecond `json:"pay_at"`
	UpdatedAt        HundredNanoSecond `json:"update_at"`
	ChargeoffAt      HundredNanoSecond `json:"chargeoff_at"`
	Status           MonthBillStatus   `json:"status"`
	IsInformed       bool              `json:"is_informed"`
	IsInformDeducted bool              `json:"is_informed"`
	Desc             string            `json:"desc"`
	Money            Money             `json:"money"`
	Details          MonthBillDetails  `json:"details"`
	Version          Version           `json:"version"`
}

type MonthBillDetail struct {
	Type    string                `json:"type"`
	Desc    string                `json:"desc"`
	Value   int64                 `json:"value"`
	Money   Money                 `json:"money"`
	Units   []MonthBillDetailUnit `json:"units"`
	Rewards []MonthBillDetailUnit `json:"rewards"`
}

type MonthBillDetails struct {
	Money     Money                        `json:"money"`
	Details   map[DataType]MonthBillDetail `json:"details"`
	Discounts []MonthBillDiscount          `json:"discounts"`
}

type MonthBillDetailUnit struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Desc   string `json:"desc"`
	From   int64  `json:"from"`
	To     int64  `json:"to"`
	Price  Money  `json:"price"`
	Value  int64  `json:"value"`
	All    Money  `json:"all"`
	Money  Money  `json:"money"`
	Reduce Money  `json:"reduce"`
}

type MonthBillDiscard struct {
	Uid   uint32 `json:"uid"`
	Month string `json:"month"` // format: 2006-01
}

type MonthBillDiscount struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Before  Money  `json:"before"`
	Change  Money  `json:"change"`
	Percent int64  `json:"percent"`
	After   Money  `json:"after"`
}

// 格式化后的月账单
type MonthBillFormated struct {
	Id      string                   `json:"id"`
	Uid     uint32                   `json:"uid"`
	Month   string                   `json:"month"` // format: 2006-01
	Status  MonthBillStatus          `json:"status"`
	Desc    string                   `json:"desc"`
	Money   Money                    `json:"money"`
	Details MonthBillFormatedDetails `json:"details"`
}

type MonthBillFormatedDetail struct {
	Desc  string                        `json:"desc"`
	Value string                        `json:"value"`
	Money Money                         `json:"money"`
	Units []MonthBillFormatedDetailUnit `json:"units"`
}

type MonthBillFormatedDetails struct {
	Money     Money                                `json:"money"`
	Details   map[DataType]MonthBillFormatedDetail `json:"details"`
	Discounts []MonthBillFormatedDiscount          `json:"discounts"`
}

type MonthBillFormatedDetailUnit struct {
	Desc   string `json:"desc"`
	Value  string `json:"value"`
	All    Money  `json:"all"`
	Money  Money  `json:"money"`
	Reduce Money  `json:"reduce"`
}

type MonthBillFormatedDiscount struct {
	Desc  string `json:"desc"`
	Money Money  `json:"money"`
}

type MonthBillGetter struct {
	Uid   uint32 `json:"uid"`
	Month string `json:"month"` // format: 200601
	Id    string `json:"id"`
}

type MonthBillGetterList struct {
	Uids  string `json:"uids"`
	Month string `json:"month"` // format: 200601
}

type MonthBillListerForAdmin struct {
	Uid    *int64           `json:"uid"`
	Month  *string          `json:"month"`
	Status *MonthBillStatus `json:"status"`
	Offset *int             `json:"offset"`
	Limit  *int             `json:"limit"`
}

type MonthBillLister struct {
	Uid    *int64  `json:"uid"`
	From   *string `json:"from"` // format: 200601
	To     *string `json:"to"`   // format: 200601
	Offset *int    `json:"offset"`
	Limit  *int    `json:"limit"`
}

type Month_billMonthsIn struct {
	Uid       *uint32 `json:"uid"`
	LastMonth string  `json:"lastmonth"` //format: 20060102
	Limit     int     `json:"limit"`
}

// 月账单设置
type MonthBillSetter struct {
	Uid     uint32  `json:"uid"`
	Month   string  `json:"month"` // format: 2006-01
	Day     int     `json:"day"`   // 月内日期  1,2,3,4...
	Desc    string  `json:"desc"`
	Money   Money   `json:"money"`
	Details string  `json:"details"` // MonthBillDetail struct json encoded
	Version Version `json:"version"`
}

// 月账单状态设置
type MonthBillStatusSetter struct {
	Uid    uint32          `json:"uid"`
	Month  string          `json:"month"` // format: 200601
	Id     string          `json:"id"`
	Status MonthBillStatus `json:"status"`
}

type Month_snapshotIn struct {
	Uid   uint32 `json:"uid"`
	Month string `json:"month"` // format: 200601
}

//用户模型
type NewUserModel struct {
	Uid    uint32 `json:"uid"`
	Excode string `json:"excode"`
	Desc   string `json:"desc"`
}

//增加合作伙伴
type PartnerAddIn struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	EffectTime string `json:"effect_time"` //2006-01-02
	Deadline   string `json:"deadline"`
	NickName   string `json:"nickname"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
}

//查找合作伙伴
type PartnerGetIn struct {
	Name string `json:"name"`
}

//合作伙伴列表
type PartnerListIn struct {
	Type   *string `json:"type"`
	Offset *int    `json:"offset"`
	Limit  *int    `json:"limit"`
}

type PartnerListOut struct {
	Name       string            `json:"name" bson:"name"`
	Type       string            `json:"type" bson:"type"`
	NickName   string            `json:"nickname" bson:"nickname"`
	Phone      string            `json:"phone" bson:"phone"`
	Email      string            `json:"email" bson:"email"`
	EffectTime HundredNanoSecond `json:"effect_time" bson:"effect_time"`
	Deadline   HundredNanoSecond `json:"deadline" bson:"deadline"`
	CreateAt   HundredNanoSecond `json:"create_at" bson:"create_at"`
	UpdateAt   HundredNanoSecond `json:"update_at" bson:"update_at"`
}

type PartnerReward struct {
	Id           string            `json:"id"`
	PartnerName  string            `json:"partner_name"`
	RewardType   string            `json:"reward_type"`
	RewardDetail interface{}       `json:"reward_detail"`
	Title        string            `json:"title"`
	Desc         string            `json:"desc"`
	EffectTime   string            `json:"effect_time"` // 2006-01-02
	Deadline     string            `json:"deadline"`    // 2006-01-02
	CreateAt     HundredNanoSecond `json:"create_at"`
	UpdateAt     HundredNanoSecond `json:"update_at"`
	IsAvaliable  bool              `json:"is_avaliable"`
}

type PartnerRewardAddIn struct {
	EffectTime   string      `json:"effect_time"`
	Title        string      `json:"title"`
	PartnerName  string      `json:"partner_name"`
	Deadline     string      `json:"deadline"`
	RewardType   string      `json:"reward_type"`
	RewardDetail interface{} `json:"reward_detail"`
	Desc         string      `json:"desc"`
}

type PartnerRewardAvailableIn struct {
	Id        string `json:"id"`
	Available bool   `json:"available"`
}

type PartnerRewardGet struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ThirdpartyLimit struct {
	Limit int `json:"limit"`
}

//充值
type RechargeIn struct {
	Excode string  `json:"excode"`
	Type   string  `json:"type"`
	Uid    uint32  `json:"uid"`
	Money  Money   `json:"money"`
	At     *Second `json:"at"` // 业务操作时间
	Desc   string  `json:"desc"`
}

//实时扣费记录
type RealtimeDeductIn struct {
	Uid   uint32 `json:"uid"`
	Month string `json:"month"` //200601
	Money Money  `json:"money"`
}

type RealtimeDeductDelete struct {
	Uid          uint32 `json:"uid"`
	Month        string `json:"month"` //200601
	DeleteBefore bool   `json:"delete_before"`
}

type RealtimeInfoOut struct {
	Amount    Money            `json:"amount"`     //coupon + cash
	Coupon    Money            `json:"coupon"`     //优惠券
	CostMoney Money            `json:"cost_money"` //用户充值的钱
	FreeMoney Money            `json:"free_money"` //用户赠送的NB
	Realtime  []RealtimeDetail `json:"realtime"`   //实时扣费金额
}

type RealtimeDetail struct {
	Month string `json:"month"` //200601
	Money Money  `json:"money"`
}

//获取充值列表
type RechargeListIn struct {
	Uid       *uint32            `json:"uid"`
	StartTime *HundredNanoSecond `json:"starttime"`
	EndTime   *HundredNanoSecond `json:"endtime"`
	Type      string             `json:"type"`
	IsHide    *bool              `json:"ishide"`
	Offset    *int64             `json:"offset"`
	Limit     *int64             `json:"limit"`
}

type RechargeMini struct {
	Excode  string  `json:"excode"`
	Uid     uint32  `json:"uid"`
	Money   int64   `json:"money"` // 0.0001yuan
	At      *Second `json:"at"`    // 业务操作时间
	Desc    string  `json:"desc"`
	Details string  `json:"details"`
}

//根据序列号获取优惠券
type RechargeRewardBy_serial_numIn struct {
	SerialNum string `json:"serial_num"`
}

//WsRechargeRewards
type RechargeTransaction struct {
	Serial_num string `json:"serail_num"`
	Type       string `json:"type"`
	Uid        uint32 `json:"uid"`
	Time       int64  `json:"time"`
	Money      int64  `json:"money"`
}

//获取充值优惠券
type RewardInfo struct {
	RechargeTran RechargeTransaction
	RewardTrans  []RewardTransaction
}

type RewardModel struct {
	Money  *int64 `json:"money"`
	Offset *int   `json:"offset"`
	Limit  *int   `json:"limit"`
}

type RewardTransaction struct {
	Uid               uint32 `json:"uid"`
	RechargeSerialNum string `json:"recharge_serial_num"`
	RewardSerialNum   string `json:"reward_serial_num"`
	Money             int64  `json"money"`
	Type              string `json:"type"`
	CreateAt          int64  `json:"create_at"`
}

//渠道分成充值
type Recharge_rewardIn struct {
	Excode           string  `json:"excode"`
	Type             string  `json:"type"`
	Desc             string  `json:"desc"`
	Uid              uint32  `json:"uid"`
	Money            Money   `json:"money"`
	Cost             Money   `json:"cost"`
	CustomerRewardId string  `json:"customer_reward_id"`
	At               *Second `json:"at"` // 业务操作时间
}

type Snapshot struct {
	Uid            uint32         `json:"uid"`
	Month          string         `json:"month"` // like: 2006-01
	Money          Money          `json:"money"`
	Cash           Money          `json:"cash"` // 0.0001
	CashEx         Money          `json:"cash_ex"`
	CashIn         Money          `json:"cash_in"`
	Coupon         Money          `json:"coupon"`
	CouponEx       Money          `json:"coupon_ex"`
	CouponIn       Money          `json:"coupon_in"`
	CouponExDetail []CouponChange `json:"coupon_ex_detail"`
	CouponInDetail []CouponChange `json:"coupon_in_detail"`
}

type TransactionAtReq struct {
	Serial_num string `json:"serial_num"`
	At         Second `json:"at"` // 业务操作时间
}

type TransactionOut struct {
	Serial_num    string         `json:"serial_num"`
	Excode        string         `json:"excode"`
	Prefix        string         `json:"prefix"`
	Type          string         `json:"type"`
	Uid           uint32         `json:"uid"`
	Time          int64          `json:"time"`
	Desc          string         `json:"desc"`
	Money         int64          `json:"money"`
	Cash          int64          `json:"cash"`
	Coupon        int64          `json:"coupon"`
	Details       string         `json:"details"`
	Cash_Detail   CashDetail     `json:"cash_detail"`
	Coupon_Detail []CouponDetail `json:"coupon_detail"`
	IsProcessed   bool           `json:"isprocessed"`
}

type User struct {
	Uid uint32 `json:"uid"`
}

// max uid len: MaxUidLen
type Users struct {
	Uids string `json:"uids"` // split by ",", eg: "11,22,33"
}

type UserListIn struct {
	UserType   *int `json:"user_type"`
	UserTag    *int `json:"user_tag"`
	DeductType *int `json:"deduct_type"`
	Offset     *int `json:"offset"`
	Limit      *int `json:"limit"`
}

type UserInfo struct {
	Uid        uint32            `json:"uid" bson:"uid"`
	Email      string            `json:"email" bson:"email"`
	Type       UserType          `json:"type" bson:"type"`
	EffectTime HundredNanoSecond `json:"effect_time" bson:"effect_time"`
	DeadTime   HundredNanoSecond `json:"dead_time" bson:"dead_time"`
	Tag        UserTag           `json:"tag" bson:"tag"`
	DeductType UserDeductType    `json:"deduct_type" bson:"deduct_type"`
	Company    string            `json:"company" bson:"company"`
	InformTO   []string          `json:"inform_to" bson:"inform_to"`
	InformCC   []string          `json:"inform_cc" bson:"inform_cc"`
	Remark     string            `json:"remark" bson:"remark"`
	NickName   string            `json:"nickname" bson:"nickname"`
}

type UserUpdateIn struct {
	Uid        uint32          `json:"uid" bson:"uid"`
	Company    *string         `json:"company" bson:"company"`
	Type       *UserType       `json:"type" bson:"type"`
	Tag        *UserTag        `json:"tag" bson:"tag"`
	DeductType *UserDeductType `json:"deduct_type" bson:"deduct_type"`
	Remark     *string         `json:"remark" bson:"remark"`
	NickName   *string         `json:"nickname" bson:"nickname"`
}

//销账
type WriteoffIn struct {
	Uid    uint32 `json:"uid"`
	Money  int64  `json:"money"`
	Prefix string `json:"prefix"` //RECHARGE, DEDUCT
	Desc   string `json:"desc"`
	Excode string `json:"excode"`
}
