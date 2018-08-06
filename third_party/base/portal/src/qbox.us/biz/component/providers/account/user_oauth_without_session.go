package account

import (
	"net/http"
	"strings"

	"qbox.us/biz/component/client"
	"qbox.us/oauth"
)

// *oauth.Transport 默认用做已登录用户的 user oauth
func UserOAuthWithoutSession(host string) interface{} {
	return func(req *http.Request, tr *client.TransportWithReqLogger) *oauth.Transport {
		var (
			token string

			auth = req.Header.Get("Authorization")
		)

		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))

		}

		userOAuth := client.NewUserOAuth(host, tr)
		userOAuth.Token = &oauth.Token{AccessToken: token}
		return userOAuth

	}

}
