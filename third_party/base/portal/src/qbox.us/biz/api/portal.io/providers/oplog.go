package providers

import (
	"github.com/teapots/inject"
	"github.com/teapots/teapot"
	"qbox.us/biz/api/portal.io/services"
	"qbox.us/biz/utils.v2/log"
	"qbox.us/oauth"
)

func OpLog(host string) interface{} {
	return inject.Provide{
		inject.Dep{0: "admin"},
		func(potalOauth *oauth.Transport, logger teapot.ReqLogger) services.OpLogService {
			return services.NewOpLogService(host, potalOauth, log.NewRpcWrapper(logger))
		},
	}

}
