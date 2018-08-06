package ustack

import (
	"net/http"
	"sync"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v2"
)

// --------------------------------------------------

type Conn interface {
	Call(l rpc.Logger, ret interface{}, method, path string) (err error)
	CallWithForm(l rpc.Logger, ret interface{}, method, path string, params map[string][]string) (err error)
	CallWithJson(l rpc.Logger, ret interface{}, method, path string, params interface{}) (err error)
}

// --------------------------------------------------

type Tenant struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type Token struct {
	IssuedAt string `json:"issued_at"`
	Expires  string `json:"expires"`
	Id       string `json:"id"`
	Tenant   Tenant `json:"tenant"`
}

func (t *Token) expired() bool {
	tim, err := time.Parse(time.RFC3339, t.Expires)
	if err != nil {
		log.Error("parse time failed:", t.Expires, err)
		return false
	}
	return tim.Before(time.Now())
}

// --------------------------------------------------

type tokensArgs struct {
	Auth struct {
		TenantName          string `json:"tenantName"`
		PasswordCredentials struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"passwordCredentials"`
	} `json:"auth"`
}

type tokensRefreshRet struct {
	Access struct {
		Token Token `json:"token"`
	} `json:"access"`
}

type transport struct {
	token       *Token
	refreshArgs *tokensArgs
	refreshConn Conn
	rt          http.RoundTripper
	mutex       sync.RWMutex
}

func newTransport(
	token *Token, refreshArgs *tokensArgs, refreshConn Conn, rt http.RoundTripper) *transport {

	if rt == nil {
		rt = http.DefaultTransport
	}
	return &transport{
		token:       token,
		refreshArgs: refreshArgs,
		refreshConn: refreshConn,
		rt:          rt,
	}
}

func (t *transport) getToken() (token *Token, err error) {

	t.mutex.RLock()
	token = t.token
	t.mutex.RUnlock()

	if token.expired() {
		return t.refreshToken()
	}
	return
}

func (t *transport) refreshToken() (token *Token, err error) {

	var ret tokensRefreshRet
	err = t.refreshConn.CallWithJson(nil, &ret, "POST", "/v2.0/tokens", t.refreshArgs)
	if err != nil {
		return
	}
	token = new(Token)
	*token = ret.Access.Token

	t.mutex.Lock()
	t.token = token
	t.mutex.Unlock()
	return
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	token, err := t.getToken()
	if err != nil {
		return
	}

	req.Header.Set("X-Auth-Token", token.Id)
	return t.rt.RoundTrip(req)
}

func (t *transport) NestedObject() interface{} {

	return t.rt
}

// --------------------------------------------------
