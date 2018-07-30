package account

import (
	"qbox.us/oauth"

	"github.com/teapots/inject"
	"github.com/teapots/teapot"

	"qbox.us/biz/services.v2/account"
)

func AdminAccountService(host string) interface{} {
	return AdminAccountServiceWithDepName(host, "admin")
}

func AdminAccountServiceWithDepName(host, dep string) interface{} {
	return inject.Provide{
		inject.Dep{0: dep},
		func(adminOAuth *oauth.Transport, log teapot.ReqLogger) (service account.AdminAccountService) {
			service = account.NewAdminAccountService(host, adminOAuth, log)
			return service
		},
	}
}
