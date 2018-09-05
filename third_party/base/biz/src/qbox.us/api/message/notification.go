package message

import (
	"time"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/message/code"
)

type SendMailIn struct {
	Uid   uint32 `json:"uid"`
	Name  string `json:"name"`
	From  string `json:"from"`
	Reply string `json:"reply"`

	To  []string `json:"to"`
	Cc  []string `json:"cc"`
	Bcc []string `json:"bcc"`

	Tag         []string          `json:"tag"`
	ExtraHeader map[string]string `json:"extra_header"`

	Subject string `json:"subject"`
	Content string `json:"content"`

	Options map[string]string `json:"options"`

	Provider   string `json:"provider"`
	ProviderId string `json:"provider_id"`
}

type ResendEmailInput struct {
	Id string `json:"id"`
}

type ResendSmsInput struct {
	Id string `json:"id"`
}

type SendSmsIn struct {
	Uid         uint32 `json:"uid"`
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
}

type SendInternalMessageIn struct {
	Uid          uint32      `json:"uid"`
	Data         interface{} `json:"data"`
	TemplateName string      `json:"template_name"`
	ChannelId    int         `json:"channel_id"`
}

type ListStatusInput struct {
	Uid    uint32   `json:"uid"`
	Mobile string   `json:"phone"`
	Email  string   `json:"email"`
	JobID  string   `json:"job_id"`
	Limit  int      `json:"limit"`
	Skip   int      `json:"skip"`
	Sorts  []string `json:"sorts"`
	Ids    []string `json:"ids"`
}

type SendOut struct {
	Oid string `json:"oid"`
}

type BatchSendOut struct {
	JobID string `json:"job_id"`
}

type SmsStatus struct {
	Id         string    `json:"id"`
	Uid        uint32    `json:"uid"`
	Mobile     string    `json:"phone_number"`
	Message    string    `json:"message"`
	ErrorMsg   string    `json:"error_msg"`
	Provider   string    `json:"provider"`
	ProviderId string    `json:"provider_id"`
	IsVoice    bool      `json:"is_voice"`
	QueueNew   bool      `json:"queue_new,omitempty"`
	Status     int       `json:"status"` //0:sending,1:success,2:fail
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type MailStatus struct {
	Id       string `json:"id"`
	Uid      uint32 `json:"uid"`
	ErrorMsg string `json:"error_msg"`

	Name        string            `json:"name"`
	From        string            `json:"from"`
	Reply       string            `json:"reply"`
	To          []string          `json:"to"`
	Cc          []string          `json:"cc"`
	Bcc         []string          `json:"bcc"`
	Tag         []string          `json:"tag"`
	ExtraHeader map[string]string `json:"extra_header"`
	Subject     string            `json:"subject"`
	Content     string            `json:"content"`
	Options     map[string]string `json:"options"`
	AttachFile  []string          `json:"attach_file"`
	Provider    string            `json:"provider"`
	ProviderId  string            `json:"provider_id"`

	QueueNew  bool      `json:"queue_new,omitempty"`
	Status    int       `json:"status"` //0:sending,1:success,2:fail
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HandleNotification struct {
	Host   string
	Client *rpc.Client
}

func NewHandleNotification(host string, client *rpc.Client) *HandleNotification {
	return &HandleNotification{host, client}
}

func (h *HandleNotification) SendMail(logger rpc.Logger, param SendMailIn) (resp SendOut, err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, &resp, h.Host+"/api/notification/send/mail", param))
	return
}

func (h *HandleNotification) BatchSendMail(logger rpc.Logger, param []SendMailIn) (resp BatchSendOut, err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, &resp, h.Host+"/api/notification/send/mail/batch", &param))
	return
}

func (h *HandleNotification) ReSendMail(logger rpc.Logger, param ResendEmailInput) (err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, nil, h.Host+"/api/notification/resend/mail", &param))
	return
}

func (h *HandleNotification) SendSms(logger rpc.Logger, param SendSmsIn) (resp SendOut, err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, &resp, h.Host+"/api/notification/send/sms", &param))
	return
}

func (h *HandleNotification) BatchSendSms(logger rpc.Logger, param []SendSmsIn) (resp BatchSendOut, err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, &resp, h.Host+"/api/notification/send/sms/batch", &param))
	return
}

func (h *HandleNotification) ReSendSms(logger rpc.Logger, param ResendSmsInput) (err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, nil, h.Host+"/api/notification/resend/sms", &param))
	return
}

func (h *HandleNotification) SendVoiceSms(logger rpc.Logger, param SendSmsIn) (resp SendOut, err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, &resp, h.Host+"/api/notification/send/voicesms", &param))
	return
}

func (h *HandleNotification) GetSmsStatus(logger rpc.Logger, param ListStatusInput) (status []SmsStatus, err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, &status, h.Host+"/api/status/sms", &param))
	return
}

func (h *HandleNotification) GetMailStatus(logger rpc.Logger, param ListStatusInput) (status []MailStatus, err error) {
	err = code.ParseErr(h.Client.CallWithJson(logger, &status, h.Host+"/api/status/mail", &param))
	return
}

func (h *HandleNotification) SendInternalMessage(logger rpc.Logger, param SendInternalMessageIn) (resp SendOut, err error) {
	err = h.Client.CallWithJson(logger, &resp, h.Host+"/api/notification/send/internalmessage", &param)
	return
}
