package gaeaadmin

import (
	"fmt"
)

type SendSmsParams struct {
	UIDs           []uint32        `json:"uids"`
	Message        string          `json:"message"`
	CustomMessages []CustomMessage `json:"custom_messages"`
}

type CustomMessage struct {
	UID     uint32 `json:"uid"`
	Message string `json:"message"`
}

type SendSmsOutput struct {
	JobID string `json:"job_id"`
}

func (s *gaeaAdminService) SendSms(params SendSmsParams) (out SendSmsOutput, err error) {
	var (
		resp struct {
			apiResultBase
			Data SendSmsOutput `json:"data"`
		}
		api = fmt.Sprintf("%s/api/notification/sms", s.host)
	)

	err = s.client.CallWithJson(s.reqLogger, &resp, api, &params)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	out = resp.Data

	return
}

type SendMailParams struct {
	UIDs           []uint32     `json:"uids"`
	Subject        string       `json:"subject"`
	Message        string       `json:"message"`
	CustomMessages []CustomMail `json:"custom_messages"`
}

type CustomMail struct {
	UID     uint32 `json:"uid"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type SendMailOutput struct {
	JobID string `json:"job_id"`
}

func (s *gaeaAdminService) SendMail(params SendMailParams) (out SendMailOutput, err error) {
	var (
		resp struct {
			apiResultBase
			Data SendMailOutput `json:"data"`
		}
		api = fmt.Sprintf("%s/api/notification/mail", s.host)
	)

	err = s.client.CallWithJson(s.reqLogger, &resp, api, &params)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	out = resp.Data

	return
}
