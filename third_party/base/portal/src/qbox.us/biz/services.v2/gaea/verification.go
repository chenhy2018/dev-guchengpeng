package gaea

import (
	"net/http"

	"github.com/qiniu/rpc.v1"

	"qbox.us/biz/api/gaea"
	"qbox.us/biz/utils.v2/log"
)

type VerificationService interface {
	Check() (ok bool, err error)
	Consume() (err error)
	CheckWithCookie(cookies []*http.Cookie, types int) (ok bool, err error)
	ConsumeWithCookie(cookies []*http.Cookie, types int) (err error)
}

type verificationService struct {
	logger    log.ReqLogger
	rpcLogger rpc.Logger

	impl *gaea.VerificationService
}

var _ VerificationService = new(verificationService)

func NewVerificationService(host string, t http.RoundTripper, logger log.ReqLogger) VerificationService {
	return &verificationService{
		logger:    logger,
		rpcLogger: log.NewRpcWrapper(logger),
		impl:      gaea.NewVerificationService(host, t),
	}
}

func (s *verificationService) Check() (ok bool, err error) {
	ok, err = s.impl.Check(s.rpcLogger)
	return
}

func (s *verificationService) Consume() (err error) {
	err = s.impl.Consume(s.rpcLogger)
	return
}

func (s *verificationService) CheckWithCookie(cookies []*http.Cookie, types int) (ok bool, err error) {
	ok, err = s.impl.CheckWithCookie(cookies, types, s.rpcLogger)
	return
}

func (s *verificationService) ConsumeWithCookie(cookies []*http.Cookie, types int) (err error) {
	err = s.impl.ConsumeWithCookie(cookies, types, s.rpcLogger)
	return
}
