package wallet

import (
	"encoding/json"
	"qbox.us/api/pay/price"
	"qbox.us/api/stat"
)

type Bill struct {
	Money         int64          `json:"money" bson:"money"`
	Cash          int64          `json:"cash" bson:"cash"`
	Coupon        int64          `json:"coupon" bson:"coupon"`
	Time          int64          `json:"time" bson:"time"`
	Serial_num    string         `json:"serial_num" bson:"serial_num"`
	Prefix        string         `json:"prefix" bson:"prefix"`
	Type          string         `json:"type" bson:"type"`
	Uid           uint32         `json:"uid" bson:"uid"`
	Desc          string         `json:"desc" bson:"desc"`
	Cash_Detail   CashDetail     `json:"cash_detail" bson:"cash_detail"`
	Coupon_Detail []CouponDetail `json:"coupon_detail" bson:"coupon_detail"`
}

type BillEx struct {
	Serial_num    string         `json:"serial_num" bson:"serial_num"`
	Prefix        string         `json:"prefix" bson:"prefix"`
	Type          string         `json:"type" bson:"type"`
	Uid           uint32         `json:"uid" bson:"uid"`
	Time          int64          `json:"time" bson:"time"`
	Desc          string         `json:"desc" bson:"desc"`
	Money         int64          `json:"money" bson:"money"`
	Cash          int64          `json:"cash" bson:"cash"`
	Coupon        int64          `json:"coupon" bson:"coupon"`
	Details       string         `json:"details" bson:"details"`
	Cash_Detail   CashDetail     `json:"cash_detail" bson:"cash_detail"`
	Coupon_Detail []CouponDetail `json:"coupon_detail" bson:"coupon_detail"`
}

///////////////////////////////////////////////////////////////////////////////////////////////////

type RangeBill struct {
	Range int64 `json:"range" bson:"range"`
	Value int64 `json:"value" bson:"value"`
	Money int64 `json:"money" bson:"money"`
}

type BaseUnit struct {
	Value  int64       `json:"value" bson:"value"`
	Money  int64       `json:"money" bson:"money"`
	Ranges []RangeBill `json:"ranges" bson:"ranges"`
}

type BaseBill struct {
	Price price.BasePrice `json:"price" bson:"price"`

	Space        BaseUnit `json:"space" bson:"space"`
	Transfer_out BaseUnit `json:"transfer_out" bson:"transfer_out"`
	Bandwidth    BaseUnit `json:"bandwidth" bson:"bandwidth"`
	Api_Get      BaseUnit `json:"api_get" bson:"api_get"`
	Api_Put      BaseUnit `json:"api_put" bson:"api_put"`
}

type DiscountBill struct {
	Price price.DiscountInfo `json:"price" bson:"price"`

	PreMoney  int64 `jons:"pre_money" bson:"pre_money"`
	PostMoney int64 `json:"post_money" bson:"post_money"`
}

type PackageBill struct {
	Price price.PackageInfo `json:"price" bson:"price"`

	Money        int64 `json:"money" bson:"money"`
	Space        int64 `json:"space" bson:"space"`
	Transfer_out int64 `json:"transfer_out" bson:"transfer_out"`
	Bandwidth    int64 `json:"bandwidth" bson:"bandwidth"`
	Api_Get      int64 `json:"api_get" bson:"api_get"`
	Api_Put      int64 `json:"api_put" bson:"api_put"`
}

type InfoUnit struct {
	Type string      `json:"type" bson:"type"` // base/discount/package
	Info interface{} `json:"info" bson:"info"`
}

type BillDetail struct {
	Uid  uint32 `json:"uid" bson:"uid"`
	Time int64  `json:"time" bson:"time"`

	Money int64         `json:"money" bson:"money"` // 0.0001 yuan
	Units []InfoUnit    `json:"units" bson:"units"`
	Data  stat.StatInfo `json:"data" bson:"data"`
}

//-----------------------------------------------------------------------------------------------//

type EncodingInfoUnit struct {
	Type string `json:"type" bson:"type"`
	Info string `json:"info" bson:"info"`
}

type EncodingBillDetail struct {
	Uid  uint32 `json:"uid" bson:"uid"`
	Time int64  `json:"time" bson:"time"`

	Money int64              `json:"money" bson:"money"`
	Units []EncodingInfoUnit `json:"units" bson:"units"`
	Data  stat.StatInfo      `json:"data" bson:"data"`
}

//-----------------------------------------------------------------------------------------------//

func MarshalBillDetail(bill BillDetail) (bill2 EncodingBillDetail, err error) {

	bill2.Uid = bill.Uid
	bill2.Time = bill.Time
	bill2.Money = bill.Money
	bill2.Data = bill.Data

	if bill.Units == nil {
		return
	}
	bill2.Units = make([]EncodingInfoUnit, len(bill.Units))
	for i, unit := range bill.Units {
		var bs []byte
		switch unit.Type {
		case price.TYPE_BASE:
			bs, err = json.Marshal(unit.Info.(BaseBill))
		case price.TYPE_DISCOUNT:
			bs, err = json.Marshal(unit.Info.(DiscountBill))
		case price.TYPE_PACKAGE:
			bs, err = json.Marshal(unit.Info.(PackageBill))
		}
		if err != nil {
			return
		}
		bill2.Units[i] = EncodingInfoUnit{unit.Type, string(bs)}
	}
	return
}

func UnmarshalBillDetail(bill EncodingBillDetail) (bill2 BillDetail, err error) {

	bill2.Uid = bill.Uid
	bill2.Time = bill.Time
	bill2.Money = bill.Money
	bill2.Data = bill.Data

	if bill.Units == nil {
		return
	}
	bill2.Units = make([]InfoUnit, len(bill.Units))
	for i, unit := range bill.Units {
		switch unit.Type {
		case price.TYPE_BASE:
			var bill_ BaseBill
			err = json.Unmarshal([]byte(unit.Info), &bill_)
			if err != nil {
				return
			}
			bill2.Units[i] = InfoUnit{price.TYPE_BASE, bill_}
		case price.TYPE_DISCOUNT:
			var bill_ DiscountBill
			err = json.Unmarshal([]byte(unit.Info), &bill_)
			if err != nil {
				return
			}
			bill2.Units[i] = InfoUnit{price.TYPE_DISCOUNT, bill_}
		case price.TYPE_PACKAGE:
			var bill_ PackageBill
			err = json.Unmarshal([]byte(unit.Info), &bill_)
			if err != nil {
				return
			}
			bill2.Units[i] = InfoUnit{price.TYPE_PACKAGE, bill_}
		}
	}
	return
}
