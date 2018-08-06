package report

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

const (
	REPORT_URI   = "/report"
	FEEDBACK_URI = "/feedback"
)

type Service struct {
	Host       string
	httpClient *http.Client
}

func New(host string) *Service {
	return &Service{host, http.DefaultClient}
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorCode        int    `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

func formatResponseBody(response *http.Response) (code int, err error) {
	code = response.StatusCode
	if code/100 != 2 {
		if response.ContentLength != 0 {
			if contentType, ok := response.Header["Content-Type"]; ok && contentType[0] == "application/json" {
				var errReceiver ErrorResponse
				json.NewDecoder(response.Body).Decode(&errReceiver)
				if errReceiver.ErrorCode != 0 {
					code = errReceiver.ErrorCode
				}
				if errReceiver.Error != "" {
					err = errors.New(errReceiver.Error)
				}
			}
		}
	}
	return
}

func (p *Service) sendPostRequest(uri string, params map[string][]string) (code int, err error) {
	url_ := p.Host + uri
	response, err := p.httpClient.PostForm(url_, url.Values(params))
	if err != nil {
		code = response.StatusCode
		return
	}
	defer response.Body.Close()
	code, err = formatResponseBody(response)
	return
}

func (p *Service) SendReport(data map[string][]string) (code int, err error) {
	code, err = p.sendPostRequest(REPORT_URI, data)
	return
}

func (p *Service) SendFeedBack(email, msg, client_type, client_version, client_agent, client_ip string) (code int, err error) {
	data := map[string][]string{
		"email":          {email},
		"msg":            {msg},
		"client_type":    {client_type},
		"client_version": {client_version},
		"client_agent":   {client_agent},
		"client_ip":      {client_ip},
	}
	code, err = p.sendPostRequest(FEEDBACK_URI, data)
	return
}
