package account

import (
	"net/http"

	"github.com/teapots/teapot"

	"qbox.us/biz/component/client"
	"qbox.us/biz/component/sessions"
	"qbox.us/oauth"
)

const (
	LoginUid     = "login_uid"
	LoginToken   = "login_token"
	LoginRefresh = "login_refresh"
	LoginExpired = "login_expired"
	LoginSSID    = "login_ssid"
	LoginUtype   = "user_utype"
	LoginPsk     = "LOGIN_USER_PSK"
)

// *oauth.Transport 默认用做已登录用户的 user oauth
func UserOAuth(host string) interface{} {
	return func(ctx teapot.Context, req *http.Request, tr *client.TransportWithReqLogger) *oauth.Transport {
		var (
			accessToken string
			sess        sessions.SessionStore
			userOAuth   = client.NewUserOAuth(host, tr)
		)

		// // basic auth mode
		// auth := req.Header.Get("Authorization")
		// if strings.HasPrefix(auth, "Bearer ") {
		// 	accessToken = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		// }

		ctx.Find(&sess, "")

		if sess != nil {
			accessToken = sess.Get(LoginToken).String()
			refreshToken := sess.Get(LoginRefresh).String()
			tokenExpiry := sess.Get(LoginExpired).MustInt64()

			if refreshToken != "" && isTokenExpired(tokenExpiry) {
				token, code, err := userOAuth.ExchangeByRefreshToken(refreshToken)

				var log teapot.Logger
				ctx.Find(&log, "")

				if code != http.StatusOK || err != nil {
					log.Warn("user refresh token:", code, err)
				} else {
					accessToken = token.AccessToken

					sess.Set(LoginToken, token.AccessToken)
					sess.Set(LoginRefresh, token.RefreshToken)
					sess.Set(LoginExpired, token.TokenExpiry)
				}
			}
		}

		userOAuth.Token = &oauth.Token{AccessToken: accessToken}
		return userOAuth
	}
}
