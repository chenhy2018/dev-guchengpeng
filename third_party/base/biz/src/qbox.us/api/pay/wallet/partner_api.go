package wallet

import (
	"net/url"
	"qbox.us/rpc"
	"strconv"
)

type PartnerSetter struct {
	Name       string `json:"name" bson:"name"`
	Type       string `json:"type" bson:"type"`
	NickName   string `json:"nickname" bson:"nickname"`
	Phone      string `json:"phone" bson:"phone"`
	Email      string `json:"email" bson:"email"`
	EffectTime string `json:"effect_time" bson:"effect_time"` // 2006-01-02
	Deadline   string `json:"deadline" bson:"deadline"`       // 2006-01-02
}

type PartnerGetter struct {
	Name       string            `json:"name" bson:"name"`
	Type       string            `json:"type" bson:"type"`
	NickName   string            `json:"nickname" bson:"nickname"`
	Phone      string            `json:"phone" bson:"phone"`
	Email      string            `json:"email" bson:"email"`
	EffectTime string            `json:"effect_time" bson:"effect_time"` // 2006-01-02
	Deadline   string            `json:"deadline" bson:"deadline"`       // 2006-01-02
	CreateAt   HundredNanoSecond `json:"create_at" bson:"create_at"`
	UpdateAt   HundredNanoSecond `json:"update_at" bson:"update_at"`
}

// -----------------------------------------------------------------
type RewardCategory string

const (
	Discount      RewardCategory = "DiscountReward"      //折扣
	SpecificPrice RewardCategory = "SpecificPriceReward" //指定价格
	RangePrice    RewardCategory = "RangePriceReward"    //阶梯价格
)

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

// -----------------------------------------------------------------

type PartnerRewardSetter struct {
	PartnerName  string      `json:"partner_name" bson:"partner_name"`
	RewardType   string      `json:"reward_type" bson:"reward_type"`
	RewardDetail interface{} `json:"reward_detail" bson:"reward_detail"`
	Title        string      `json:"title" bson:"title"`
	Desc         string      `json:"desc" bson:"desc"`
	EffectTime   string      `json:"effect_time" bson:"effect_time"` // 2006-01-02
	Deadline     string      `json:"deadline" bson:"deadline"`       // 2006-01-02
}

type PartnerReward struct {
	Id           string            `json:"id" bson:"_id"`
	PartnerName  string            `json:"partner_name" bson:"partner_name"`
	RewardType   string            `json:"reward_type" bson:"reward_type"`
	RewardDetail interface{}       `json:"reward_detail" bson:"reward_detail"`
	Title        string            `json:"title" bson:"title"`
	Desc         string            `json:"desc" bson:"desc"`
	EffectTime   string            `json:"effect_time" bson:"effect_time"` // 2006-01-02
	Deadline     string            `json:"deadline" bson:"deadline"`       // 2006-01-02
	CreateAt     HundredNanoSecond `json:"create_at" bson:"create_at"`
	UpdateAt     HundredNanoSecond `json:"update_at" bson"update_at"`
	IsAvaliable  bool              `json:"is_avaliable" bson:"is_avaliable"`
}

type DiscountTransaction struct {
	Serial_num  string            `json:"serial_num"`
	Time        HundredNanoSecond `json:"time"` // 100 ns
	PartnerName string            `json:"partner_name"`
	CustomerID  string            `json:"customer_id"`
	PartnerID   string            `json:"partner_id"`
	Money       int64             `json:"money"`
	Pay         int64             `json:"pay"`
	Cost        int64             `json:"cost"`
}

type DiscountTransactionSummary struct {
	Count int   `json:"count"`
	Money int64 `json:"money"`
	Pay   int64 `json:"pay"`
	Cost  int64 `json:"cost"`
}

//	Args:
//	type，可选
//	offset，可选
//	limit，可选
func ListPartners(c rpc.Client, host string, args map[string]string) (
	partners []PartnerGetter, code int, err error) {
	values := url.Values{}
	for key, value := range args {
		values.Add(key, value)
	}
	code, err = c.Call(&partners, host+"/partner/list?"+values.Encode())
	return
}

func GetPartner(c rpc.Client, host string, name string) (
	partner PartnerGetter, code int, err error) {
	values := url.Values{}
	values.Add("name", name)
	code, err = c.Call(&partner, host+"/partner/get?"+values.Encode())
	return
}

func AddPartner(c rpc.Client, host string, partner PartnerSetter) (code int, err error) {
	code, err = c.CallWithJson(nil, host+"/partner/add", partner)
	return
}

func UpdatePartner(c rpc.Client, host string, partner PartnerSetter) (code int, err error) {
	code, err = c.CallWithJson(nil, host+"/partner/update", partner)
	return
}

func AddPartnerReward(c rpc.Client, host, partnerName string, rewardType RewardCategory, title, desc,
	effectTime, deadline string, reward interface{}) (code int, err error) {
	pr := PartnerRewardSetter{
		PartnerName:  partnerName,
		RewardType:   string(rewardType),
		Title:        title,
		Desc:         desc,
		EffectTime:   effectTime,
		Deadline:     deadline,
		RewardDetail: reward,
	}
	code, err = c.CallWithJson(nil, host+"/partner/reward/add", pr)
	return
}

