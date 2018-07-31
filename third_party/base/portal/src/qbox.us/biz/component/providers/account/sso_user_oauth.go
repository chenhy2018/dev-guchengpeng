package account

import (
	"net/http"
	"time"

	"github.com/teapots/teapot"

	"qbox.us/biz/component/client"
	"qbox.us/biz/component/sessions"
	"qbox.us/biz/services.v2/account"
	"qbox.us/oauth"
)

func SSOUserOAuth(host string) interface{} {
	return func(ctx teapot.Context, adminService account.AdminAccountService, req *http.Request, tr *client.TransportWithReqLogger) *oauth.Transport {
		var (
			uid         uint32
			accessToken string
			log         teapot.Logger
		)

		ctx.Find(&log, "")

		// // basic auth mode
		// auth := req.Header.Get("Authorization")
		// if strings.HasPrefix(auth, "Bearer ") {
		// 	accessToken = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		// }

		var sess sessions.SessionStore
		ctx.Find(&sess, "")

		if sess != nil {
			uid = sess.Get(LoginUid).MustUint32()
			refreshToken := sess.Get(LoginRefresh).String()
			tokenExpiry := sess.Get(LoginExpired).MustInt64()

			var emptyOAuth func() *oauth.Transport

			if refreshToken != "" && isTokenExpired(tokenExpiry) {
				ctx.Find(&emptyOAuth, "")
			}

			if emptyOAuth != nil {
				token, code, err := emptyOAuth().ExchangeByRefreshToken(refreshToken)

				if code != http.StatusOK || err != nil {
					log.Warn("user refresh token:", code, err)
					uid = 0
				} else {

					expiry := time.Now().Unix() + token.TokenExpiry
					uid = token.Uid

					sess.Set(LoginUid, token.Uid)
					sess.Set(LoginToken, token.AccessToken)
					sess.Set(LoginRefresh, token.RefreshToken)
					sess.Set(LoginExpired, expiry)
					sess.Set(LoginSSID, token.SSID)
				}
			}
		}

		userOAuth := client.NewUserOAuth(host, tr)

		token, err := adminService.TokenCreate(uid)
		if err != nil {
			log.Warn("developer token create failed:", uid, err)
		} else {
			accessToken = token.AccessToken
		}

		userOAuth.Token = &oauth.Token{AccessToken: accessToken}
		return userOAuth
	}
}
