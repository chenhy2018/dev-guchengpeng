package gaea

import (
	"github.com/teapots/inject"
	"github.com/teapots/teapot"

	"qbox.us/biz/services.v2/gaea"
	"qbox.us/oauth"
)

func Developer(host, oauthName string) interface{} {
	return inject.Provide{
		inject.Dep{0: oauthName},
		func(gaeaOAuth *oauth.Transport, log teapot.ReqLogger) gaea.DeveloperService {
			return gaea.NewDeveloperService(host, gaeaOAuth, log)
		},
	}
}
