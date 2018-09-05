package auth

import (
	"github.com/qiniu/xlog.v1"

	"qbox.us/servend/account"
)

type AppAuthorizer interface {
	CanCall(xl *xlog.Logger, user *account.UserInfo, uappName string) error
	CanOp(xl *xlog.Logger, user *account.UserInfo, uappName string) error
}
