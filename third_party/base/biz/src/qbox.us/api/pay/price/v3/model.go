package v3

import (
	"strconv"
	"time"

	. "qbox.us/api/pay/pay"
	. "qbox.us/zone"
)

//----------------------------------------------------------------------------//

type Kind string

var (
	KIND_BASE     Kind = "base"     // 基础价格，阶梯计费
	KIND_REWARD   Kind = "reward"   // <废弃，扩展成package>赞助，免费额度
	KIND_PACKAGE  Kind = "package"  // 套餐，当费用为0时特例化成原来的reward
	KIND_DISCOUNT Kind = "discount" // 折扣
	KIND_REBATE   Kind = "rebate"   // 返现，现金免费额度
	KIND_RESPACK  Kind = "respack"  // 资源包
)

////////////////////////////////////////////////////////////////////////////////

type ItemBasePriceType string

var (
	ITEM_BASE_COMMON ItemBasePriceType = "COMMON"
	ITEM_BASE_VACANT ItemBasePriceType = "VACANT" // 指定价格类型为vacant，则相应业务不计费
)

type ItemDataType string // 取值方式

var (
	ITEM_DATA_BANDWIDTH  ItemDataType = "bandwidth"
	ITEM_DATA_CONCURRENT ItemDataType = "concurrent"
)

type ItemCountType string // 计费方式

var (
	ITEM_COUNT_BANDWIDTH_95P        ItemCountType = "bandwith:95%"            // 以带宽形式，第95%的峰值点
	ITEM_COUNT_BANDWIDTH_TOP4       ItemCountType = "bandwith:top4"           // 以带宽形式，第4日峰值
	ITEM_COUNT_BANDWIDTH_MONTH_TOP4 ItemCountType = "bandwith:month_top4"     // 以带宽形式，月第4峰值
	ITEM_COUNT_BANDWIDTH_TOP        ItemCountType = "bandwith:top"            // 以带宽形式，月峰值
	ITEM_COUNT_BANDWIDTH_DAILYAVG   ItemCountType = "bandwith:daily_avg"      // 以带宽形式，日峰均值
	ITEM_COUNT_BANDWIDTH_DAILY_95P  ItemCountType = "bandwidth:daily_avg_95%" // 以带宽形式，月计费，日95峰值均值

	ITEM_COUNT_BANDWIDTH_DAILY_TOP    ItemCountType = "bandwidth:daily:top"  // 以带宽形式，日计费，日峰值
	ITEM_COUNT_BANDWIDTH_DAILY_TOP95P ItemCountType = "bandwidth:daily:95%"  // 以带宽形式，日计费，日95峰值
	ITEM_COUNT_BANDWIDTH_DAILY_TOP4   ItemCountType = "bandwidth:daily:top4" // 以带宽形式，日计费，日Top4峰值
)

func (t ItemCountType) IsCountByDaily() bool {
	switch t {
	case ITEM_COUNT_BANDWIDTH_DAILY_TOP, ITEM_COUNT_BANDWIDTH_DAILY_TOP95P, ITEM_COUNT_BANDWIDTH_DAILY_TOP4:
		return true
	default:
		return false
	}
}

type CumulativeType int

const (
	CumulativeTypeMonth CumulativeType = iota
	CumulativeTypeDay
	CumulativeTypeHour
)

func (c CumulativeType) String() string {
	switch c {
	case CumulativeTypeMonth:
		return "按月累计"
	case CumulativeTypeDay:
		return "按日累计"
	case CumulativeTypeHour:
		return "按小时累计"
	default:
		return "未知累计周期"
	}
}

func (c CumulativeType) IsValid() bool {
	switch c {
	case CumulativeTypeMonth, CumulativeTypeDay, CumulativeTypeHour:
		return true
	default:
		return false
	}
}

type BillPeriodType int

const (
	BillPeriodMonth BillPeriodType = iota
	BillPeriodDay
	BillPeriodHour
)

