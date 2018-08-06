package v2

import (
	"strconv"
)

import (
	adminacc "qbox.us/admin_api/account.v2"
)

const (
	DEFAULT_STD_REWARD_ID          = "normal"
	DEFAULT_EXP_REWARD_ID          = "normal-exp"
	DEFAULT_STD_SPACE_REWARD_ID    = "normal_space"
	DEFAULT_EXP_SPACE_REWARD_ID    = "normal-exp_space"
	DEFAULT_STD_TRANSFER_REWARD_ID = "normal_transfer_out"
	DEFAULT_EXP_TRANSFER_REWARD_ID = "normal-exp_transfer_out"
	DEFAULT_STD_APIGET_REWARD_ID   = "normal_api_get"
	DEFAULT_EXP_APIGET_REWARD_ID   = "normal-exp_api_get"
	DEFAULT_STD_APIPUT_REWARD_ID   = "normal_api_put"
	DEFAULT_EXP_APIPUT_REWARD_ID   = "normal-exp_api_put"
	DEFAULT_DEAD_TIME              = 1<<63 - 1
	DEFAULT_STD_BASEPRICE_ID       = "v2"
	DEFAULT_STD_SPACE_BP_ID        = "v2_space"
	DEFAULT_STD_TRANSFER_BP_ID     = "v2_transfer_out"
	DEFAULT_STD_APIGET_BP_ID       = "v2_api_get"
	DEFAULT_STD_APIPUT_BP_ID       = "v2_api_put"
	DEFAULT_FREE_REWARD_TYPE       = "FREE"
)

//----------------------------------------------------------------------------//

type BasePriceType string

const (
	BASEPRICE_TYPE_BASE BasePriceType = "BASE"
	BASEPRICE_TYPE_VIP  BasePriceType = "VIP"
)

func (b BasePriceType) ToString() string {
	return string(b)
}

type BasePriceRangeType string

const (
	UNITPRICE    BasePriceRangeType = "UNITPRICE"    // 各阶梯单价
	FIRST_BUYOUT BasePriceRangeType = "FIRST_BUYOUT" // 第一阶梯保底
	EACH_BUYOUT  BasePriceRangeType = "EACH_BUYOUT"  // 各阶梯保底，...
	BUYOUT       BasePriceRangeType = "BUYOUT"       // 逐个阶梯保底，....
	ONEOFF       BasePriceRangeType = "ONEOFF"       // 一口价
)

//----------------------------------------------------------------------------//

type CustomerGroup adminacc.CustomerGroup

func (c CustomerGroup) ToString() string {
	return strconv.Itoa(int(c))
}

//----------------------------------------------------------------------------//

type Field string

const (
	SPACE       Field = "space"
	TRANSFEROUT Field = "transfer_out"
	APIGET      Field = "api_get"
	APIPUT      Field = "api_put"
	BANDWIDTH   Field = "bandwidth"
	SERVICE     Field = "service"
	ALL         Field = "all" // 仅用于折扣绑定，不能按此field来拉取价格
)

func (f Field) ToString() string {
	return string(f)
}

var FIELDS = []Field{SPACE, TRANSFEROUT, APIGET, APIPUT, BANDWIDTH, SERVICE, ALL}

//----------------------------------------------------------------------------//

type VType string // 取值类型

const (
	VALUE_TYPE_DEFAULT       VType = ""
	VALUE_TYPE_TOP4TH        VType = "top4th"
	VALUE_TYPE_TOP95_PERCENT VType = "top95"
	VALUE_TYPE_DAY_AVERAGE   VType = "day_average"
)

func (v VType) ToString() string {
	return string(v)
}

func (v VType) Format() string {
	switch v {
	case VALUE_TYPE_TOP4TH:
		return "第四日峰值"
	case VALUE_TYPE_TOP95_PERCENT:
		return "95%峰值点"
	case VALUE_TYPE_DAY_AVERAGE:
		return "日峰值平均值"
	default:
		return ""
	}
}

////////////////////////////////////////////////////////////////////////////////

type BasePrice struct {
	Id      string               `json:"id"`
	Type    BasePriceType        `json:"type"`
	Desc    string               `json:"desc"`
	Details map[Field]FieldPrice `json:"details"`
}

type BasePriceId struct {
	Id string `json:"id"`
}

type BasePriceListReq struct {
	Type   BasePriceType `json:"type"`
	Offset int           `json:"offset"`
	Limit  int           `json:"limit"`
}

type Discount struct {
	Id         string         `json:"id"`
	Type       string         `json:"type"`
	EffectTime int64          `json:"effect_time"` // hundred nanosecond
	DeadTime   int64          `json:"dead_time"`   // hundred nanosecond
	Days       int            `json:"days"`
	Name       string         `json:"name"`
	Desc       string         `json:"desc"`
	Money      int64          `json:"money"`
	Percent    int            `json:"percent"`
	Details    map[Field]bool `json:"details"`
}

