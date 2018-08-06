package wallet

import (
	"net/url"
	"qbox.us/rpc"
	"strconv"
)

//隐藏帐单
func HideBill(r rpc.Client, host string, modelIn HideBillIn) (code int, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("ishide", strconv.FormatBool(modelIn.IsHide))

	code, err = r.CallWithForm(nil, host+"/hide/bill", value)
	return
}

//新建用户
func Newuser(r rpc.Client, host string, modelIn NewUserModel) (code int, err error) {
	value := url.Values{}
	value.Add("serial_num", modelIn.Serial_num)
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("excode", modelIn.Excode)
	value.Add("desc", modelIn.Desc)

	code, err = r.CallWithForm(nil, host+"/newuser", value)
	return
}

//自定义的充值赠送。开放式赠送金额，全额充入FreeNB
func RechargeFreeReward(r rpc.Client, host string, modelIn RechargeMini) (code int, err error) {
	value := url.Values{}
	value.Add("excode", modelIn.Excode)
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("money", strconv.FormatInt(int64(modelIn.Money), 10))
	value.Add("desc", modelIn.Desc)
	value.Add("details", modelIn.Details)

	code, err = r.CallWithForm(nil, host+"/recharge/free/reward", value)
	return
}

//销售销帐接口
func Writeoff(r rpc.Client, host string, modelIn WriteoffIn) (code int, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("money", strconv.FormatInt(int64(modelIn.Money), 10))
	value.Add("prefix", modelIn.Prefix)
	value.Add("desc", modelIn.Desc)
	value.Add("excode", modelIn.Excode)

	code, err = r.CallWithForm(nil, host+"/writeoff", value)
	return
}
