package gaeaadmin

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"labix.org/v2/mgo/bson"
	"qbox.us/api/pay/pay"
	"qbox.us/api/pay/product.v1"
	"qbox.us/biz/api/gaeaadmin/enums"
	"qbox.us/biz/services.v2/account"
	"qbox.us/biz/utils.v2/json"
	"qbox.us/biz/utils.v2/types"
)

const (
	FreezePathV2           = "%s/api/user/%d/v2/freeze"
	UnFreezePathV2         = "%s/api/user/%d/v2/unfreeze"
	ShowDelayPathV2        = "%s/api/user/%d/v2/freeze/delay"
	ListFrozenUsersPathV2  = "%s/api/user/v2/frozen"
	ListLogsPathV2         = "%s/user/v2/freeze/logs"
	BalanceNotEnoughPathV2 = "%s/api/user/%d/notification/v2/balance-not-enough"
	GetLatestFreezeLogV2   = "%s/api/user/%d/v2/freeze/logs/latest"
)

type FreezeIn struct {
	Source   enums.FreezeSource   `json:"source"`
	Operator string               `json:"operator"`
	Force    bool                 `json:"force"`
	Type     account.DisabledType `json:"type"`
	Reason   string               `json:"reason"`
	Sync     bool                 `json:"sync"`
	Dummy    bool                 `json:"dummy"`
}

func (s *gaeaAdminService) FreezeV2(uid uint32, in FreezeIn) (err error) {
	var out json.CommonResponse

	url := fmt.Sprintf(FreezePathV2, s.host, uid)

	err = s.client.CallWithJson(s.reqLogger, &out, url, &in)
	if err != nil {
		return
	}

	err = out.Error()

	return
}

type UnfreezeIn struct {
	Source   enums.FreezeSource `json:"source"`
	Operator string             `json:"operator"`
	Force    bool               `json:"force"`
	Sync     bool               `json:"sync"`
}

func (s *gaeaAdminService) UnfreezeV2(uid uint32, in UnfreezeIn) (err error) {
	var out json.CommonResponse

	url := fmt.Sprintf(UnFreezePathV2, s.host, uid)

	err = s.client.CallWithJson(s.reqLogger, &out, url, &in)
	if err != nil {
		return
	}

	err = out.Error()

	return
}

type BalanceNotEnoughIn struct {
	Balance     types.Money `json:"balance"`
	Consumption types.Money `json:"consumption"`
	RemainDays  int         `json:"remain_days"`
}

func (s *gaeaAdminService) BalanceNotEnoughNotificationV2(uid uint32, in BalanceNotEnoughIn) (err error) {
	var out json.CommonResponse

	url := fmt.Sprintf(BalanceNotEnoughPathV2, s.host, uid)

	err = s.client.CallWithJson(s.reqLogger, &out, url, &in)
	if err != nil {
		return
	}

	err = out.Error()

	return
}

type DelayShowOut struct {
	Id          bson.ObjectId            `json:"id"`
	Uid         uint32                   `json:"uid"`
	ScheduleAt  time.Time                `json:"schedule_at"`
	BlockType   account.DisabledType     `json:"type"`
	BlockReason string                   `json:"reason"`
	Balance     types.Money              `json:"balance"`
	Consumption types.Money              `json:"consumption"`
	Status      enums.FreezeTicketStatus `json:"status"`
	UpdatedAt   time.Time                `json:"updated_at"`
	CreatedAt   time.Time                `json:"created_at"`
}

type getDelayFreezeTicketResp struct {
	json.CommonResponse

	Data DelayShowOut `json:"data"`
}

func (s *gaeaAdminService) GetDelayFreezeTicketV2(uid uint32) (ticket *DelayShowOut, err error) {
	var out getDelayFreezeTicketResp

	url := fmt.Sprintf(ShowDelayPathV2, s.host, uid)

	err = s.client.GetCall(s.reqLogger, &out, url)
	if err != nil {
		return
	}

	err = out.Error()

	if err == nil {
		ticket = &out.Data
	}

	return
}

type ListFrozenUsersIn struct {
	From *time.Time `param:"from"`
	To   *time.Time `param:"to"`
}

func (i *ListFrozenUsersIn) Values() url.Values {
	values := url.Values{}

	if i.From != nil {
		values.Set("from", i.From.Format(time.RFC3339))
	}

	if i.To != nil {
		values.Set("to", i.To.Format(time.RFC3339))
	}

	return values
}

