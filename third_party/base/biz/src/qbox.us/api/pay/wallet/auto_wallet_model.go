package wallet

//----------------------------------------------------------------------------//

type Money int64

func (m Money) AsYuan() float64 {
	return float64(m) / 10000
}

//----------------------------------------------------------------------------//

type CashDetail struct {
	Before int64 `json:"before" bson:"before"`
	Change int64 `json:"change" bson:"change"`
	After  int64 `json:"after" bson:"after"`
}

type CouponDetail struct {
	Id     string `json:"id" bson:"id"`
	Before int64  `json:"before" bson:"before"`
	Change int64  `json:"change" bson:"change"`
	After  int64  `json:"after" bson:"after"`
}

//隐藏帐单
type HideBillIn struct {
	Id     string `json:"id"`
	IsHide bool   `json:"ishide"`
}

//用户模型
type NewUserModel struct {
	Serial_num string `json:"serial_num"`
	Uid        uint32 `json:"uid"`
	Excode     string `json:"excode"`
	Desc       string `json:"desc"`
}

type RechargeMini struct {
	Excode  string `json:"excode"`
	Uid     uint32 `json:"uid"`
	Money   int64  `json:"money"` // 0.0001yuan
	Desc    string `json:"desc"`
	Details string `json:"details"`
}

type RechargeModel struct {
	Uid       uint32 `json:"uid"`
	StartTime string `json:"starttime"`
	EndTime   string `json:"endtime"`
	Type      string `json:"type_"`
	Offset    int64  `json:"offset"`
	Limit     int64  `json:"limit"`
}

type RechargeModelOut struct {
	Serial_num    string         `json:"serial_num"`
	Excode        string         `json:"excode"`
	Excode2       string         `json:"-"`
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

//WsRechargeRewards
type RechargeTransaction struct {
	Serial_num string
	Type       string
	Uid        uint32
	Time       int64
	Money      int64
}

type RewardInfo struct {
	RechargeTran RechargeTransaction
	RewardTrans  []RewardTransaction
}

type RewardModel struct {
	Money  int64 `json:"money"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
}

type RewardTransaction struct {
	Uid               uint32 `json:"uid"`
	RechargeSerialNum string `json:"recharge_serial_num"`
	RewardSerialNum   string `json:"reward_serial_num"`
	Money             int64  `json"money"`
	Type              string `json:"type"`
	CreateAt          int64  `json:"create_at"`
}

//销账
type WriteoffIn struct {
	Uid    uint32 `json:"uid"`
	Money  int64  `json:"money"`
	Prefix string `json:"prefix"` //RECHARGE, DEDUCT
	Desc   string `json:"desc"`
	Excode string `json:"excode"`
}
