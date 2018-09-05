package gaea

import (
	"net/http"

	"github.com/qiniu/rpc.v1"

	"qbox.us/biz/api/gaea"
	"qbox.us/biz/utils.v2/log"
)

type DeveloperService interface {
	Info() (*gaea.DeveloperInfo, error)
}

type developerService struct {
	logger    log.ReqLogger
	rpcLogger rpc.Logger

	impl *gaea.DeveloperService
}

var _ DeveloperService = new(developerService)

func NewDeveloperService(host string, t http.RoundTripper, logger log.ReqLogger) DeveloperService {
	return &developerService{
		logger:    logger,
		rpcLogger: log.NewRpcWrapper(logger),
		impl:      gaea.NewDeveloperService(host, t),
	}
}

func (s *developerService) Info() (info *gaea.DeveloperInfo, err error) {
	info, err = s.impl.Info(s.rpcLogger)
	return
}
