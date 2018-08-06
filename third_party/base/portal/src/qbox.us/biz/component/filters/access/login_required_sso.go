package access

import (
	"github.com/teapots/teapot"
	"qbox.us/biz/component/sessions"
	"qbox.us/biz/services.v2/account"
	l "qbox.us/biz/utils.v2/log"
	"qbox.us/oauth"
)

func SSOLoginRequired(sess sessions.SessionStore, uinfo *account.UserInfo, adminOAuth *oauth.Transport,
	log teapot.ReqLogger) (pass bool) {
	state := sess.Get(account.SSOLoginState).MustString()
	ssoInfo := sess.Get(account.SSOLoginInfo).Value().(map[string]interface{})
	host := ssoInfo["host"].(string)
	clientId := ssoInfo["clientid"].(string)

	if len(host) == 0 || len(clientId) == 0 {
		return
	}

	ssoSvc := account.NewSSOService(host, clientId, adminOAuth)
	if state == account.SSOLoginStateToken {
		return verifyToken(sess, ssoSvc, log, uinfo)
	}
	return
}

func verifyToken(sess sessions.SessionStore, svc *account.SSOService, log teapot.ReqLogger,
	uinfo *account.UserInfo) (pass bool) {
	loginToken := sess.Get(account.SSOLoginToken).MustString()
	if len(loginToken) == 0 {
		log.Error("sso login token is empty")
		return
	}

	info, err := svc.LoginRequired(l.NewRpcWrapper(log), loginToken)
	if err != nil {
		log.Error("sso LoginRequired failed", err)
		return
	}

	if info.Uid != uinfo.Uid {
		return
	}

	pass = true
	return
}