type DiscountDesc struct {
	Id   string `json:"id"`
	Desc string `json:"desc"`
}

type DiscountId struct {
	Id string `json:"id"`
}

type DiscountListReq struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type DiscountName struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type FieldPrice struct {
	Type   BasePriceRangeType `json:"type"`
	Desc   string             `json:"desc"`
	Ranges []RangePrice       `json:"ranges"`
}

type FieldPricePortal struct {
	Type  BasePriceRangeType `json:"type"`
	Desc  string             `json:"desc"`
	Units []UnitPrice        `json:"units"`
}

type FieldPriceFormated struct {
	Type  BasePriceRangeType   `json:"type"`
	Desc  string               `json:"desc"`
	Units []RangePriceFormated `json:"units"`
}

type PriceUsersCount struct {
	Id   string `json:"id"`   // default: "", id and type canot be empty both
	Type string `json:"type"` // default: ""
	Time string `json:"time"` // default: "", format "2006-01-02"
}

type PriceUsersList struct {
	Id     string `json:"id"`     // default: "", id and type canot be empty both
	Type   string `json:"type"`   // default: ""
	Time   string `json:"time"`   // default: "", format "2006-01-02"
	Offset int    `json:"offset"` // default: 0
	Limit  int    `json:"limit"`  // default: 0
}

type RangePrice struct {
	Range int64 `json:"range"`
	Price int64 `json:"price"`
}

type RangePriceFormated struct {
	Range string `json:"range"`
	Price string `json:"price"`
}

type Reward struct {
	Id         string          `json:"id"`
	EffectTime int64           `json:"effect_time"` // hundred nanosecond
	DeadTime   int64           `json:"dead_time"`   // hundred nanosecond
	Days       int             `json:"days"`
	Type       string          `json:"type"` // SPONSOR, REWARD, FREE
	Name       string          `json:"name"`
	Desc       string          `json:"desc"`
	Details    map[Field]int64 `json:"details"`
}

type RewardDesc struct {
	Id   string `json:"id"`
	Desc string `json:"desc"`
}

type RewardId struct {
	Id string `json:"id"`
}

type RewardListReq struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type RewardName struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type UnitPrice struct {
	From  int64 `json:"from"`  // space,transer:GB; apt_get,api_put:K; bandwidith:Mb
	To    int64 `json:"to"`    // space,transer:GB; apt_get,api_put:K; bandwidith:Mb
	Price int64 `json:"price"` // 0.0001 yuan
}

type UserBasePriceTimeRangeUpdater struct {
	Uid           uint32        `json:"uid"`
	CustomerGroup CustomerGroup `json:"customer_group"`
	Id            string        `json:"id"`
	EffectTime    int64         `json:"effect_time"` // hundred nanosecond, for user
	DeadTime      int64         `json:"dead_time"`   // hundred nanosecond, for user
	OpId          string        `json:"op_id"`
}

type UserCustomgroupUpdater struct {
	Uid           uint32        `json:"uid"`
	CustomerGroup CustomerGroup `json:"customer_group"`
	Time          int64         `json:"time"` // 对齐指每天的零点
}

// 指定业务的时间区间内的唯一价格
type UserPriceInFieldAndTimeRange struct {
	ValueType  UserValueType      `json:"value_type"`
	BasePrice  UserFieldBasePrice `json:"base_price"`
	Rewards    []UserFieldReward  `json:"rewards"`
	Discounts  []UserDiscount     `json:"discounts"`
	EffectTime int64              `json:"effect_time"`
	DeadTime   int64              `json:"dead_time"`
}

// 指定时间段内的价格，按业务分割
type UserPriceWithTimeRangeByFiled struct {
	Uid    uint32                                   `json:"uid"`
	Prices map[Field][]UserPriceInFieldAndTimeRange `json:"prices"`
}

type UserPriceWithRangeTimeGetter struct {
	Uid           uint32        `json:"uid"`
	CustomerGroup CustomerGroup `json:"customer_group"`
	Start         int64         `json:"start"`
	End           int64         `json:"end"`
}

// 某点时间的价格，旧
type UserPrice struct {
	Uid        uint32                  `json:"uid"`
	ValueTypes map[Field]UserValueType `json:"value_types"`
	BasePrice  UserBasePrice           `json:"base_price"`
	Rewards    []UserReward            `json:"rewards"`
	Discounts  []UserDiscount          `json:"discounts"`
}

