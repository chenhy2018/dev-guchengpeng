package wallet

import (
	"encoding/json"
)

import (
	. "qbox.us/api/pay/pay"
	P "qbox.us/api/pay/price/v3"
	. "qbox.us/zone"
)

////////////////////////////////////////////////////////////////////////////////

const (
	BASEBILL_VERSION_CURR       string = "v2"
	MONTHSTATEMENT_VERSION_CURR string = "v2"
)

// 月对账单
type MonthStatementStatus int

const (
	MONTH_STATEMENT_STATUS_NOT_OUT   MonthStatementStatus = 1
	MONTH_STATEMENT_STATUS_CHARGEOFF MonthStatementStatus = 2 //月对账单出帐
	MONTH_STATEMENT_STATUS_PAID      MonthStatementStatus = 3 //月对账单已支付
)

//----------------------------------------------------------------------------//

type BillType int

const (
	BILLTYPE_NIL      BillType = 0
	BILLTYPE_BASEBILL BillType = 1
)

////////////////////////////////////////////////////////////////////////////////
type ReqID struct {
	ID string `json:"id"`
}

type ReqUid struct {
	Uid uint32 `json:"uid"`
}

type ReqUidAndRange struct {
	Uid     uint32  `json:"uid"`
	Product Product `json:"product"`
	Zone    *Zone   `json:"zone"`
	From    Day     `json:"from"`
	To      Day     `json:"to"`
}

type ReqUidAndRangeSecond struct {
	Uid     uint32  `json:"uid"`
	Product Product `json:"product"`
	Zone    *Zone   `json:"zone"`
	From    Second  `json:"from"`
	To      Second  `json:"to"`
}

type ReqUidAndSecond struct {
	Uid  uint32 `json:"uid"`
	Zone *Zone  `json:"zone"`
	Time Second `json:"time"`
}

type ReqListMonthBillCost struct {
	Uid  uint32 `json:"uid"`
	Zone Zone   `json:"zone"`
	// Format: '201603'
	Month string `json:"month"`
}

type RespUserCost map[string]int64

////////////////////////////////////////////////////////////////////////////////

const (
	T_PREFIX_SYS      = "SYS"
	T_PREFIX_DEDUCT   = "DEDUCT"
	T_PREFIX_RECHARGE = "RECHARGE"
	T_PREFIX_COUPON   = "COUPON"
)

const (
	T_ALIPAY           = "alipay"
	T_BANK             = "BANK"
	T_NEWUSER          = "newuser"
	T_COUPON_ACTIVE    = "active"
	T_COUPON_DISCARD   = "discard"
	T_REWARD           = "reward"  // recharge
	T_PRESENT          = "PRESENT" // recharge freenb default
	T_PREPAID_CARD     = "prepaid_card"
	T_WRITEOFF         = "writeoff"
	T_STORAGE_DAY      = "storage-day"
	T_STORAGE_MONTH    = "storage-month"
	T_THIRDPARTY       = "thirdparty"
	T_BASEBILL         = "basebill"
	T_CUSTOMBILL       = "custombill"
	T_RTBILL           = "rtbill"
	T_REFUND           = "refund"
	T_AGENCY           = "agency"
	T_TRANSFER_ACC_IN  = "transferaccin"
	T_TRANSFER_ACC_OUT = "transferaccout"
)

const (
	BASEBILL_IDENT = "basebill" //标识帐单业务类型,方便后续查询使用,当前只有basebill
)

const (
	DefaultThirdpartyLimit = 1000 * 10000 //￥1000
)

//////////////////////////// Resource Group ////////////////////////////////////

func getResourceGroup(resGrpList *P.ModelResourceGroupList, resGrpIdx int) interface{} {
	if resGrpIdx == 0 {
		// Default resource group
		return nil
	} else {
		// Custom resource group
		group, err := resGrpList.UnmarshalGroup(resGrpIdx - 1)
		if err == nil {
			return group
		} else {
			return nil
		}

	}
}

func getItemPrice(price *P.ModelItemBasePrice, resGrpIdx int) P.ModelItemBasePrice {
	// ResGrpIdx == 0 means default resource group, so list index is `ResGrpIdx - 1`
	idx := resGrpIdx - 1
	if idx >= 0 && idx < len(price.ResourceGroupList.Groups) {
		resGrp := price.ResourceGroupList.Groups[idx]

		var meta P.ResourceGroupMeta
		err := json.Unmarshal(resGrp, &meta)

		if err == nil {
			return meta.Price
		}
	}

	return *price
}

////////////////////////////////////////////////////////////////////////////////

type ModelInfo struct {
	Amount    Money `json:"amount"`     //amount = cash
	Cash      Money `json:"cash"`       //cash = CostMoney + FreeMoney
	Coupon    Money `json:"coupon"`     //优惠券
	CostMoney Money `json:"cost_money"` //用户充值的钱
	FreeMoney Money `json:"free_money"` //用户赠送的NB
}

