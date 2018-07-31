package ustack

import (
	"net/http"
	"sync"

	"github.com/qiniu/rpc.v2"
	"github.com/qiniu/rpc.v2/failover"
)

// --------------------------------------------------

type Conn interface {
	Call(l rpc.Logger, ret interface{}, method, path string) (err error)
	CallWithForm(l rpc.Logger, ret interface{}, method, path string, params map[string][]string) (err error)
	CallWithJson(l rpc.Logger, ret interface{}, method, path string, params interface{}) (err error)
	DoRequestWithJson(l rpc.Logger, method, path string, data interface{}) (resp *http.Response, err error)
}

func defaultNewConn(hosts []string, rt http.RoundTripper) Conn {

	return failover.New(hosts, &failover.Config{
		Http: &http.Client{
			Transport: rt,
		},
	})
}

// --------------------------------------------------

type tokensArgs struct {
	Auth struct {
		Identity struct {
			Methods  []string `json:"methods"`
			Password struct {
				User struct {
					Id       string `json:"id"`
					Password string `json:"password"`
				} `json:"user"`
			} `json:"password"`
		} `json:"identity"`
		Scope struct {
			Project struct {
				Id string `json:"id"`
			} `json:"project"`
		} `json:"scope"`
	} `json:"auth"`
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

	var ret tokensRet
	resp, err := t.refreshConn.DoRequestWithJson(nil, "POST", "/v3/auth/tokens", t.refreshArgs)
	if err != nil {
		return
	}
	err = rpc.CallRet(nil, &ret, resp)
	if err != nil && rpc.HttpCodeOf(err)/100 != 2 {
		return
	}

	token = new(Token)
	*token = ret.Token
	token.Id = resp.Header.Get("X-Subject-Token")

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
