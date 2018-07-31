package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleAgencyRecharge struct {
	Host   string
	Client *rpc.Client
}

func NewHandleAgencyRecharge(host string, client *rpc.Client) *HandleAgencyRecharge {
	return &HandleAgencyRecharge{host, client}
}

type ReqAgencyRechargeBalance struct {
	Excode           string `json:"excode"`
	Uid              uint32 `json:"uid"`
	Money            Money  `json:"money"`
	Cost             Money  `json:"cost"`
	CustomerRewardId string `json:"customer_reward_id"`
	At               Second `json:"at"`
	Desc             string `json:"desc"`
}

type RespAgencyRechargeBalance struct {
	Id       string `json:"id"`
	RewardId string `json:"reward_id"`
}

func (r HandleAgencyRecharge) RechargeBalance(logger rpc.Logger, req ReqAgencyRechargeBalance) (resp RespAgencyRechargeBalance, err error) {
	value := url.Values{}
	value.Add("excode", req.Excode)
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("money", req.Money.ToString())
	value.Add("cost", req.Cost.ToString())
	value.Add("customer_reward_id", req.CustomerRewardId)
	value.Add("at", req.At.ToString())
	value.Add("desc", req.Desc)
	err = r.Client.CallWithForm(logger, &resp, r.Host+"/v3/agencyrecharge/recharge/balance", map[string][]string(value))
	return
}

type ReqAgencyRecharge struct {
	Excode           string `json:"excode"`
	Type             string `json:"type"`
	Desc             string `json:"desc"`
	Uid              uint32 `json:"uid"`
	Money            Money  `json:"money"`
	Cost             Money  `json:"cost"`
	CustomerRewardId string `json:"customer_reward_id"`
	At               Second `json:"at"`
}

func (r HandleAgencyRecharge) Recharge(logger rpc.Logger, req ReqAgencyRecharge) (id string, err error) {
	value := url.Values{}
	value.Add("excode", req.Excode)
	value.Add("type", req.Type)
	value.Add("desc", req.Desc)
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("money", req.Money.ToString())
	value.Add("cost", req.Cost.ToString())
	value.Add("customer_reward_id", req.CustomerRewardId)
	value.Add("at", req.At.ToString())
	err = r.Client.CallWithForm(logger, &id, r.Host+"/v3/agencyrecharge/recharge", map[string][]string(value))
	return
}

type ModelDiscountTransaction struct {
	Serial_num  string            `json:"serial_num"`
	Time        HundredNanoSecond `json:"time"`
	PartnerName string            `json:"partner_name"`
	CustomerID  string            `json:"customer_id"`
	PartnerID   string            `json:"partner_id"`
	Money       Money             `json:"money"`
	Pay         Money             `json:"pay"`
	Cost        Money             `json:"cost"`
}

type ModelDiscountTransactionSummary struct {
	Count int   `json:"count"`
	Money Money `json:"money"`
	Pay   Money `json:"pay"`
	Cost  Money `json:"cost"`
}

type ReqTransactionGet struct {
	SerialNum string `json:"serial_num"`
}

func (r HandleAgencyRecharge) TransactionGet(logger rpc.Logger, req ReqTransactionGet) (resp ModelDiscountTransaction, err error) {
	value := url.Values{}
	value.Add("serial_num", req.SerialNum)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/agencyrecharge/transaction/get?"+value.Encode())
	return
}

func (r HandleAgencyRecharge) BalanceTransactionGet(logger rpc.Logger, req ReqTransactionGet) (resp ModelDiscountTransaction, err error) {
	value := url.Values{}
	value.Add("serial_num", req.SerialNum)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/agencyrecharge/balance/transaction/get?"+value.Encode())
	return
}

type ReqTransactionList struct {
	RewardId    string `json:"reward_id"`
	PartnerName string `json:"partner_name"`
	From        Day    `json:"from"`
	To          Day    `json:"to"`
	Offset      int    `json:"offset"`
	Limit       int    `json:"limit"`
}

func (r HandleAgencyRecharge) TransactionList(logger rpc.Logger, req ReqTransactionList) (resp []ModelDiscountTransaction, err error) {
	value := url.Values{}
	value.Add("reward_id", req.RewardId)
	value.Add("partner_name", req.PartnerName)
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/agencyrecharge/transaction/list?"+value.Encode())
	return
}

func (r HandleAgencyRecharge) BalanceTransactionList(logger rpc.Logger, req ReqTransactionList) (resp []ModelDiscountTransaction, err error) {
	value := url.Values{}
	value.Add("reward_id", req.RewardId)
	value.Add("partner_name", req.PartnerName)
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/agencyrecharge/balance/transaction/list?"+value.Encode())
	return
}

type ReqBillDiscountSum struct {
	RewardId    string `json:"reward_id"`
	PartnerName string `json:"partner_name"`
	From        Day    `json:"from"`
	To          Day    `json:"to"`
}

func (r HandleAgencyRecharge) TransactionSum(logger rpc.Logger, req ReqBillDiscountSum) (resp ModelDiscountTransactionSummary, err error) {
	value := url.Values{}
	value.Add("reward_id", req.RewardId)
	value.Add("partner_name", req.PartnerName)
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/agencyrecharge/transaction/sum?"+value.Encode())
	return
}

func (r HandleAgencyRecharge) BalanceTransactionSum(logger rpc.Logger, req ReqBillDiscountSum) (resp ModelDiscountTransactionSummary, err error) {
	value := url.Values{}
	value.Add("reward_id", req.RewardId)
	value.Add("partner_name", req.PartnerName)
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/agencyrecharge/balance/transaction/sum?"+value.Encode())
	return
}
