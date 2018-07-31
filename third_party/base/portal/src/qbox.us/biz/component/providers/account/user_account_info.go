package account

import (
	"github.com/teapots/teapot"

	"qbox.us/biz/services.v2/account"
)

func UserAccountInfo() interface{} {
	return func(accountService account.AccountService, log teapot.ReqLogger) *account.UserInfo {
		userInfo, err := accountService.UserInfo()
		if err != nil {
			log.Info("account err", err)
			return nil
		}
		log.Info("current uid:", userInfo.Uid)
		return userInfo
	}
}
