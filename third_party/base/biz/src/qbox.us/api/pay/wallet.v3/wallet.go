package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
	"qbox.us/zone"
)

type HandleWallet struct {
	Host   string
	Client *rpc.Client
}

func NewHandleWallet(host string, client *rpc.Client) *HandleWallet {
	return &HandleWallet{host, client}
}

func (r HandleWallet) DeductBasebill(logger rpc.Logger, req ModelBaseBill) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/wallet/deduct/basebill", req)
	return
}

func (r HandleWallet) DeductRevisebill(logger rpc.Logger, req ModelBaseBill) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/wallet/deduct/revisebill", req)
	return
}

func (r HandleWallet) DeductEvmbasebill(logger rpc.Logger, req ModelBaseBillEvm) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/wallet/deduct/evmbasebill", req)
	return
}

func (r HandleWallet) DeductCustombill(logger rpc.Logger, req ModelCustomBill) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/wallet/deduct/custombill", req)
	return
}

type ReqRecharge struct {
	Excode  string `json:"excode"`
	Type    string `json:"type"`
	Uid     uint32 `json:"uid"`
	Money   Money  `json:"money"`
	At      Second `json:"at"` // 业务操作时间
	Desc    string `json:"desc"`
	Details string `json:"details"`
}

func (r HandleWallet) Recharge(logger rpc.Logger, req ReqRecharge) (id string, err error) {
	value := url.Values{}
	value.Add("excode", req.Excode)
	value.Add("type", req.Type)
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("money", req.Money.ToString())
	value.Add("at", req.At.ToString())
	value.Add("desc", req.Desc)
	value.Add("details", req.Details)
	err = r.Client.CallWithForm(logger, &id, r.Host+"/v3/wallet/recharge", map[string][]string(value))
	return
}

func (r HandleWallet) RechargeFreenb(logger rpc.Logger, req ReqRecharge) (id string, err error) {
	value := url.Values{}
	value.Add("excode", req.Excode)
	value.Add("type", req.Type)
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("money", req.Money.ToString())
	value.Add("at", req.At.ToString())
	value.Add("desc", req.Desc)
	value.Add("details", req.Details)
	err = r.Client.CallWithForm(logger, &id, r.Host+"/v3/wallet/recharge/freenb", map[string][]string(value))
	return
}

type ModelBalance struct {
	Money  Money `json:"money"`  //cash+Nb
	Cash   Money `json:"cash"`   //现金
	Coupon Money `json:"coupon"` //优惠券
	Nb     Money `json:"nb"`     //牛币
}

func (r HandleWallet) Balance(logger rpc.Logger, req ReqUid) (resp ModelBalance, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/wallet/balance?"+value.Encode())
	return
}

type ModelUndeductBills struct {
	BaseBills   []ModelBaseBill    `json:"basebills"`
	EvmBills    []ModelBaseBillEvm `json:"evmbills"`
	CustomBills []ModelCustomBill  `json:"custombills"`
	ReviseBills []ModelBaseBill    `json:"revisebills"`
}

type ModelWalletOverview struct {
	Balance                 Money                `json:"balance"`                  // CASH + NB - 未支付流水金额
	Cash                    Money                `json:"cash"`                     // 现金
	Nb                      Money                `json:"nb"`                       // 牛币
	UndeductBillsMoney      Money                `json:"undeduct_bills_money"`     // 未扣费的账单金额
	UncompletedTsMoney      Money                `json:"uncompleted_ts_money"`     // 未完全支付的流水金额
	CouponUnused            Money                `json:"coupon_unused"`            // 优惠券, 未使用
	CouponUsed              Money                `json:"coupon_used"`              // 优惠券, 已使用
	CouponOverdue           Money                `json:"coupon_overdue"`           // 优惠券, 已过期
	UndeductBills           ModelUndeductBills   `json:"undeduct_bills"`           // 已出账未扣费账单
	UncompletedTransactions []ModelTransactionV4 `json:"uncompleted_transactions"` // 未完全支付的扣费流水
}

func (r HandleWallet) Overview(logger rpc.Logger, req ReqUid) (resp ModelWalletOverview, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/wallet/overview?"+value.Encode())
	return
}

type ModelBalanceWithCoupon struct {
	Money         Money `json:"money"`         //cash+Nb
	Cash          Money `json:"cash"`          //现金
	Nb            Money `json:"nb"`            //牛币
	CouponUnused  Money `json:"couponUnused"`  //优惠券, 未使用
	CouponUsed    Money `json:"couponUsed"`    //优惠券, 已使用
	CouponOverdue Money `json:"couponOverdue"` //优惠券, 已过期
}

func (r HandleWallet) BalanceWithCoupon(logger rpc.Logger, req ReqUid) (resp ModelBalanceWithCoupon, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/wallet/balance/with/coupon?"+value.Encode())
	return
}

type ModelRealTimeInfo struct {
	Uid     uint32          `json:"uid"`
	Day     Day             `json:"day"`
	Money   Money           `json:"money"`
	Details map[Group]Money `json:"details"`
}

type ReqRealtimeInfo struct {
	Uid uint32 `json:"uid"`
	Day Day    `json:"day"`
}

func (r HandleWallet) RealtimeInfo(logger rpc.Logger, req ReqRealtimeInfo) (resp ModelRealTimeInfo, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("day", req.Day.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/wallet/realtime/info?"+value.Encode())
	return
}

type ReqWriteoff struct {
	Uid                 uint32      `json:"uid"`
	IsMonthstatement    bool        `json:"is_monthstatement"`    // true for monthstatement writeoff
	RegenMonthstatement bool        `json:"regen_monthstatement"` // only for monthstatement writeoff
	Scope               Scope       `json:"scope"`                // only for monthstatement writeoff
	Zones               []zone.Zone `json:"zones"`                // only for monthstatement writeoff, empty means all zones
	Month               Month       `json:"month"`                // only for monthstatement writeoff
	IsRecharge          bool        `json:"is_recharge"`          // true for recharge writeoff
	RechargeIds         []string    `json:"recharge_ids"`         // only for recharge writeoff
}

func (r HandleWallet) Writeoff(logger rpc.Logger, req ReqWriteoff) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/wallet/writeoff", req)
	return
}

func (r HandleWallet) WriteoffCustombill(logger rpc.Logger, req ReqID) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/wallet/writeoff/custombill", req)
	return
}

type ReqWriteoffRtbills struct {
	Uid     uint32    `json:"uid"`
	Product Product   `json:"product"`
	Zone    zone.Zone `json:"zone"`
	From    Second    `json:"from"`
	To      Second    `json:"to"`
}

func (r HandleWallet) WriteoffRtbills(logger rpc.Logger, req ReqWriteoffRtbills) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/wallet/writeoff/rtbills", req)
	return
}
