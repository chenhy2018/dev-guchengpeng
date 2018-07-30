package account

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/rpc.v1"
)

const (
	SSOLoginToken = "sso_login_token"
	SSOLoginSSID  = "sso_login_ssid"
	SSOLoginState = "sso_login_state"
	SSOLoginInfo  = "sso_login_info"

	SSOLoginStateToken  = "token"
	SSOLoginStateCookie = "cookie"
)

type SSOInfo struct {
	Host     string
	ClientId string
}

// SSOService sso admin auth api client
type SSOService struct {
	adminClient *rpc.Client
	host        string
	clientId    string
}

// NewSSOService
// host: sso host
// adminTr: admin bearer token
func NewSSOService(host string, clientId string, adminTr http.RoundTripper) *SSOService {
	adminClient := &rpc.Client{Client: &http.Client{Transport: adminTr}}
	return &SSOService{adminClient: adminClient, host: host, clientId: clientId}
}

// SSOUserInfo sso return userinfo
type SSOUserInfo struct {
	Uid        uint32 `json:"uid"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	LoginToken string `json:"login_token"`
}

// SSOError sso return error
type SSOError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error
func (se SSOError) Error() string {
	return se.Message
}

func (s *SSOService) getCall(xl rpc.Logger, url_ string, ret interface{}) (err error) {
	return getCall(s.adminClient, xl, url_, ret)
}

func getCall(client *rpc.Client, xl rpc.Logger, u string, ret interface{}) (err error) {
	resp, err := client.Get(xl, u)
	if err != nil {
		return
	}

	return ssoCallRet(resp, ret)
}

func ssoCallRet(resp *http.Response, ret interface{}) (err error) {
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode/100 == 2 {
		if ret != nil && resp.ContentLength != 0 {
			err = json.NewDecoder(resp.Body).Decode(ret)
			if err != nil {
				return
			}
		}
		if resp.StatusCode == 200 || resp.StatusCode == 204 {
			return nil
		}
	}

	ssoErr := &SSOError{}
	if resp.StatusCode/100 != 2 {
		if resp.ContentLength != 0 {
			err = json.NewDecoder(resp.Body).Decode(ssoErr)
			if err != nil {
				return
			}

			err = ssoErr
		} else {
			err = fmt.Errorf("%s", http.StatusText(resp.StatusCode))
		}
	}
	return
}

// LoginRequired every req should check this
func (s *SSOService) LoginRequired(xl rpc.Logger, loginToken string) (ret SSOUserInfo, err error) {
	params := url.Values{}
	params.Set("client_id", s.clientId)
	if len(loginToken) != 0 {
		params.Set("login_token", loginToken)
	}
	url_ := fmt.Sprintf("%s/loginrequired?%s", s.host, params.Encode())
	err = s.getCall(xl, url_, &ret)
	return
}

// UinfoByToken when use token, you should use this API to get user info
func (s *SSOService) UinfoByToken(xl rpc.Logger, token string) (ret SSOUserInfo, err error) {
	params := url.Values{}
	params.Set("client_id", s.clientId)
	if len(token) != 0 {
		params.Set("token", token)
	}

	url_ := fmt.Sprintf("%s/uinfo/token?%s", s.host, params.Encode())
	err = s.getCall(xl, url_, &ret)

	return
}

// UinfoBySid when use cookie, you should use this API to get user info
func (s *SSOService) UinfoBySid(xl rpc.Logger, sid string) (ret SSOUserInfo, err error) {
	params := url.Values{}
	params.Set("client_id", s.clientId)
	if len(sid) != 0 {
		params.Set("ssid", sid)
	}
	url_ := fmt.Sprintf("%s/uinfo/session?%s", s.host, params.Encode())
	err = s.getCall(xl, url_, &ret)
	return
}

// SSODecodeToken decode token
func SSODecodeToken(raw string, secret string) (token string, ok bool) {
	rawByte, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		return
	}
	raw = string(rawByte)
	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return
	}

	vRaw := strings.TrimSpace(parts[0])
	vHash := strings.TrimSpace(parts[1])

	if len(vRaw) == 0 || len(vHash) == 0 {
		return
	}

	vRaw, err = url.QueryUnescape(vRaw)
	if err != nil {
		return
	}

	h := hmac.New(sha512.New, []byte(secret))
	if _, err = h.Write([]byte(vRaw)); err != nil {
		return
	}

	if hash_ := hex.EncodeToString(h.Sum(nil)); hash_ != vHash {
		return
	}

	token = vRaw
	ok = true
	return
}

// SSODecodeCookieValue decode cookie, secretkey must be the same as sso's secretkey
func SSODecodeCookieValue(value string, secretKey string) (raw string, createdAt time.Time, ok bool) {
	rawBytes, _ := base64.URLEncoding.DecodeString(value)
	value = string(rawBytes)

	parts := strings.SplitN(value, ",", 3)
	if len(parts) < 3 {
		return
	}

	vRaw := strings.TrimSpace(parts[0])
	vCreated := strings.TrimSpace(parts[1])
	vHash := strings.TrimSpace(parts[2])

	if vRaw == "" || vCreated == "" || vHash == "" {
		return
	}

	vTime, _ := strconv.ParseInt(vCreated, 10, 64)
	if vTime <= 0 {
		return
	}

	vRaw, err := url.QueryUnescape(vRaw)
	if err != nil {
		return
	}

	h := hmac.New(sha512.New, []byte(secretKey))
	_, err = h.Write([]byte(vRaw + vCreated))
	if err != nil {
		return
	}

	if hex.EncodeToString(h.Sum(nil)) != vHash {
		return
	}

	raw = vRaw
	createdAt = time.Unix(0, vTime)
	ok = true
	return
}

// SSOUserService sso user auth API client
type SSOUserService struct {
	userClient  *rpc.Client
	host        string
	clientId    string
	xforwardfor string
}

// NewSSOUserService
func NewSSOUserService(host, clientId string, userTr http.RoundTripper) *SSOUserService {
	client := &rpc.Client{&http.Client{Transport: userTr}}
	return &SSOUserService{userClient: client, host: host, clientId: clientId}
}

// SSOSinOut sso api signout 使用场景：修改密码过后需要调用这个接口退出
func (s *SSOUserService) SSOSignOut(xl rpc.Logger, loginToken string) (err error) {
	u := fmt.Sprintf("%s/api/signout?login_token=%s", s.host, loginToken)

	if len(s.xforwardfor) == 0 {
		return getCall(s.userClient, xl, u, nil)
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return
	}

	req.Header.Set("X-Forwarded-For", s.xforwardfor)
	resp, err := s.userClient.Do(xl, req)
	if err != nil {
		return
	}

	return ssoCallRet(resp, nil)
}

// SSOSignOutAll 退出这个loginToken 对应用户的所有登录session
func (s *SSOUserService) SSOSignOutAll(xl rpc.Logger, loginToken string) (err error) {
	u := fmt.Sprintf("%s/api/signout/all?login_token=%s", s.host, loginToken)

	if len(s.xforwardfor) == 0 {
		return getCall(s.userClient, xl, u, nil)
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return
	}

	req.Header.Set("X-Forwarded-For", s.xforwardfor)
	resp, err := s.userClient.Do(xl, req)
	if err != nil {
		return
	}

	return ssoCallRet(resp, nil)
}

// SetXForwardFor  设置 X-Forwarded-For 的header,把clientIp带到sso
func (s *SSOUserService) SetXForwardFor(ip string) {
	s.xforwardfor = ip
}