var billPeriodStrings = map[BillPeriodType]string{
	BillPeriodMonth: "monthly",
	BillPeriodDay:   "daily",
	BillPeriodHour:  "hourly",
}

func (b BillPeriodType) String() string {
	if s, ok := billPeriodStrings[b]; ok {
		return s
	}
	return "unknown bill period"
}

func (b BillPeriodType) IsValid() bool {
	_, ok := billPeriodStrings[b]
	return ok
}

var billPeriodHumanStrings = map[BillPeriodType]string{
	BillPeriodMonth: "按月出账",
	BillPeriodDay:   "按日出账",
	BillPeriodHour:  "按小时出账",
}

func (b BillPeriodType) HumanString() string {
	if s, ok := billPeriodHumanStrings[b]; ok {
		return s
	}
	return "未知账单周期"
}

type RangePriceType string

var (
	UNITPRICE     RangePriceType = "UNITPRICE"     // 各阶梯单价
	TOP_UNITPRICE RangePriceType = "TOP_UNITPRICE" // 最高阶梯单价
	FIRST_BUYOUT  RangePriceType = "FIRST_BUYOUT"  // 第一阶梯保底，其余单价
	EACH_BUYOUT   RangePriceType = "EACH_BUYOUT"   // 各阶梯保底
	BUYOUT        RangePriceType = "BUYOUT"        // 逐层阶梯保底
	ONEOFF        RangePriceType = "ONEOFF"        // <废弃，原描述不准确> 一口价
)

func (r RangePriceType) Desc() string {
	switch r {
	case UNITPRICE:
		return "各阶梯单价"
	case TOP_UNITPRICE:
		return "最高阶梯单价"
	case FIRST_BUYOUT:
		return "第一阶梯保底"
	case EACH_BUYOUT:
		return "各阶梯保底"
	case BUYOUT:
		return "逐个阶梯保底"
	case ONEOFF:
		return "一口价"
	default:
		return "暂不支持" // NOT SUPPORT
	}
}

////////////////////////////////////////////////////////////////////////////////
type EvmPriceType string

var (
	EVM_NORMAL    EvmPriceType = "NORMAL"
	EVM_MONTH_FEE EvmPriceType = "MONTH_FEE"
	EVM_YEAR_FEE  EvmPriceType = "YEAR_FEE"
)

func (r EvmPriceType) Desc() string {
	switch r {
	case EVM_NORMAL:
		return "单价"
	case EVM_MONTH_FEE:
		return "包月价"
	case EVM_YEAR_FEE:
		return "包年价"
	default:
		return "暂不支持"
	}
}

type EvmPkgType string

var (
	EVM_PKG_TYPE_UNIQ EvmPkgType = "uniq"
)

////////////////////////////////////////////////////////////////////////////////

type Selected int

var (
	YES Selected = 1
	NO  Selected = -1
)

// 业务作用范围
type Scope struct {
	All      bool                 `bson:"all"`      // 作用于所有产品
	Products map[Product]Selected `bson:"products"` // 作用于某类产品
	Groups   map[Group]Selected   `bson:"groups"`   // 作用于某类业务
	Items    map[Item]Selected    `bson:"items"`    // 作用于某项业务
}

////////////////////////////////////////////////////////////////////////////////

// 有效期
type LifeCycle struct {
	EffectTime Day `json:"effect_time"`
	DeadTime   Day `json:"dead_time"`
}

// 绑定的信息
type BindingInfo struct {
	BindId     string            `json:"bind_id,omitempty"` // 绑定关系的唯一标示
	OP         string            `json:"op"`                // 绑定操作的唯一标示
	EffectTime Day               `json:"effect_time"`       // 绑定生效时间
	DeadTime   Day               `json:"dead_time"`         // 绑定结束时间
	At         HundredNanoSecond `json:"at"`                // 操作时间
}

