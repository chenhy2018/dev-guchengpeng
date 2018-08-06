package gaea

import (
	"net/http"

	"github.com/qiniu/rpc.v1"

	"qbox.us/biz/api/gaea"
	"qbox.us/biz/utils.v2/log"
)

type AdminDeveloperService interface {
	InfoByUid(uid uint32) (*gaea.DeveloperInfo, error)
	InfoByEmail(email string) (*gaea.DeveloperInfo, error)
}

type adminDeveloperService struct {
	logger    log.ReqLogger
	rpcLogger rpc.Logger

	impl *gaea.AdminDeveloperService
}

var _ AdminDeveloperService = new(adminDeveloperService)

func NewAdminDeveloperService(host string, t http.RoundTripper, logger log.ReqLogger) AdminDeveloperService {
	return &adminDeveloperService{
		logger:    logger,
		rpcLogger: log.NewRpcWrapper(logger),
		impl:      gaea.NewAdminDeveloperService(host, t),
	}
}

func (s *adminDeveloperService) InfoByUid(uid uint32) (info *gaea.DeveloperInfo, err error) {
	info, err = s.impl.InfoByUid(s.rpcLogger, uid)
	return
}

func (s *adminDeveloperService) InfoByEmail(email string) (info *gaea.DeveloperInfo, err error) {
	info, err = s.impl.InfoByEmail(s.rpcLogger, email)
	return
}
