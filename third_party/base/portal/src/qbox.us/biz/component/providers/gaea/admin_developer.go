package gaea

import (
	"github.com/teapots/inject"
	"github.com/teapots/teapot"

	"qbox.us/biz/services.v2/gaea"
	"qbox.us/oauth"
)

func AdminDeveloper(host, oauthName string) interface{} {
	return inject.Provide{
		inject.Dep{0: oauthName},
		func(adminOAuth *oauth.Transport, log teapot.ReqLogger) gaea.AdminDeveloperService {
			return gaea.NewAdminDeveloperService(host, adminOAuth, log)
		},
	}
}
