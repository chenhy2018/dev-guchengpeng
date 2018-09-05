package gaeaadmin

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"labix.org/v2/mgo/bson"
	"qbox.us/api/pay/pay"
)

func (s *gaeaAdminService) TransactionList(params TrListParams) (res []Transaction, err error) {
	var (
		resp struct {
			apiResultBase
			Data []Transaction `json:"data"`
		}
		api = fmt.Sprintf("%s/api/finance/transaction", s.host)
	)

	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, params.Values())
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data
	return
}

type TrListParams struct {
	Uid      *uint32            `param:"uid" url:"url"`
	Uids     []uint32           `param:"uids" url:"uids"`
	Email    *string            `param:"email" url:"email"`
	From     *time.Time         `param:"from" url:"from"`
	To       *time.Time         `param:"to" url:"to"`
	Paytype  *PaymentType       `param:"payment_type" url:"payment_type"`
	BankCode *string            `param:"bank_code" url:"bank_code"`
	Status   *TransactionStatus `param:"status" url:"status"`
	PageSize int                `param:"page_size" url:"page_size"`
	Marker   *string            `param:"marker" url:"marker"`
}

func (p *TrListParams) Values() url.Values {
	values := url.Values{}

	if p.Uid != nil {
		values.Set("uid", strconv.FormatUint(uint64(*p.Uid), 10))
	}

	if len(p.Uids) > 0 {
		values.Set("uids", strings.Join(func(uids []uint32) []string {
			uidStrs := make([]string, len(uids))
			for i, uid := range uids {
				uidStrs[i] = strconv.FormatUint(uint64(uid), 10)
			}
			return uidStrs
		}(p.Uids), ","))
	}

	if p.Email != nil {
		values.Set("email", *p.Email)
	}

	if p.From != nil {
		values.Set("from", p.From.Format(time.RFC3339))
	}

	if p.To != nil {
		values.Set("to", p.To.Format(time.RFC3339))
	}

	if p.Paytype != nil {
		values.Set("payment_type", string(*p.Paytype))
	}

	if p.BankCode != nil {
		values.Set("bank_code", *p.BankCode)
	}

	if p.Status != nil {
		values.Set("status", strconv.Itoa(int(*p.Status)))
	}

	if p.Marker != nil {
		values.Set("marker", *p.Marker)
	}

	if p.PageSize > 0 {
		values.Set("page_size", strconv.Itoa(p.PageSize))
	}

	return values
}

type Transaction struct {
	Id                  bson.ObjectId     `json:"id"`
	Uid                 uint32            `json:"uid"`
	Email               string            `json:"email"`
	TimeStamp           time.Time         `json:"time"`
	TotalFee            pay.Money         `json:"total_fee"`
	PayType             PaymentType       `json:"payment_type"` // 充值流水类型
	BankCode            string            `json:"bank_code"`    // 银行码， 线上，线下银行充值可能会有
	RewardId            string            `json:"reward_id"`
	PrepaidCardNum      string            `json:"prepaid_card_num"`
	Payer               string            `json:"payer"`
	NotifyId            string            `json:"notify_id"`
	Pay3rdTradeId       string            `json:"pay_3rd_trade_no"`
	WalletTranscationId string            `json:"wallet_tx_id"`
	Status              TransactionStatus `json:"status"`
}

type TransactionStatus int

const (
	Unpay               TransactionStatus = 0
	PayFinished         TransactionStatus = 3
	PayFail             TransactionStatus = 20
	PayRechargeFail     TransactionStatus = 21
	Pay3rdInvalidStatus TransactionStatus = 22
	Pay3rdInvalidArg    TransactionStatus = 24
)

func (s TransactionStatus) IsInvalid() bool {
	return s < Unpay || (s > PayFinished && s < PayFail) || s > Pay3rdInvalidArg
}

type PaymentType string

const (
	PayTypeAlipay       PaymentType = "alipay"        // 支付宝充值，包括通过支付宝网管的银联充值
	PayTypeBankTransfer PaymentType = "bank_transfer" // 线下银行转账
	PayTypePrepaidCard  PaymentType = "prepaid_card"  // 储值卡充值
)

func (t *PaymentType) Valid() bool {
	switch *t {
	case PayTypeAlipay:
		return true
	case PayTypeBankTransfer:
		return true
	case PayTypePrepaidCard:
		return true
	default:
		return false
	}
}