type UserPriceFormated struct {
	Uid        uint32                          `json:"uid"`
	ValueTypes map[Field]UserValueTypeFormated `json:"value_types"`
	BasePrice  UserBasePriceFormated           `json:"base_price"`
	Rewards    []UserRewardFormated            `json:"rewards"`
	Discounts  []UserDiscountFormated          `json:"discounts"`
}

type UserPricePortal struct {
	Uid        uint32                  `json:"uid"`
	ValueTypes map[Field]UserValueType `json:"value_types"`
	BasePrice  UserBasePricePortal     `json:"base_price"`
	Rewards    UserRewards             `json:"rewards"`
	Discounts  []UserDiscount          `json:"discounts"`
}

type UserPriceGetter struct {
	Uid           uint32        `json:"uid"`
	CustomerGroup CustomerGroup `json:"customer_group"`
}

type UserPriceId struct {
	Uid  uint32 `json:"uid"`
	Id   string `json:"id"`
	OpId string `json:"op_id"`
}

type UserPriceTimeRange struct {
	Uid        uint32 `json:"uid"`
	Id         string `json:"id"`
	EffectTime int64  `json:"effect_time"` // hundred nanosecond, for user
	DeadTime   int64  `json:"dead_time"`   // hundred nanosecond, for user
	OpId       string `json:"op_id"`
}

type UserPriceWithTimeGetter struct {
	Uid           uint32        `json:"uid"`
	CustomerGroup CustomerGroup `json:"customer_group"`
	Time          int64         `json:"time"`
}

type UserAllPrice struct {
	Uid        uint32          `json:"uid"`
	ValueTypes []UserValueType `json:"value_types"`
	BasePrices []UserBasePrice `json:"base_prices"`
	Rewards    []UserReward    `json:"rewards"`
	Discounts  []UserDiscount  `json:"discounts"`
}

type UserAllPriceFormated struct {
	Uid        uint32                  `json:"uid"`
	ValueTypes []UserValueTypeFormated `json:"value_types"`
	BasePrices []UserBasePriceFormated `json:"base_prices"`
	Rewards    []UserRewardFormated    `json:"rewards"`
	Discounts  []UserDiscountFormated  `json:"discounts"`
}

type UserBasePrice struct {
	Id         string               `json:"id"`
	OpId       string               `json:"op_id"`
	Type       BasePriceType        `json:"type"`
	Desc       string               `json:"desc"`
	Details    map[Field]FieldPrice `json:"details"`
	EffectTime int64                `json:"effect_time"` // hundred nanosecond, for user
	DeadTime   int64                `json:"dead_time"`   // hundred nanosecond, for user
	ActiveTime int64                `json:"active_time"` // hundred nanosecond, for user
}

type UserBasePricePortal struct {
	Id         string                     `json:"id"`
	Type       BasePriceType              `json:"type"`
	Desc       string                     `json:"desc"`
	Details    map[Field]FieldPricePortal `json:"details"`
	EffectTime int64                      `json:"effect_time"` // hundred nanosecond, for user
	DeadTime   int64                      `json:"dead_time"`   // hundred nanosecond, for user
	ActiveTime int64                      `json:"active_time"` // hundred nanosecond, for user
}

type UserBasePriceFormated struct {
	Id         string                       `json:"id"`
	Type       BasePriceType                `json:"type"`
	Desc       string                       `json:"desc"`
	EffectTime string                       `json:"effect_time"`
	DeadTime   string                       `json:"dead_time"`
	ActiveTime string                       `json:"active_time"`
	Details    map[Field]FieldPriceFormated `json:"details"`
}

type UserBasePriceSetter struct {
	Uid           uint32               `json:"uid"`
	CustomerGroup CustomerGroup        `json:"customer_group"`
	Id            string               `json:"id"`
	Type          BasePriceType        `json:"type"`
	Desc          string               `json:"desc"`
	Details       map[Field]FieldPrice `json:"details"`
	EffectTime    int64                `json:"effect_time"` // hundred nanosecond, for user
	DeadTime      int64                `json:"dead_time"`   // hundred nanosecond, for user
}

type UserFieldBasePrice struct {
	Id         string        `json:"id"`
	Field      Field         `json:"field"`
	Type       BasePriceType `json:"type"`
	Desc       string        `json:"desc"`
	Price      FieldPrice    `json:"price"`
	EffectTime int64         `json:"effect_time"` // hundred nanosecond, for user
	DeadTime   int64         `json:"dead_time"`   // hundred nanosecond, for user
	ActiveTime int64         `json:"active_time"` // hundred nanosecond, for user
}

type UserDiscount struct {
	Id         string         `json:"id"`
	OpId       string         `json:"op_id"`
	Type       string         `json:"type"`
	EffectTime int64          `json:"effect_time"` // hundred nanosecond, for user
	DeadTime   int64          `json:"dead_time"`   // hundred nanosecond, for user
	ActiveTime int64          `json:"active_time"` // hundred nanosecond, for user
	Name       string         `json:"name"`
	Desc       string         `json:"desc"`
	Money      int64          `json:"money"`
	Percent    int            `json:"percent"`
	Details    map[Field]bool `json:"details"`
}

