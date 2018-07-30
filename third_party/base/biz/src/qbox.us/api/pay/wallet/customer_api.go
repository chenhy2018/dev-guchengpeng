package wallet

import (
	"net/url"
	//"launchpad.net/mgo/bson"
	//"github.com/qiniu/log.v1"
	"strconv"
)

type CustomerRewardSetter struct {
	Title              string      `json:"title" bson:"title"`
	Desc               string      `json:"desc" bson:"desc"`
	PartnerName        string      `json:"partner_name" bson:"partner_name"`
	CustomerRewardType string      `json:"client_reward_type" bson:"client_reward_type"`
	CustomerReward     interface{} `json:"client_reward" bson:"client_reward"`
	EffectTime         int64       `json:"effect_time" bson:"effect_time"`
	Deadline           int64       `json:"deadline" bson:"deadline"`
	ApplicablePeople   int         `json:"applicable_people" bson:"applicable_people"`
}

type CustomerReward struct {
	Id                 string      `json:"id" bson:"id"`
	Title              string      `json:"title" bson:"title"`
	Desc               string      `json:"desc" bson:"desc"`
	PartnerName        string      `json:"partner_name" bson:"partner_name"`
	PartnerRewardId    string      `json:"partner_reward" bson:"partner_reward"`
	CustomerRewardType string      `json:"client_reward_type" bson:"client_reward_type"`
	CustomerReward     interface{} `json:"client_reward" bson:"client_reward"`
	EffectTime         int64       `json:"effect_time" bson:"effect_time"`
	Deadline           int64       `json:"deadline" bson:"deadline"`
	CreateAt           int64       `json:"create_at" bson:"create_at"`
	UpdateAt           int64       `json:"update_at" bson"update_at"`
	ApplicablePeople   int         `json:"applicable_people" bson:"applicable_people"`
	IsAvaliable        bool        `json:"is_avaliable" bson:"is_avaliable"`
}

//	Args:
//	effectTime, deadline: per 100ns
//	customerReward type: DiscountReward | SpecificPriceReward | RangePriceReward
func (r ServiceEx) AddCustomerReward(host string, title, desc, partnerName string,
	customerRewardType RewardCategory, effectTime, deadline int64,
	applicablePeople int, customerReward interface{}) (id string, code int, err error) {
	pr := CustomerRewardSetter{
		Title:              title,
		Desc:               desc,
		PartnerName:        partnerName,
		CustomerRewardType: string(customerRewardType),
		CustomerReward:     customerReward,
		EffectTime:         effectTime,
		Deadline:           deadline,
		ApplicablePeople:   applicablePeople,
	}
	code, err = r.Conn.CallWithJson(&id, host+"/customer/reward/add", pr)
	return
}

func (r ServiceEx) GetCustomerReward(host string, id string) (pr CustomerReward, code int, err error) {
	v := url.Values{}
	v.Add("id", id)
	code, err = r.Conn.Call(&pr, host+"/customer/reward/get?"+v.Encode())
	return
}

func (r ServiceEx) GetCustomerRewardList(host string, name string, offset, limit int) (pr []CustomerReward, code int, err error) {
	v := url.Values{}
	v.Add("name", name)
	v.Add("offset", strconv.Itoa(offset))
	v.Add("limit", strconv.Itoa(limit))
	code, err = r.Conn.Call(&pr, host+"/customer/reward/list?"+v.Encode())
	return
}

func (r ServiceEx) SetCustomerRewardAvailable(host string, id string, avail bool) (code int, err error) {
	v := url.Values{}
	v.Add("id", id)
	v.Add("available", strconv.FormatBool(avail))
	return r.Conn.CallWithForm(nil, host+"/customer/reward/available", v)
}

func (r ServiceEx) UpdateCustomerRewardDesc(host string, id, desc string) (code int, err error) {
	v := url.Values{}
	v.Add("id", id)
	v.Add("desc", desc)
	return r.Conn.CallWithForm(nil, host+"/customer/reward/updatedesc", v)
}

func (r ServiceEx) CalculateCustomerMoney(host string, uid uint32, cost int64, id string) (get int64, code int, err error) {
	v := url.Values{}
	v.Add("id", id)
	v.Add("uid", strconv.FormatUint(uint64(uid), 10))
	v.Add("cost", strconv.FormatInt(cost, 10))
	code, err = r.Conn.CallWithForm(&get, host+"/customer/reward/calculate", v)
	return
}
