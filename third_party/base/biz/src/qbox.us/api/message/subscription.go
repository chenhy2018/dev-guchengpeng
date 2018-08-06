package message

import (
	"net/url"
	"strconv"
	"time"
)

import (
	"github.com/qiniu/rpc.v1"
	"labix.org/v2/mgo/bson"
)

type HandleSubscription struct {
	Host   string
	Client *rpc.Client
}

type ChannelGlobalSettingModel struct {
	Id        bson.ObjectId `json:"id"`
	Type      int           `json:"type"`
	ChannelId int           `json:"channel_id"`
	Name      string        `json:"name"`

	SmsIsSet                 bool `json:"sms_is_set"`
	SmsCanChange             bool `json:"sms_can_change"`
	MailIsSet                bool `json:"mail_is_set"`
	MailCanChange            bool `json:"mail_can_change"`
	InternalMessageIsSet     bool `json:"internal_message_is_set"`
	InternalMessageCanChange bool `json:"internal_message_can_change"`
	WechatIsSet              bool `json:"wechat_is_set"`
	WechatCanChange          bool `json:"wechat_can_change"`

	Extra string `json:"extra"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubscriptionUpdateIn struct {
	Uid      uint32                `json:"uid"`
	Settings []SubscriptionSetting `json:"settings"`
}

type SubscriptionSetting struct {
	ChannelId int `json:"channel_id"`

	SmsIsSet             bool `json:"sms_is_set"`
	MailIsSet            bool `json:"mail_is_set"`
	InternalMessageIsSet bool `json:"internal_message_is_set"`
	WechatIsSet          bool `json:"wechat_is_set"`

	Extra string `json:"extra"`
}

type ResetIn struct {
	Uid uint32 `json:"uid"`
}

func NewHandleSubscription(host string, client *rpc.Client) *HandleSubscription {
	return &HandleSubscription{host, client}
}

func (h *HandleSubscription) List(logger rpc.Logger, uid uint32) (resp []ChannelGlobalSettingModel, err error) {
	value := url.Values{}
	value.Add("uid", strconv.Itoa(int(uid)))
	err = h.Client.GetCall(logger, &resp, h.Host+"/api/subscription/list?"+value.Encode())
	return
}

func (h *HandleSubscription) Update(logger rpc.Logger, param SubscriptionUpdateIn) (resp []ChannelGlobalSettingModel, err error) {
	err = h.Client.CallWithJson(logger, &resp, h.Host+"/api/subscription/update", &param)
	return
}

func (h *HandleSubscription) Reset(logger rpc.Logger, param ResetIn) (resp []ChannelGlobalSettingModel, err error) {
	err = h.Client.CallWithJson(logger, &resp, h.Host+"/api/subscription/reset", &param)
	return
}
