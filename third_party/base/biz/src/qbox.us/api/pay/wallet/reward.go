package wallet

import (
	"net/url"
	"qbox.us/rpc"
	"strconv"
)

func AddReward(c rpc.Client, host, excode, serial_num, type_ string, uid uint32, money int64,
	desc string) (code int, err error) {
	value := url.Values{}
	value.Add("excode", excode)
	value.Add("serial_num", serial_num)
	value.Add("type", type_)
	value.Add("uid", strconv.FormatUint(uint64(uid), 10))
	value.Add("money", strconv.FormatInt(money, 10))
	value.Add("desc", desc)
	code, err = c.CallWithForm(nil, host+"/recharge/add_reward", value)
	return
}

func GetRewards(c rpc.Client, host string, money int64, offset, limit int) (rewardInfo []RewardInfo, code int, err error) {
	value := url.Values{}
	value.Add("money", strconv.FormatInt(money, 10))
	value.Add("offset", strconv.FormatInt(int64(offset), 10))
	value.Add("limit", strconv.FormatInt(int64(limit), 10))
	code, err = c.CallWithForm(&rewardInfo, host+"/recharge/rewards", value)
	return
}

func GetRewardBySerialNum(c rpc.Client, host string, serailNum string) (rewardInfo RewardInfo, code int, err error) {
	value := url.Values{}
	value.Add("serail_num", serailNum)
	code, err = c.CallWithForm(&rewardInfo,
		host+"/recharge/reward/by_serial_num", value)
	return
}