func GetPartnerReward(c rpc.Client, host string, id, name string) (
	pr PartnerReward, code int, err error) {
	values := url.Values{}
	if id != "" {
		values.Add("id", id)
	} else {
		values.Add("name", name)
	}

	code, err = c.Call(&pr, host+"/partner/reward/get?"+values.Encode())
	return
}

func SetPartnerRewardAvailable(c rpc.Client, host string, id string, avail bool) (code int, err error) {
	values := url.Values{}
	values.Add("id", id)
	values.Add("available", strconv.FormatBool(avail))
	code, err = c.CallWithForm(nil, host+"/partner/reward/available", values)
	return
}

func GetDiscountBill(c rpc.Client, host string, serialNum string) (
	dt DiscountTransaction, code int, err error) {
	values := url.Values{}
	values.Add("serial_num", serialNum)

	code, err = c.Call(&dt, host+"/bill/discount/get?"+values.Encode())
	return
}

func ListDiscountBill(c rpc.Client, host string, rewardId, partnerName, from, to string, offset, limit int) (
	dts []DiscountTransaction, code int, err error) {
	values := url.Values{}
	if rewardId != "" {
		values.Add("reward_id", rewardId)
	}
	if partnerName != "" {
		values.Add("partner_name", partnerName)
	}
	if from != "" {
		values.Add("from", from)
	}
	if to != "" {
		values.Add("to", to)
	}
	if offset != 0 {
		values.Add("offset", strconv.Itoa(offset))
	}
	if limit != 0 {
		values.Add("limit", strconv.Itoa(limit))
	}

	code, err = c.Call(&dts, host+"/bill/discount/list?"+values.Encode())
	return
}

func DiscountBillSummary(c rpc.Client, host string, rewardId, partnerName, from, to string) (
	ds DiscountTransactionSummary, code int, err error) {
	values := url.Values{}
	if rewardId != "" {
		values.Add("reward_id", rewardId)
	}
	if partnerName != "" {
		values.Add("partner_name", partnerName)
	}
	if from != "" {
		values.Add("from", from)
	}
	if to != "" {
		values.Add("to", to)
	}

	code, err = c.Call(&ds, host+"/bill/discount/sum?"+values.Encode())
	return
}

func (r ServiceEx) ListPartners(host string, args map[string]string) (
	partners []PartnerGetter, code int, err error) {
	return ListPartners(r.Conn, host, args)
}

func (r ServiceEx) GetPartner(host string, name string) (
	partner PartnerGetter, code int, err error) {
	return GetPartner(r.Conn, host, name)
}

func (r ServiceEx) AddPartner(host string, partner PartnerSetter) (
	code int, err error) {
	return AddPartner(r.Conn, host, partner)
}

func (r ServiceEx) UpdatePartner(host string, partner PartnerSetter) (
	code int, err error) {
	return UpdatePartner(r.Conn, host, partner)
}

//	Args:
//	effectTime, deadline: 2006-01-02
//	reward type: DiscountReward | SpecificPriceReward | RangePriceReward
func (r ServiceEx) AddPartnerReward(host string, partnerName string, rewardType RewardCategory,
	title, desc, effectTime, deadline string, reward interface{}) (code int, err error) {
	return AddPartnerReward(r.Conn, host, partnerName, rewardType, title, desc, effectTime, deadline, reward)
}

func (r ServiceEx) GetPartnerReward(host string, id, name string) (
	pr PartnerReward, code int, err error) {
	return GetPartnerReward(r.Conn, host, id, name)
}

func (r ServiceEx) SetPartnerRewardAvailable(host string, id string, avail bool) (
	code int, err error) {
	return SetPartnerRewardAvailable(r.Conn, host, id, avail)
}

func (r ServiceEx) GetDiscountBill(host string, serialNum string) (
	dt DiscountTransaction, code int, err error) {
	return GetDiscountBill(r.Conn, host, serialNum)
}

//	Args:
//	rewardId，可选，不选设为: ""
//	partnerName，可选，不选设为: “”
//	from，可选，不选设为:""，格式: "20060102"
//	to，可选，不选设为:""，格式: "20060102"
//	offset，可选，不选设为: 0
//	limit，可选，不选设为: 0
func (r ServiceEx) ListDiscountBill(host string, rewardId, partnerName, from, to string, offset, limit int) (
	dts []DiscountTransaction, code int, err error) {
	return ListDiscountBill(r.Conn, host, rewardId, partnerName, from, to, offset, limit)
}

//	Args:
//	rewardId，可选，不选设为: ""
//	partnerName，可选，不选设为: “”
//	from，可选，不选设为:""，格式: "20060102"
//	to，可选，不选设为:""，格式: "20060102"
func (r ServiceEx) DiscountBillSummary(host string, rewardId, partnerName, from, to string) (
	ds DiscountTransactionSummary, code int, err error) {
	return DiscountBillSummary(r.Conn, host, rewardId, partnerName, from, to)
}
