package biz

import (
	"net/url"
)

import (
	"github.com/qiniu/rpc.v1"
)

// 发送文本短信
// phone: 电话号码，多个电话号码通过","来分隔
// message: 短信内容，最大长度为 350
func (s *BizService) SendSMS(l rpc.Logger, phone string, message string) (err error) {

	param := url.Values{
		"phone":   {phone},
		"message": {message},
	}
	return s.rpc.CallWithForm(l, nil, s.host+"/sms/sendsms", param)
}

// 发送语音短信
// phone: 电话号码，只能有一个电话号码
// message: 短信内容，不能使用中文，最大长度为 8
func (s *BizService) SendVMS(l rpc.Logger, phone, message string) (err error) {

	param := url.Values{
		"phone":   {phone},
		"message": {message},
	}
	return s.rpc.CallWithForm(l, nil, s.host+"/sms/sendvms", param)
}
