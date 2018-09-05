package auth

import (
	"qiniu.com/apigate.v1/auth/pandora"
	"qiniu.com/apigate.v1/auth/qbox"
	"qiniu.com/apigate.v1/auth/qiniu"
	"qiniu.com/apigate.v1/auth/quser"
	"qiniu.com/apigate.v1/auth/uptoken"
	"qiniu.com/auth/account.v1"
	"qiniu.com/auth/account.v1/static.v1"
)

// --------------------------------------------------------------------

func InitMock() {

	qbox.InitMock()
	qiniu.InitMock()
	uptoken.InitMock()
	quser.InitMock()
	pandora.InitMock()
}

func Init(pacc account.Account, staticAcc static.Account) (err error) {

	qbox.Init(pacc, staticAcc)
	qiniu.Init(pacc, staticAcc)
	uptoken.Init(pacc)
	quser.Init(pacc)
	pandora.Init(pacc, staticAcc)
	return nil
}

// --------------------------------------------------------------------
