package account

import (
	"github.com/teapots/inject"

	"qbox.us/biz/services.v2/account"
	"qbox.us/oauth"
)

func SSOService(host string, clientId string) interface{} {
	return inject.Provide{
		inject.Dep{0: "admin"},
		func(adminTr *oauth.Transport) *account.SSOService {
			return account.NewSSOService(host, clientId, adminTr)
		},
	}
}