type UserDiscountFormated struct {
	Id         string `json:"id"`
	Type       string `json:"type"`
	EffectTime string `json:"effect_time"`
	DeadTime   string `json:"dead_time"`
	ActiveTime string `json:"active_time"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Details    string `json:"details"`
}

type UserDiscountSetter struct {
	Uid           uint32        `json:"uid"`
	CustomerGroup CustomerGroup `json:"customer_group"`
	Id            string        `json:"id"`
	EffectTime    int64         `json:"effect_time"` // hundred nanosecond, for user
	DeadTime      int64         `json:"dead_time"`   // hundred nanosecond, for user
}

type UserFieldReward struct {
	Id         string `json:"id"`
	OpId       string `json:"op_id"`
	Field      Field  `json:"field"` // TODO space,transer:GB; apt_get,api_put:K; bandwidith:Mb
	Type       string `json:"type"`
	EffectTime int64  `json:"effect_time"` // hundred nanosecond, for user
	DeadTime   int64  `json:"dead_time"`   // hundred nanosecond, for user
	ActiveTime int64  `json:"active_time"` // hundred nanosecond, for user
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Value      int64  `json:"value"`
}

type UserListItem struct {
	Uid        uint32 `json:"uid" bson:"uid"`
	EffectTime int64  `json:"effect_time" bson:"effect_time"` // hundred nanosecond
	DeadTime   int64  `json:"dead_time" bson:"dead_time"`     // hundred nanosecond
}

type UserReward struct {
	Id         string          `json:"id"`
	OpId       string          `json:"op_id"`
	Type       string          `json:"type"`
	EffectTime int64           `json:"effect_time"` // hundred nanosecond, for user
	DeadTime   int64           `json:"dead_time"`   // hundred nanosecond, for user
	ActiveTime int64           `json:"active_time"` // hundred nanosecond, for user
	Name       string          `json:"name"`
	Desc       string          `json:"desc"`
	Details    map[Field]int64 `json:"details"` // space,transer:GB; apt_get,api_put:K; bandwidith:Mb
}

type UserRewardFormated struct {
	Id         string `json:"id"`
	Type       string `json:"type"`
	EffectTime string `json:"effect_time"`
	DeadTime   string `json:"dead_time"`
	ActiveTime string `json:"active_time"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Details    string `json:"details"`
}

type UserRewards struct {
	Total   map[Field]int64 `json:"total_reward"` // space,transer:GB; apt_get,api_put:K; bandwidith:Mb
	Rewards []UserReward    `json:"rewards"`
}

type UserRewardSetter struct {
	Uid           uint32        `json:"uid"`
	CustomerGroup CustomerGroup `json:"customer_group"`
	Id            string        `json:"id"`
	EffectTime    int64         `json:"effect_time"` // hundred nanosecond, for user
	DeadTime      int64         `json:"dead_time"`   // hundred nanosecond, for user
}

type UserValueType struct {
	Id         string `json:"id"`
	OpId       string `json:"op_id"`
	EffectTime int64  `json:"effect_time"` // hundred nanosecond, for user
	DeadTime   int64  `json:"dead_time"`   // hundred nanosecond, for user
	ActiveTime int64  `json:"active_time"` // hundred nanosecond, for user
	Field      Field  `json:"field"`
	Type       VType  `json:"type"`
	Desc       string `json:"desc"`
}

type UserValueTypeFormated struct {
	Desc       string `json:"desc"`
	EffectTime string `json:"effect_time"`
	DeadTime   string `json:"dead_time"`
}

type UserValueTypeSetter struct {
	Uid           uint32        `json:"uid"`
	CustomerGroup CustomerGroup `json:"customer_group"`
	Id            string        `json:"id"`
	EffectTime    int64         `json:"effect_time"` // hundred nanosecond, for user
	DeadTime      int64         `json:"dead_time"`   // hundred nanosecond, for user
}

type ValueType struct {
	Id    string `json:"id" bson:"id"`
	Field Field  `json:"field" bson:"field"` // space, transfer_out, api_get, api_put, bandwidth
	Type  VType  `json:"type" bson:"type"`   // defult, top4th, top95%, day_average
	Desc  string `json:"desc" bson:"desc"`
}

type ValueTypeId struct {
	Id string `json:"id"`
}

type ValueTypeListReq struct {
	Field  Field `json:"field"`
	Type   VType `json:"type"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
}

type ValueTypeDesc struct {
	Id   string `json:"id"`
	Desc string `json:"desc"`
}