type KindInfo struct {
	ID   string `json:"id"`   // 唯一标示
	Type string `json:"type"` // 类型
	Name string `json:"name"`
	Desc string `json:"desc"`
}

////////////////////////////////////////////////////////////////////////////////

type UserGroup int

const (
	UserGroupInvalid UserGroup = iota
	UserGroupPersonalUncertified
	UserGroupPersonalCertified
	UserGroupEnterpriseUncertified
	UserGroupEnterpriseCertified
)

func (g UserGroup) ToString() string {
	return strconv.Itoa(int(g))
}

////////////////////////////////////////////////////////////////////////////////

type DeductType int

const (
	DeductAny     DeductType = iota // 不限
	DeductMonthly                   // 只能月扣
)

type StatSrc int

const (
	StatSrcJedi StatSrc = iota + 1
	StatSrcDora
	StatSrcKirk
	StatSrcUfop
	StatSrcPili
	StatSrcColdStorage
	StatSrcAtar
	StatSrcKylin
	StatSrcPandora
	StatSrcUfop2
	StatSrcFusionOv
)

type StatMethod int

const (
	StatMethodAvg StatMethod = iota + 1
	StatMethodSum
	StatMethodBandwidth
	StatMethodMax
)

func (i ModelItemDef) GetZonePrice(z Zone) string {
	return i.Price[z.String()]
}

func (i ModelItemDef) IsAvailableInZone(z Zone) bool {
	if i.IsMultiZone {
		_, ok := i.Price[z.String()]
		return ok
	} else {
		return z == ZONE_NB
	}
}

func (i ModelItemDef) ExtraGet(dataType ItemDataType, key string) string {
	if dataDef, ok := i.DataDef[dataType]; ok {
		if val, ok := dataDef.Extra[key]; ok {
			return val
		}
	}

	return i.Extra[key]
}

////////////////////////////////////////////////////////////////////////////////

type CarryOverPolicy int

const (
	CarryOverToBindEnd   CarryOverPolicy = iota // 结转至绑定期结束，绑定期内按月分配
	NoCarryOver                                 // 无结转，绑定期内按月分配
	NoCarryOverToBindEnd                        // 无结转，绑定期内一次分配
)

////////////////////////////////////////////////////////////////////////////////

type ReqUid struct {
	Uid  uint32 `json:"uid"`
	Zone Zone   `json:"zone"`
}

type ReqUidAndTime struct {
	Uid     uint32  `json:"uid"`
	Zone    Zone    `json:"zone"`
	Product Product `json:"product"`
	FromSec Second  `json:"from_sec"`
	ToSec   Second  `json:"to_sec"`

	// Deprecated
	From Day `json:"from"`
	To   Day `json:"to"`
}

func (r *ReqUidAndTime) FromTime() (time.Time, error) {
	if r.FromSec > 0 {
		return r.FromSec.Time(), nil
	} else {
		return r.From.Time()
	}
}

func (r *ReqUidAndTime) ToTime() (time.Time, error) {
	if r.ToSec > 0 {
		return r.ToSec.Time(), nil
	} else {
		return r.To.Time()
	}
}

type ReqID struct {
	ID string `json:"id"`
}

type ReqList struct {
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Type   string `json:"type"`
}

type ReqListTime struct {
	Marker uint32            `json:"marker"`
	Zone   Zone              `json:"zone"`
	Limit  int               `json:"limit"`
	Kind   string            `json:"kind"`
	From   HundredNanoSecond `json:"from"`
	To     HundredNanoSecond `json:"to"`
}

type ReqWithZones struct {
	Products string `json:"products"`
	Zones    string `json:"zones"`
	When     *Day   `json:"when"`
}

type BaseWithZone struct {
	Zone Zone      `json:"zone"`
	Base ModelBase `json:"base"`
}

type ReqListUserPrice struct {
	Uid  uint32 `json:"uid"`
	When *Day   `json:"when"`
	Zone *Zone  `json:"zone"`
}
