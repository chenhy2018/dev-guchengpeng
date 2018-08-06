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

type MessageTemplateModel struct {
	Id              bson.ObjectId `json:"id"`
	ChannelId       int           `json:"channel_id"`
	Name            string        `json:"name"`
	InternalMessage string        `json:"internal_message"`
	Msg             string        `json:"msg"`
	Wechat          string        `json:"wechat"`
	MailTitle       string        `json:"mail_title"`
	MailContent     string        `json:"mail_content"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateIn struct {
	ChannelId       int    `json:"channel_id"`
	Name            string `json:"name"`
	InternalMessage string `json:"internal_message"`
	Msg             string `json:"msg"`
	Wechat          string `json:"wechat"`
	MailTitle       string `json:"mail_title"`
	MailContent     string `json:"mail_content"`
}

type UpdateIn struct {
	ChannelId       int    `json:"channel_id"`
	Name            string `json:"name"`
	InternalMessage string `json:"internal_message"`
	Msg             string `json:"msg"`
	Wechat          string `json:"wechat"`
	MailTitle       string `json:"mail_title"`
	MailContent     string `json:"mail_content"`
}

type HandleMessageTemplate struct {
	Host   string
	Client *rpc.Client
}

func NewHandleMessageTemplate(host string, client *rpc.Client) *HandleMessageTemplate {
	return &HandleMessageTemplate{host, client}
}

func (h *HandleMessageTemplate) List(logger rpc.Logger, offset int, limit int) (resp []MessageTemplateModel, err error) {
	value := url.Values{}
	value.Add("offset", strconv.Itoa(offset))
	value.Add("limit", strconv.Itoa(limit))
	err = h.Client.GetCall(logger, &resp, h.Host+"/api/message_template/list?"+value.Encode())
	return
}

func (h *HandleMessageTemplate) Create(logger rpc.Logger, param CreateIn) (resp MessageTemplateModel, err error) {
	err = h.Client.CallWithJson(logger, &resp, h.Host+"/api/message_template/create", &param)
	return
}

func (h *HandleMessageTemplate) Update(logger rpc.Logger, param UpdateIn) (resp MessageTemplateModel, err error) {
	err = h.Client.CallWithJson(logger, &resp, h.Host+"/api/message_template/update", &param)
	return
}

func (h *HandleMessageTemplate) Delete(logger rpc.Logger, id string) (err error) {
	value := url.Values{}
	value.Add("id", id)
	err = h.Client.CallWithJson(logger, nil, h.Host+"/api/message_template/delete", nil)
	return
}
