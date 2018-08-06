package account

import (
	"net/http"
	"sync"
	"time"

	"github.com/teapots/teapot"

	"qbox.us/biz/component/client"
	"qbox.us/oauth"
)

func isTokenExpired(tokenExpiry int64) bool {
	// 减120s，提前换取新token
	return time.Now().Unix() >= tokenExpiry-120
}

func AdminOAuth(host, user, pass string) interface{} {
	token := &oauth.Token{}
	mutex := sync.Mutex{}

	return func(log teapot.ReqLogger, tr *client.TransportWithReqLogger) *oauth.Transport {
		// 每请求创建，用于使用 TransportWithReqLogger
		adminOAuth := client.NewAdminOAuth(host, tr)

		if isTokenExpired(token.TokenExpiry) {
			mutex.Lock()

			// 双层逻辑确保不会重新刷新
			if isTokenExpired(token.TokenExpiry) {
				tk, code, err := adminOAuth.ExchangeByPassword(user, pass)

				log.Info("admin token refresh status", code, err)

				if code != http.StatusOK || err != nil {
					log.Alertf("admin token refresh failed:", code, err)
				}

				// 缓存 token
				*token = *tk
			}
			mutex.Unlock()
		}

		adminOAuth.Token = token
		return adminOAuth
	}
}