func (s *gaeaAdminService) ListFrozenUsersV2(params ListFrozenUsersIn) (uids []uint32, err error) {
	var out struct {
		json.CommonResponse
		Data []uint32 `json:"data"`
	}

	url := fmt.Sprintf(ListFrozenUsersPathV2, s.host)

	err = s.client.GetCallWithForm(s.reqLogger, &out, url, params.Values())

	if err != nil {
		return
	}

	err = out.Error()

	if err == nil {
		uids = out.Data
	}

	return
}

type FreezePayloadFinance struct {
	Balance            types.Money                               `bson:"balance" json:"balance"`                           // CASH + NB - 未支付流水金额
	Coupon             types.Money                               `bson:"coupon" json:"coupon"`                             // coupon
	Sum                types.Money                               `bson:"sum" json:"sum"`                                   // 余额 + 抵用券
	UndeductBillsMoney types.Money                               `bson:"undeduct_bills_money" json:"undeduct_bills_money"` // 未扣费的账单金额
	UngenbillsMoney    types.Money                               `bson:"ungen_bills_money" json:"ungen_bills_money"`       // 未出账金额
	Fee                types.Money                               `bson:"fee" json:"fee"`                                   // 消费
	Usage              map[pay.Item]product.RespModelUsage       `bson:"usage" json:"usage"`
	Quotas             map[pay.Item]product.RespModelQuota       `bson:"quotas" json:"quotas"`
	Consumptions       map[pay.Item]product.RespModelConsumption `bson:"consumptions" json:"consumptions"`
	Overflow           bool                                      `bson:"overflow" json:"overflow"`
	SumRecharge        types.Money                               `bson:"sum_recharge" json:"sum_recharge"` //历史充值总额
	Surge              bool                                      `bson:"surge" json:"surge"`
	AvgDailyCost       types.Money                               `bson:"avg_daily_cost" json:"avg_daily_cost"`
}

type FreezePayload struct {
	Base struct {
		Uid            uint32                  `bson:"uid" json:"uid"`
		UType          account.UserType        `bson:"utype" json:"utype"`
		IdentityType   IdentityType            `bson:"identity_type" json:"identity_type"`     // 身份认证类型
		IdentityStatus DeveloperIdentityStatus `bson:"identity_status" json:"identity_status"` // 身份认证状态
		Disabled       bool                    `bson:"disabled" json:"disabled"`
	} `bson:"base" json:"base"`
	Freeze struct {
		FreezeStrategy          enums.FreezeStrategy `bson:"freeze_strategy" json:"freeze_strategy"`                 //落库的冻结策略
		DynamicFreezeStrategy   enums.FreezeStrategy `bson:"dynamic_freeze_strategy" json:"dynamic_freeze_strategy"` //动态计算的冻结策略
		FreezeStrategyReason    string               `bson:"freeze_strategy_reason" json:"freeze_strategy_reason"`   //动态计算出的冻结策略的理由
		FreezeStrategyOperator  string               `bson:"freeze_strategy_operator" json:"freeze_strategy_operator"`
		FreezeStrategyUpdatedAt time.Time            `bson:"freeze_strategy_updated_at" json:"freeze_strategy_updated_at"`
	} `bson:"freeze" json:"freeze"`
	LastFreezeLog struct {
		Id                      bson.ObjectId `bson:"id,omitempty" json:"id,omitempty"`
		CreatedAt               time.Time     `bson:"created_at" json:"created_at"`
		Sum                     types.Money   `bson:"sum" json:"sum"`                                               // 余额 + 抵用券
		Fee                     types.Money   `bson:"fee" json:"fee"`                                               // 消费
		RechargeAfterLastFreeze types.Money   `bson:"recharge_after_last_freeze" json:"recharge_after_last_freeze"` // 上一次冻结到现在的充值金额
	} `bson:"last_freeze_log" json:"last_freeze_log"`
	Finance FreezePayloadFinance `bson:"finance" json:"finance"`
}

type ListLogsIn struct {
	PageSize   int                    `param:"page_size"`
	Uid        *uint32                `param:"uid"`
	From       *time.Time             `param:"from"`
	To         *time.Time             `param:"to"`
	Type       *enums.FreezeType      `param:"type"`
	FreezeType *account.DisabledType  `param:"freeze_type"`
	Status     *enums.FreezeLogStatus `param:"status"`
	Prev       *string                `param:"prev"`
	Next       *string                `param:"next"`
}

