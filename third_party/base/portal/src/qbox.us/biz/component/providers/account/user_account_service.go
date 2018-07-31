package account

import (
	"github.com/teapots/teapot"

	"qbox.us/biz/services.v2/account"
	"qbox.us/oauth"
)

func UserAccountService(host string) interface{} {
	return func(userOAuth *oauth.Transport, log teapot.ReqLogger) account.AccountService {
		return account.NewAccountService(host, userOAuth, log)
	}
}
