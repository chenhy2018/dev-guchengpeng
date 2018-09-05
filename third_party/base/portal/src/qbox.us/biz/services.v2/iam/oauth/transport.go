package oauth

import (
	"net/http"
	"strconv"

	"qbox.us/oauth"
)

// Transport holds IAM OAuth Token and auto refresh it if it comes to overdue.
type Transport struct {
	*oauth.Transport
	rootUID uint32
}

// NewTransport creates an IAM OAuth Transport.
func NewTransport(tr *oauth.Transport) *Transport {
	return &Transport{
		Transport: tr,
	}
}

// WithRootUID sets the root uid of current transport.
func (t *Transport) WithRootUID(uid uint32) *Transport {
	t.rootUID = uid
	return t
}

// ExchangeByPassword takes user & passwd and gets access token from the remote server.
func (t *Transport) ExchangeByPassword(user string, passwd string) (*oauth.Token, int, error) {
	return t.ExchangeByPasswordEx2(user, passwd, nil)
}

// ExchangeByRefreshToken takes user's refresh token and gets access token from the remote server.
func (t *Transport) ExchangeByRefreshToken(refreshToken string) (*oauth.Token, int, error) {
	return t.Transport.ExchangeByRefreshToken(refreshToken)
}

// ExchangeByPasswordEx takes user & passwd & devid and gets access Token from the remote server.
func (t *Transport) ExchangeByPasswordEx(user, passwd, devid string) (*oauth.Token, int, error) {
	params := map[string][]string{
		"device_id": {devid},
	}
	return t.ExchangeByPasswordEx2(user, passwd, params)
}

// ExchangeByPasswordEx2 takes user & passwd & extra params and gets access Token from the remote server.
func (t *Transport) ExchangeByPasswordEx2(user, passwd string, params map[string][]string) (*oauth.Token, int, error) {
	if params == nil {
		params = make(map[string][]string)
	}
	if len(params["root_uid"]) == 0 {
		params["root_uid"] = []string{strconv.Itoa(int(t.rootUID))}
	}
	return t.Transport.ExchangeByPasswordEx2(user, passwd, params)
}

// ExchangeByRefreshTokenEx takes user's refresh token & extra params and gets access token from the remote server.
func (t *Transport) ExchangeByRefreshTokenEx(refreshToken string, params map[string][]string) (*oauth.Token, int, error) {
	return t.Transport.ExchangeByRefreshTokenEx(refreshToken, params)
}

// Exchange takes a code and gets access Token from the remote server.
func (t *Transport) Exchange(code string) (*oauth.Token, int, error) {
	return t.Transport.Exchange(code)
}

// RoundTrip executes a single HTTP transaction using the Transport's
// Token as authorization headers.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.Transport.RoundTrip(req)
}

// NestedObject returns the nested transport.
func (t *Transport) NestedObject() interface{} {
	return t.Transport.NestedObject()
}

// Client returns an *http.Client that makes OAuth-authenticated requests.
func (t *Transport) Client() *http.Client {
	return &http.Client{
		Transport: t.Transport,
	}
}
