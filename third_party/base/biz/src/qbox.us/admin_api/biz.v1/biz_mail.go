package biz

import (
	"time"
)

import (
	"github.com/qiniu/rpc.v1"
	wallet "qbox.us/api/pay/wallet/v2"
)

type MailArg struct {
	Uid          uint32       `json:"uid"`
	Money        wallet.Money `json:"money,omitempty"`
	Cash         wallet.Money `json:"cash,omitempty"`
	Coupon       wallet.Money `json:"coupon,omitempty"`
	Now          *time.Time   `json:"now,omitempty"`
	FreezeTime   *time.Time   `json:"freeze_time,omitempty"`
	UnfreezeTime *time.Time   `json:"unfreeze_time,omitempty"`
	FreezeReason string       `json:"reason,omitempty"`
}

// 欠费冻结前提醒
// Uid: 该用户uid
// Money: 需充值金额
// Cash: 账户现金余额
// Coupon: 账户抵用券余额
// FreezeTime: 冻结时间
// Now: 当前时间
func (s *BizService) MailFreezeRemind(l rpc.Logger, arg *MailArg) (err error) {
	return s.rpc.CallWithJson(l, nil, s.host+"/mail/freeze-remind", arg)
}

// 欠费冻结
// Uid: 该用户uid
// Money: 需充值金额
// Cash: 账户现金余额
// Coupon: 账户抵用券余额
// FreezeTime: 冻结时间
// Now: 当前时间
func (s *BizService) MailFrozeByArrearage(l rpc.Logger, arg *MailArg) (err error) {
	return s.rpc.CallWithJson(l, nil, s.host+"/mail/froze-arrearage", arg)
}

// 人工冻结
// Uid: 该用户uid
// FreezeTime: 冻结时间
// FreezeReason: 冻结原因
func (s *BizService) MailFrozeByManual(l rpc.Logger, arg *MailArg) (err error) {
	return s.rpc.CallWithJson(l, nil, s.host+"/mail/froze-manual", arg)
}

// 解冻
// Uid: 该用户uid
// UnfreezeTime: 解冻时间
// Now: 当前时间
// Cash: 账户现金余额
// Coupon: 账户抵用券余额
func (s *BizService) MailUnfreeze(l rpc.Logger, arg *MailArg) (err error) {
	return s.rpc.CallWithJson(l, nil, s.host+"/mail/unfreeze", arg)
}

// 欠费提醒
// Uid: 该用户uid
// Money: 需充值金额
// Cash: 账户现金余额
// Coupon: 账户抵用券余额
// Now: 当前时间
func (s *BizService) MailArrearage(l rpc.Logger, arg *MailArg) (err error) {
	return s.rpc.CallWithJson(l, nil, s.host+"/mail/arrearage", arg)
}

// 余额不足提醒
// Uid: 该用户uid
// Money: 需充值金额
// Cash: 账户现金余额
// Coupon: 账户抵用券余额
// Now: 当前时间
func (s *BizService) MailCashNotEnough(l rpc.Logger, arg *MailArg) (err error) {
	return s.rpc.CallWithJson(l, nil, s.host+"/mail/cash-not-enough", arg)
}

type StatementMailArg struct {
	Uid                uint32       `json:"uid"`
	StatementMonth     *time.Time   `json:"statement_month"`
	StatementMoney     wallet.Money `json:"statement_money"`
	StatementLastDate  *time.Time   `json:"statement_last_date"`
	StatementLeftMoney wallet.Money `json:"statement_left_monety"`
}

// 月清单（对账单）
// Uid: 该用户uid
// StatementMonth: 清单月份
// StatementMoney: 清单金额
// StatementLastDate: 月清单最后一天
// StatementLeftMoney: 月清单最后一天的现金余额

func (s *BizService) MailMonthStatement(l rpc.Logger, arg *StatementMailArg) (err error) {
	return s.rpc.CallWithJson(l, nil, s.host+"/mail/statement/month", arg)
}