type ModelTransactionV4 struct {
	Serial_num  string            `json:"serial_num"`
	Excode      string            `json:"excode"`
	Prefix      string            `json:"prefix"`
	Type        string            `json:"type"`
	Uid         uint32            `json:"uid"`
	Time        HundredNanoSecond `json:"time"`
	Desc        string            `json:"desc"`
	Money       Money             `json:"money"`
	Cash        Money             `json:"cash"`
	FreeNb      Money             `json:"freenb"`
	Coupon      Money             `json:"coupon"`
	Left        Money             `json:"left"`
	Details     string            `json:"details"`
	IsProcessed bool              `json:"isprocessed"`
	PayDetails  []PayTransaction  `json:"pay_details"`
}

type PayTool string

const (
	PAY_TOOL_CASH   PayTool = "cash"   // 现金
	PAY_TOOL_NB     PayTool = "nb"     // 牛币
	PAY_TOOL_COUPON PayTool = "coupon" // 抵用券
)

type PayTransaction struct {
	ID        string            `json:"id"`
	PayUid    uint32            `json:"pay_uid"`
	PayAt     HundredNanoSecond `json:"pay_at"`
	PayTool   PayTool           `json:"pay_tool"`
	PayToolID string            `json:"pay_tool_id"` // only for coupon
	Before    Money             `json:"before"`
	Change    Money             `json:"change"`
	After     Money             `json:"after"`
	Left      Money             `json:"left"` // transaction left money
}

type ModelTransaction struct {
	Serial_num    string              `json:"serial_num"`
	Excode        string              `json:"excode"`
	Prefix        string              `json:"prefix"`
	Type          string              `json:"type"`
	Uid           uint32              `json:"uid"`
	Time          HundredNanoSecond   `json:"time"`
	Desc          string              `json:"desc"`
	Money         Money               `json:"money"`
	Cash          Money               `json:"cash"`
	FreeNb        Money               `json:"freenb"`
	Coupon        Money               `json:"coupon"`
	Details       string              `json:"details"`
	Cash_Detail   ModelCashDetail     `json:"cash_detail"`
	FreeNb_Detail ModelCashDetail     `json:"freenb_detail"`
	Coupon_Detail []ModelCouponDetail `json:"coupon_detail"`
	IsProcessed   bool                `json:"isprocessed"`
	PayUid        uint32              `json:"pay_uid"`
	PayAt         HundredNanoSecond   `json:"pay_at"`
	RelatedMonth  Second              `json:"related_month"`
}

type ModelCashDetail struct {
	Before Money `json:"before"`
	Change Money `json:"change"`
	After  Money `json:"after"`
}

type ModelCouponDetail struct {
	Id     string `json:"id"`
	Before Money  `json:"before"`
	Change Money  `json:"change"`
	After  Money  `json:"after"`
}

type EvmBasebillBaseType string

const (
	EVM_BASEBILL_BASE      EvmBasebillBaseType = "base"
	EVM_BASEBILL_UPGRADE   EvmBasebillBaseType = "upgrade"
	EVM_BASEBILL_DOWNGRADE EvmBasebillBaseType = "downgrade"
	EVM_BASEBILL_RENEW     EvmBasebillBaseType = "renew"
)

type Selected int

var (
	YES Selected = 1
	NO  Selected = -1
)

type Scope struct {
	All      bool                 `json:"all"`
	Products map[Product]Selected `json:"products"`
	Groups   map[Group]Selected   `json:"groups"`
	Items    map[Item]Selected    `json:"items"`
}

func (r Scope) isEmpty() bool {
	if r.All == false && len(r.Products) == 0 &&
		len(r.Groups) == 0 && len(r.Items) == 0 {
		return true
	}
	return false
}

func (r Scope) ToString() string {
	bytes, _ := json.Marshal(&r)
	return string(bytes)
}

func (r Scope) Check(product Product, group Group, item Item) bool {
	// empty means all
	if r.isEmpty() {
		return true
	}

	if item != "" {
		if r.Items[item] == NO || r.Items[item] == YES {
			return r.Items[item] == YES
		}
		group = item.Group()
	}

	if group != "" {
		if r.Groups[group] == NO || r.Groups[group] == YES {
			return r.Groups[group] == YES
		}
		product = group.Product()
	}

	if product != "" {
		if r.Products[product] == YES || r.Products[product] == NO {
			return r.Products[product] == YES
		}
	}

	if r.All {
		return true
	}

	return false
}

////////////////////// Coupon //////////////////////////////////////////
type CouponType string

func (r *CouponType) ToString() string {
	return string(*r)
}

var (
	COUPON_TYPE_NEWUSER  CouponType = "NEWUSER"
	COUPON_TYPE_RECHARGE CouponType = "RECHARGE"
)

type CouponStatus int

var (
	COUPON_STATUS_IGNORE  CouponStatus = 0
	COUPON_STATUS_NEW     CouponStatus = 1
	COUPON_STATUS_ACTIVE  CouponStatus = 2
	COUPON_STATUS_OVERDUE CouponStatus = 3
)
