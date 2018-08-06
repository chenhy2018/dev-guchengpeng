package account

import (
	"net/http"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/account.v2"

	"qbox.us/biz/utils.v2/log"
)

type AccountService interface {
	UserInfo() (*UserInfo, error)
	ChangePassword(password, newPassword string) error
	Signout(refreshToken string) error

	//subaccount
	CreateChild(email, password string) (uinfo *UserInfo, err error)
	DisableChild(uid uint32, reason string) (uinfo *UserInfo, err error)
	EnableChild(uid uint32) (uinfo *UserInfo, err error)
}

type accountService struct {
	logger log.ReqLogger

	service   account.Service
	rpcLogger rpc.Logger
}

var _ AccountService = new(accountService)

func NewAccountService(host string, userOAuth http.RoundTripper, logger log.ReqLogger) AccountService {
	acc := new(accountService)
	acc.logger = logger

	acc.rpcLogger = log.NewRpcWrapper(logger)
	acc.service = *account.New(host, userOAuth)
	return acc
}

func (a *accountService) UserInfo() (userInfo *UserInfo, err error) {
	accUserInfo, err := a.service.UserInfo(a.rpcLogger)
	if err != nil {
		return
	}
	userInfo = convertFromAccUserInfo(accUserInfo)
	return
}

func (a *accountService) ChangePassword(password, newPassword string) error {
	return a.service.ChangePassword(password, newPassword, a.rpcLogger)
}

func (a *accountService) Signout(refreshToken string) error {
	return a.service.Logout(refreshToken, a.rpcLogger)
}
