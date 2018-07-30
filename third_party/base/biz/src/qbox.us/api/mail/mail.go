package mail

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	SEND_FEEDBACK_URI         = "/mail/send_feedback"
	SEND_INVITATION_URI       = "/mail/send_invitation"
	SEND_REGISTRATION_URI     = "/mail/send_registration"
	SEND_ACTIVATION_URI       = "/mail/send_activation"
	FORGET_PASSWORD_URI       = "/mail/forget_password"
	SEND_INVITATION_BOUNS_URI = "/mail/send_invitation_bouns"
)

type MailService struct {
	Host         string
	ClientSecret string
	httpClient   *http.Client
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func New(host, clientSecret string) *MailService {
	return &MailService{host, clientSecret, http.DefaultClient}
}

func (s *MailService) createSign(url_, to, timeNow string) string {
	salt := url_ + "?t=" + timeNow + "&email=" + to
	h := hmac.New(sha1.New, []byte(s.ClientSecret))
	h.Write([]byte(salt))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *MailService) sendPostRequest(receiver interface{}, uri string, params map[string][]string) (code int, err error) {
	url_ := s.Host + uri
	timeStr := fmt.Sprintf("%v", time.Now())
	params["t"] = []string{timeStr}
	to := ""
	if len(params["to"]) > 0 {
		to = params["to"][0]
	}
	params["sign"] = []string{s.createSign(url_, to, timeStr)}
	response, err1 := s.httpClient.PostForm(url_, url.Values(params))
	if err1 != nil {
		err = err1
		code = 599
		log.Printf("POST Failed: %s\n\nError: %#v\n\n", url_, err)
		return
	}
	defer response.Body.Close()
	code = response.StatusCode
	if code/100 == 2 {
		if receiver != nil && response.ContentLength != 0 {
			err = json.NewDecoder(response.Body).Decode(receiver)
		}
	} else {
		if response.ContentLength != 0 {
			if contentType, ok := response.Header["Content-Type"]; ok && contentType[0] == "application/json" {
				var errReceiver ErrorResponse
				json.NewDecoder(response.Body).Decode(&errReceiver)
				if errReceiver.Error != "" {
					err = errors.New(errReceiver.Error)
				}
			}
		}
	}
	return
}

func (s *MailService) SendInvitationMail(from, to, url_ string) (code int, err error) {
	data := map[string][]string{"from": {from}, "to": {to}, "url": {url_}}
	code, err = s.sendPostRequest(nil, SEND_INVITATION_URI, data)
	return
}

func (s *MailService) SendRegistrationMail(to, url_ string) (code int, err error) {
	data := map[string][]string{"to": {to}, "url": {url_}}
	code, err = s.sendPostRequest(nil, SEND_REGISTRATION_URI, data)
	return
}

func (s *MailService) SendActivationMail(to, referurl, inviteurl string) (code int, err error) {
	data := map[string][]string{"to": {to}, "url": {referurl}, "invite_url": {inviteurl}}
	code, err = s.sendPostRequest(nil, SEND_ACTIVATION_URI, data)
	return
}

func (s *MailService) SendInvitationBounsMail(to, referurl, inviteurl, num string) (code int, err error) {
	data := map[string][]string{"to": {to}, "url": {referurl}, "invite_url": {inviteurl}, "num": {num}}
	code, err = s.sendPostRequest(nil, SEND_INVITATION_BOUNS_URI, data)
	return
}

func (s *MailService) SendForgetPasswordMail(to, url_ string) (code int, err error) {
	data := map[string][]string{"to": {to}, "url": {url_}}
	code, err = s.sendPostRequest(nil, FORGET_PASSWORD_URI, data)
	return
}

func (s *MailService) SendFeedbackMail(from, body string) (code int, err error) {
	data := map[string][]string{"from": {from}, "to": {"feedback@qbox.net"}, "body": {body}}
	code, err = s.sendPostRequest(nil, SEND_FEEDBACK_URI, data)
	return
}
