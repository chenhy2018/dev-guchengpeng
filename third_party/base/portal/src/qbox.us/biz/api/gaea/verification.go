package gaea

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

type VerificationService struct {
	host   string
	client rpc.Client
}

func NewVerificationService(host string, t http.RoundTripper) *VerificationService {
	return &VerificationService{
		host: host,
		client: rpc.Client{
			&http.Client{Transport: t},
		},
	}
}

// VerificationService.Check provides the state after user performs two-factor authentication on
// `account.qiniu.com`. If ok, you can continue sensitive operations.
func (s *VerificationService) Check(l rpc.Logger) (ok bool, err error) {
	var resp CommonResponse

	err = s.client.GetCall(l, &resp, s.host+"/api/gaea/oauth/verification/check")
	if err != nil {
		return
	}

	ok = (resp.Code == CodeOK)

	return
}

// VerificationService.Consume destroys the successful state after performs two-factor authentication on
// `account.qiniu.com`. After consumed, user will no longer be permitted to do any sensitive operation.
func (s *VerificationService) Consume(l rpc.Logger) (err error) {
	err = s.client.Call(l, nil, s.host+"/api/gaea/oauth/verification/consume")
	return
}

// due to some terrible things i need to add this func
func (s *VerificationService) CheckWithCookie(cookies []*http.Cookie, types int, l rpc.Logger) (ok bool, err error) {
	var res CommonResponse

	req, err := http.NewRequest("GET", s.host+"/api/gaea/oauth/verification/check/"+strconv.Itoa(types), nil)
	if err != nil {
		return
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	err = s.callWithCookie(req, &res, l)
	if err != nil {
		return
	}

	ok = (res.Code == CodeOK)

	return
}

func (s *VerificationService) ConsumeWithCookie(cookies []*http.Cookie, types int, l rpc.Logger) (err error) {
	req, err := http.NewRequest("POST", s.host+"/api/gaea/oauth/verification/consume/"+strconv.Itoa(types), nil)
	if err != nil {
		return
	}

	return s.callWithCookie(req, nil, l)
}

func (s *VerificationService) callWithCookie(req *http.Request, ret interface{}, l rpc.Logger) (err error) {
	resp, err := s.client.Do(l, req)
	if err != nil {
		return
	}

	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode/100 == 2 {
		if ret != nil && resp.ContentLength != 0 {
			err = json.NewDecoder(resp.Body).Decode(ret)
			if err != nil {
				return
			}
		}
		if resp.StatusCode == 200 {
			return nil
		}
	}
	return rpc.ResponseError(resp)
}
