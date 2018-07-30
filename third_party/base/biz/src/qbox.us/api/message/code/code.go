package code

import (
	"github.com/qiniu/rpc.v1"
)

const (
	//message
	ReSendEmailTooFrequent errCode = 760
	ReSendSmsTooFrequent   errCode = 761
	SendMailFailed         errCode = 770 //发送邮件失败
	SendSmsFailed          errCode = 771 //发送短信失败
	SendVoiceSmsFailed     errCode = 772 //发送语音短信失败
	ReSendMessageFailed    errCode = 773 //重发消息失败
	//status
	GetSmsStatusFailed             errCode = 780 //获取短信发送状态失败
	GetMailStatusFailed            errCode = 781 //获取邮件发送状态失败
	GetInternalMessageStatusFailed errCode = 782 //获取站内信发送状态失败
	SmsSendQPSLimit                errCode = 791
	MailSendQPSLimit               errCode = 792
	FeatureDisable                 errCode = 793
)

var (
	errCodeHumanize = map[errCode]string{
		// message
		ReSendEmailTooFrequent:         "too many mail requests",
		ReSendSmsTooFrequent:           "too many sms request",
		SendMailFailed:                 "send mail failed",
		SendSmsFailed:                  "send sms failed",
		SendVoiceSmsFailed:             "send voice sms failed",
		ReSendMessageFailed:            "resend message failed",
		GetSmsStatusFailed:             "get sms status failed",
		GetMailStatusFailed:            "get mail status failed",
		GetInternalMessageStatusFailed: "get internal message status failed",
		SmsSendQPSLimit:                "sms send is limit by qps",
		MailSendQPSLimit:               "mail send is limit by qbs",
		FeatureDisable:                 "feature is disable",
	}
)

type errCode int

func (e errCode) Code() int {
	return int(e)
}

func (e errCode) Humanize() string {
	return errCodeHumanize[e]
}

func (e errCode) Error() string {
	return e.Humanize()
}

func ParseErr(err error) error {
	if rpcErr, ok := err.(*rpc.ErrorInfo); ok {
		bizErr := errCode(rpcErr.HttpCode())
		if bizErr.Error() != "" {
			return bizErr
		}
	}
	return err
}
