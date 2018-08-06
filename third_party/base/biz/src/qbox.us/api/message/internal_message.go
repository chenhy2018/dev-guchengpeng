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

type InternalMessageModel struct {
	Id           bson.ObjectId `json:"id"`
	Uid          uint32        `json:"uid"`
	TemplateName string        `json:"template_name"`
	Data         interface{}   `json:"data"`
	ChannelId    int           `json:"channel_id"`
	IsRead       bool          `json:"is_read"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type HandleInternalMessage struct {
	Host   string
	Client *rpc.Client
}

type MarkReadIn struct {
	Id  string `json:"id"`
	Uid uint32 `json:"uid"`
}

func NewHandleInternalMessage(host string, client *rpc.Client) *HandleInternalMessage {
	return &HandleInternalMessage{host, client}
}

func (h *HandleInternalMessage) List(logger rpc.Logger, uid uint32, start time.Time) (resp []InternalMessageModel, err error) {
	value := url.Values{}
	value.Add("uid", strconv.Itoa(int(uid)))
	value.Add("start", start.Format("20060102"))
	err = h.Client.GetCall(logger, &resp, h.Host+"/api/internal_message/list?"+value.Encode())
	return
}

func (h *HandleInternalMessage) UnreadMessages(logger rpc.Logger, uid uint32) (count int, err error) {
	value := url.Values{}
	value.Add("uid", strconv.Itoa(int(uid)))
	err = h.Client.GetCall(logger, &count, h.Host+"/api/internal_message/unread_number?"+value.Encode())
	return
}

func (h *HandleInternalMessage) MarkRead(logger rpc.Logger, param MarkReadIn) (err error) {
	err = h.Client.CallWithJson(logger, nil, h.Host+"/api/internal_message/markread", &param)
	return
}
