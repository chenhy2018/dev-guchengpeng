package access

import (
	"net/http"

	"qbox.us/oauth"

	"github.com/teapots/teapot"

	"qbox.us/biz/component/api"
	paccount "qbox.us/biz/component/providers/account"
	"qbox.us/biz/component/sessions"
	"qbox.us/biz/services.v2/account"
	"qbox.us/biz/utils.v2"
)

func LoginRequired(ctx teapot.Context, rw http.ResponseWriter, req *http.Request, log teapot.ReqLogger) {
	pass := false
	defer func() {
		if !pass {
			res := &api.JsonResult{
				Code: api.Unauthorized,
			}
			res.Write(ctx, rw, req)
		}
	}()

	var userInfo *account.UserInfo
	err := ctx.Find(&userInfo, "")
	if err != nil {
		log.Warn(err)
	}
	if userInfo == nil {
		return
	}
	pass = true

	var sess sessions.SessionStore
	err = ctx.Find(&sess, "")
	if err != nil {
		log.Warn(err)
	}
	if sess == nil {
		return
	}
	sessUtype := sess.Get(paccount.LoginUtype).MustString()
	if sessUtype == "" {
		return
	}
	if sessUtype != utils.ToStr(userInfo.UserType) {
		pass = false
		return
	}

	// 如果不是 sso 登录那么就不需要去 sso 验证状态
	if sess.Has(account.SSOLoginState) {
		var adminOAuth *oauth.Transport
		err := ctx.Find(&adminOAuth, "admin")
		if err != nil || adminOAuth == nil {
			log.Error("need admin oauth")
			pass = false
			return
		}
		pass = SSOLoginRequired(sess, userInfo, adminOAuth, log)
		return
	}
}