func (p *ListLogsIn) Values() url.Values {
	values := url.Values{}

	values.Set("page_size", strconv.Itoa(p.PageSize))

	if p.Uid != nil {
		values.Set("uid", strconv.FormatUint(uint64(*p.Uid), 10))
	}

	if p.From != nil {
		values.Set("from", p.From.Format(time.RFC3339))
	}

	if p.To != nil {
		values.Set("to", p.To.Format(time.RFC3339))
	}

	if p.Type != nil {
		values.Set("type", strconv.Itoa(int(*p.Type)))
	}

	if p.FreezeType != nil {
		values.Set("freeze_type", strconv.Itoa(int(*p.FreezeType)))
	}

	if p.Status != nil {
		values.Set("status", strconv.Itoa(int(*p.Status)))
	}

	if p.Prev != nil {
		values.Set("prev", *p.Prev)
	}

	if p.Next != nil {
		values.Set("next", *p.Next)
	}

	return values
}

type ListLogsOut struct {
	Id                bson.ObjectId            `json:"id"`
	Type              enums.FreezeType         `json:"type"`
	FreezeOperation   *enums.FreezeOperation   `json:"freeze_operation,omitempty"`
	UnfreezeOperation *enums.UnfreezeOperation `json:"unfreeze_operation,omitempty"`
	FreezeType        *account.DisabledType    `json:"freeze_type,omitempty"`
	Status            enums.FreezeLogStatus    `json:"status"`
	Uid               uint32                   `json:"uid"`
	Source            enums.FreezeSource       `json:"source"`
	Operator          string                   `json:"operator"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	FailedReason      string                   `json:"failed_reason"`
	Reason            string                   `json:"reason"`
	TicketId          bson.ObjectId            `json:"ticket_id,omitempty"`
	Payload           *FreezePayload           `json:"payload"`
	FusionTaskId      string                   `json:"fusion_task_id"`
	Ticket            struct {
		Id         bson.ObjectId            `json:"id"`
		Uid        uint32                   `json:"uid"`
		Status     enums.FreezeTicketStatus `json:"status"`
		LogIds     []bson.ObjectId          `json:"log_ids"` // 相关的logIds
		CreatedAt  time.Time                `json:"created_at"`
		UpdatedAt  time.Time                `json:"updated_at"`
		ScheduleAt time.Time                `json:"schedule_at"`
	} `json:"ticket,omitempty"`
}

func (s *gaeaAdminService) ListLogsV2(in ListLogsIn) (logs []*ListLogsOut, err error) {
	var out struct {
		json.CommonResponse
		Data []*ListLogsOut `json:"data"`
	}

	url := fmt.Sprintf(ListLogsPathV2, s.host)

	err = s.client.GetCallWithForm(s.reqLogger, &out, url, in.Values())
	if err != nil {
		return
	}

	err = out.Error()

	if err == nil {
		logs = out.Data
	}

	return
}

type FreezeLog struct {
	Id                bson.ObjectId            `bson:"_id" json:"id"`
	Type              enums.FreezeType         `bson:"type" json:"type"`
	FreezeOperation   *enums.FreezeOperation   `bson:"freeze_operation,omitempty" json:"freeze_operation,omitempty"`
	UnfreezeOperation *enums.UnfreezeOperation `bson:"unfreeze_operation,omitempty" json:"unfreeze_operation,omitempty"`
	FreezeType        *account.DisabledType    `bson:"freeze_type,omitempty" json:"freeze_type,omitempty"`
	Status            enums.FreezeLogStatus    `bson:"status" json:"status"`
	Uid               uint32                   `bson:"uid" json:"uid"`
	Source            enums.FreezeSource       `bson:"source" json:"source"`
	Operator          string                   `bson:"operator" json:"operator"`
	CreatedAt         time.Time                `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time                `bson:"updated_at" json:"updated_at"`
	FailedReason      string                   `bson:"failed_reason" json:"failed_reason"`
	Reason            string                   `bson:"reason" json:"reason"`
	TicketId          bson.ObjectId            `bson:"ticket_id,omitempty" json:"ticket_id,omitempty"`
	Payload           *FreezePayload           `bson:"payload" json:"payload"`
	FusionTaskId      string                   `bson:"fusion_task_id" json:"fusion_task_id"`
}

func (s *gaeaAdminService) GetLatestFreezeLogV2(uid uint32) (log *FreezeLog, err error) {
	var out struct {
		json.CommonResponse
		Data FreezeLog `json:"data"`
	}

	url := fmt.Sprintf(GetLatestFreezeLogV2, s.host, uid)

	err = s.client.GetCall(s.reqLogger, &out, url)
	if err != nil {
		return
	}

	err = out.Error()

	if err == nil {
		log = &out.Data
	}

	return
}
