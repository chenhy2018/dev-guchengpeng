package access

import (
	"net/http"

	"github.com/teapots/teapot"

	"qbox.us/biz/component/api"
	"qbox.us/biz/services.v2/account"
)

func EvmUserRequired(ctx teapot.Context, rw http.ResponseWriter, req *http.Request, log teapot.ReqLogger) {
	var userInfo *account.UserInfo

	err := ctx.Find(&userInfo, "")

	// 获取到 userInfo，登录用户获取正确
	if userInfo != nil {
		if userInfo.UserType.IsCCUser() {
			return
		}

		log.Info("not evm user, uid:", userInfo.Uid)
	}

	if err != nil {
		log.Error(err)
	}

	res := &api.JsonResult{Code: api.Forbidden}
	res.Write(ctx, rw, req)
}
