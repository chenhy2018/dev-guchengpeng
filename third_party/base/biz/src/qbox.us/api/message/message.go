package message

import (
	"github.com/qiniu/rpc.v1"

	"qbox.us/api/message/code"
)

type SendIn struct {
	Uid            uint32               `json:"uid"`
	ChannelId      int                  `json:"channel_id"`
	TemplateName   string               `json:"template_name"`
	Data           interface{}          `json:"data"`
	WechatTemplate *WechatTemplateModel `json:"wechatTemplate"`
	Tag            []string             `json:"tag"`
}

type ResendIn struct {
	Oid string `json:"oid"`
}

type WechatTemplateModel struct {
	ToUser     string                       `json:"touser"`
	TemplateId string                       `json:"template_id"`
	Url        string                       `json:"url"`
	Topcolor   string                       `json:"topcolor"`
	Data       map[string]map[string]string `json:"data"`
}

type HandleMessage struct {
	Host   string
	Client *rpc.Client
}

func NewHandleMessage(host string, client *rpc.Client) *HandleMessage {
	return &HandleMessage{host, client}
}

func (h *HandleMessage) Send(logger rpc.Logger, param SendIn) (oid string, err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, &oid, h.Host+"/api/message/send", &param))
	return
}

func (h *HandleMessage) Resend(logger rpc.Logger, param ResendIn) (err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, nil, h.Host+"/api/message/resend", &param))
	return
}
