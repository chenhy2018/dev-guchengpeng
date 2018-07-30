// account 相关API, 文档位于: https://github.com/qbox/service/blob/develop/apidoc/v6/acc.md
package account

import (
	"net/http"
	"strconv"
	"time"

	"github.com/qiniu/rpc.v1"
	"qbox.us/oauth"
)

const (
	CHANGE_PASSWORD_URI = "/user/change_password"
	LOGOUT_URI          = "/user/logout"
	USERINFO_URI        = "/user/info"
)

type Service struct {
	Host string
	Conn rpc.Client
}

type UserInfo struct {
	Uid                   uint32    `json:"uid"`
	UserId                string    `json:"userid"`
	Email                 string    `json:"email"`
	Username              string    `json:"username"`
	ParentUid             uint32    `json:"parent_uid"`
	IsActivated           bool      `json:"is_activated"`
	UserType              uint32    `json:"user_type"`
	DeviceNum             int       `json:"device_num"`
	InvitationNum         int       `json:"invitation_num"`
	LastParentOperationAt time.Time `json:"last_parent_operation_at"`
}

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

func NewService(host, clientId, clientSecret, username, password string) (service *Service, err error) {
	transport := &oauth.Transport{
		Config: &oauth.Config{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			Scope:        "Scope",
			AuthURL:      "",
			TokenURL:     host + "/oauth2/token",
			RedirectURL:  "",
		},
		Transport: http.DefaultTransport, // it is default
	}
	_, _, err = transport.ExchangeByPassword(username, password)
	if err != nil {
		return
	}
	client := &http.Client{Transport: transport}
	service = &Service{
		Host: host,
		Conn: rpc.Client{client},
	}
	return
}

func NewServiceByRefreshToken(host, clientId, clientSecret, refreshToken string) (service *Service, err error) {
	transport := &oauth.Transport{
		Config: &oauth.Config{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			Scope:        "Scope",
			AuthURL:      "",
			TokenURL:     host + "/oauth2/token",
			RedirectURL:  "",
		},
		Transport: http.DefaultTransport, // it is default
	}
	_, _, err = transport.ExchangeByRefreshToken(refreshToken)
	if err != nil {
		return
	}
	client := &http.Client{Transport: transport}
	service = &Service{
		Host: host,
		Conn: rpc.Client{client},
	}
	return
}

func (r Service) OAuthToken() *oauth.Token {
	return r.Conn.Client.Transport.(*oauth.Transport).Token
}

func (r *Service) ChangePassword(password, new_password string, l rpc.Logger) (err error) {
	err = r.Conn.CallWithForm(l, nil, r.Host+CHANGE_PASSWORD_URI,
		map[string][]string{"password": {password}, "new_password": {new_password}})
	return
}

func (r *Service) Logout(refresh_token string, l rpc.Logger) (err error) {
	err = r.Conn.CallWithForm(l, nil, r.Host+LOGOUT_URI, map[string][]string{
		"refresh_token": {refresh_token},
	})
	return
}

func (r *Service) UserInfo(l rpc.Logger) (respData UserInfo, err error) {
	err = r.Conn.CallWithForm(l, &respData, r.Host+USERINFO_URI, map[string][]string{
		"dummy": {"a"},
	})
	return
}

func (r *Service) UserCreateChild(email, password string, l rpc.Logger) (info UserInfo, err error) {
	err = r.Conn.CallWithForm(l, &info, r.Host+"/user/create_child", map[string][]string{
		"email":    {email},
		"password": {password},
	})
	return
}

func (r *Service) UserDisableChild(uid uint32, reason string, l rpc.Logger) (info UserInfo, err error) {
	err = r.Conn.CallWithForm(l, &info, r.Host+"/user/disable_child", map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"reason": {reason},
	})
	return
}

func (r *Service) UserEnableChild(uid uint32, l rpc.Logger) (info UserInfo, err error) {
	err = r.Conn.CallWithForm(l, &info, r.Host+"/user/enable_child", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

func (r *Service) UserChildren(offset, limit int, l rpc.Logger) (infos []UserInfo, err error) {
	err = r.Conn.CallWithForm(l, &infos, r.Host+"/user/children", map[string][]string{
		"offset": {strconv.FormatInt(int64(offset), 10)},
		"limit":  {strconv.FormatInt(int64(limit), 10)},
	})
	return
}
