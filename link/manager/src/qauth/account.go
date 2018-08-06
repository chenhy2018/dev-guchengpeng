package qauth

import (
	"fmt"
	"github.com/qiniu/xlog.v1"
	. "qiniu.com/qauth/config"

	admin "qbox.us/admin_api/account.v2"
	account "qbox.us/api/account.v2"
)

// get qiniu account info by email and password
func GetUserInfo(email, password string) (acc account.UserInfo, err error) {
	l := xlog.NewDummy()
	if QTransport == nil {
		return acc, fmt.Errorf("cannot get transport for admin")
	}
	_, _, err = QTransport.ExchangeByPassword(email, password)
	if err != nil {
		return acc, err
	}
	return QConfAccount.UserInfo(l)
}

// get qiniu account info by uid(formatted as time stamp)
func GetUserInfoByUid(uid uint32) (acc admin.Info, err error) {
	l := xlog.NewDummy()
	if QTransport == nil {
		return acc, fmt.Errorf("cannot get transport for account")
	}
	_, _, err = QTransport.ExchangeByPassword(
		Pubconf.GetSection(CONF_ACCOUNT).String(CLIENT_INIT_USER),
		Pubconf.GetSection(CONF_ACCOUNT).String(CLIENT_INIT_PASSWORD),
	)
	if err != nil {
		return acc, err
	}
	return QConfAdmin.UserInfoByUid(uid, l)
}
