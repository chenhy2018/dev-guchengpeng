package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/log.v1"
	"qbox.us/oauth"
)

const (
	SIGNUP_URI          = "/user/signup"
	SEND_ACTIVATION_URI = "/user/send_activation"
	ACTIVATE_URI        = "/user/activate"
	CHANGE_PASSWORD_URI = "/user/change_password"
	LOGOUT_URI          = "/user/logout"
	USERINFO_URI        = "/user/info"
	SET_USER_TYPE_URI   = "/set_user_type"
	CREATE_URI          = "/user/create"
)

type Service struct {
	Host         string
	ClientId     string
	ClientSecret string
	httpClient   *http.Client
}

type UserInfo struct {
	Uid           uint32 `json:"uid"`
	UserId        string `json:"userid"`
	Email         string `json:"email"`
	IsActivated   bool   `json:"is_activated"`
	UserType      uint32 `json:"user_type"`
	DeviceNum     int    `json:"device_num"`
	InvitationNum int    `json:"invitation_num"`
}

func New(accountHost, clientId, clientSecret string) *Service {
	return &Service{accountHost, clientId, clientSecret, http.DefaultClient}
}

func (account *Service) sendPostRequest(receiver interface{}, uri string, params map[string][]string) (code int, err error) {
	url_ := account.Host + uri
	params["client_id"] = []string{account.ClientId}
	response, err1 := account.httpClient.PostForm(url_, url.Values(params))
	if err1 != nil {
		err = err1
		if response != nil {
			code = response.StatusCode
		} else {
			code = 9999
		}

		log.Printf("POST Failed: %s\n\nError: %#v\n\n", url_, err)
		return
	}
	code = response.StatusCode
	log.Println(code, url_)
	defer response.Body.Close()
	if code/100 == 2 {
		if receiver != nil && response.ContentLength != 0 {
			err = json.NewDecoder(response.Body).Decode(receiver)
		}
	} else {
		if response.ContentLength != 0 {
			if contentType, ok := response.Header["Content-Type"]; ok && contentType[0] == "application/json" {
				var errReceiver oauth.ErrorResponse
				json.NewDecoder(response.Body).Decode(&errReceiver)
				if errReceiver.ErrorCode != 0 {
					code = errReceiver.ErrorCode
				}
				if errReceiver.Error != "" {
					err = errors.New(errReceiver.Error)
				}
			}
		}
		if err == nil {
			err = errors.New("E" + strconv.Itoa(code))
		}
	}
	return
}

func (account *Service) ChangePassword(access_token, password, new_password string) (code int, err error) {
	data := map[string][]string{"access_token": {access_token}, "password": {password}, "new_password": {new_password}}
	code, err = account.sendPostRequest(nil, CHANGE_PASSWORD_URI, data)
	return
}

func (account *Service) Logout(access_token, refresh_token string) (code int, err error) {
	data := map[string][]string{"access_token": {access_token}, "refresh_token": {refresh_token}}
	code, err = account.sendPostRequest(nil, LOGOUT_URI, data)
	return
}

func (account *Service) UserInfo(access_token string) (respData UserInfo, code int, err error) {
	data := map[string][]string{"access_token": {access_token}}
	code, err = account.sendPostRequest(&respData, USERINFO_URI, data)
	return
}

func (account *Service) GetOAuthTransport(username, password string) (res *oauth.Transport, err error) {
	res = &oauth.Transport{
		Config: &oauth.Config{
			ClientId:     account.ClientId,
			ClientSecret: account.ClientSecret,
			Scope:        "Scope",
			AuthURL:      "",
			TokenURL:     account.Host + "/oauth2/token",
			RedirectURL:  "",
		},
		Transport: http.DefaultTransport,
	}
	_, _, err = res.ExchangeByPassword(username, password)
	if err != nil {
		return nil, err
	}
	return
}

// DEPREACTED

func (account *Service) SetUserType(user_type, user_id, access_token string) (code int, err error) {
	log.Warn("[DEPRECATED] please use qbox.us/admin_api/v2/account.Service.UserSetUserType")
	data := map[string][]string{"access_token": {access_token}, "user_id": {user_id}, "user_type": {user_type}}
	code, err = account.sendPostRequest(nil, SET_USER_TYPE_URI, data)
	return
}
func (account *Service) Create(username, password string) (code int, err error) {
	log.Warn("[DEPRECATED] please use qbox.us/admin_api/v2/account.Service.UserCreateByPassword")
	data := map[string][]string{"email": {username}, "password": {password}}
	code, err = account.sendPostRequest(nil, CREATE_URI, data)
	return
}

func (account *Service) Signup(userid, password, fromuid string) (responseData oauth.Token, code int, err error) {
	log.Warn("[DEPRECATED] please use qbox.us/admin_api/v2/account.Service.UserCreateByPassword")
	data := map[string][]string{"email": {userid}, "password": {password}, "from": {fromuid}}
	code, err = account.sendPostRequest(&responseData, SIGNUP_URI, data)
	return
}
