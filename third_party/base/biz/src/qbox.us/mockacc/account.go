package mockacc

import (
	"net/http"
	"strconv"
	"strings"

	"qbox.us/api"
	"github.com/qiniu/log.v1"
	"qbox.us/net/httputil"
	"qbox.us/servend/account"
	"qbox.us/servestk"
	"qbox.us/servestk/logh"

	apiaccount "qbox.us/admin_api/account.v2"
)

type Impl interface {
	InfoById(user string) (ui account.UserInfo, passwd string, err error)
	InfoByUid(uid uint32) (ui account.UserInfo, passwd string, err error)
}

const (
	utypeOperator = account.USER_TYPE_ADMIN | account.USER_TYPE_OP
)

func replyError(w http.ResponseWriter, err error) {
	httputil.Error(w, err)
}

func oauth2TokenResponse(user string, passwd string, acc Impl) (resp interface{}, err error) {

	ui, passwd2, err := acc.InfoById(user)
	if err != nil {
		return
	}
	if passwd != passwd2 {
		err = api.EBadToken
		return
	}
	resp = map[string]interface{}{
		"token_type":    "bearer",
		"access_token":  Instance.MakeAccessToken(ui),
		"refresh_token": user + ":" + passwd,
		"expires_in":    360000,
	}
	return
}

func oauth2Token(w http.ResponseWriter, req *http.Request, acc Impl) {

	err := req.ParseForm()
	if err != nil {
		httputil.ReplyWithCode(w, 400)
		return
	}
	{
		grantType := req.Form["grant_type"]
		if grantType == nil {
			httputil.Error(w, api.EInvalidArgs)
			return
		}

		switch grantType[0] {
		case "password":
			user := req.Form["username"]
			passwd := req.Form["password"]
			if user == nil || passwd == nil {
				httputil.Error(w, api.EBadToken)
				return
			}
			resp, err := oauth2TokenResponse(user[0], passwd[0], acc)
			if err != nil {
				httputil.Error(w, api.EBadToken)
				return
			}
			httputil.Reply(w, 200, resp)
			return
		case "refresh_token":
			refreshToken := req.Form["refresh_token"]
			if refreshToken == nil {
				httputil.Error(w, api.EBadToken)
				return
			}
			tokens := strings.SplitN(refreshToken[0], ":", 2)
			resp, err := oauth2TokenResponse(tokens[0], tokens[1], acc)
			if err != nil {
				httputil.Error(w, api.EBadToken)
				return
			}
			httputil.Reply(w, 200, resp)
			return
		}
		httputil.ReplyError(w, "bad auth", 401)
		return
	}
}

func info(w http.ResponseWriter, req *http.Request, acc Impl) {

	user, err := account.GetAuth(Instance, req)

	if err != nil {
		tok := req.FormValue("access_token")
		if tok == "" {
			httputil.Error(w, api.EBadToken)
			return
		}

		user, err = Instance.ParseAccessToken(tok)
		if err != nil {
			httputil.Error(w, err)
			return
		}
	}

	output := apiaccount.Info{
		Uid:       user.Uid,
		Utype:     user.Utype,
		Activated: true,
	}
	httputil.Reply(w, 200, output)
}

func adminInfo(w http.ResponseWriter, req *http.Request, acc Impl) {

	admin, err := account.GetAuth(Instance, req)
	if err != nil || (admin.Utype&utypeOperator) == 0 {
		replyError(w, api.EBadToken)
		return
	}

	var info *account.UserInfo
	req.ParseForm()
	if user1, ok := req.Form["id"]; ok {
		info2, _, err := acc.InfoById(user1[0])
		if err != nil {
			log.Warn("Get UserInfo failed:", err)
			replyError(w, err)
			return
		}
		info = &info2
	}
	if user1, ok := req.Form["uid"]; ok {
		uid, _ := strconv.ParseUint(user1[0], 10, 64)
		info2, _, err := acc.InfoByUid(uint32(uid))
		if err != nil {
			log.Warn("Get UserInfo failed:", err)
			replyError(w, err)
			return
		}
		info = &info2
	}
	if info != nil {
		output := apiaccount.Info{
			Uid:       info.Uid,
			Utype:     info.Utype,
			Activated: true,
		}
		httputil.Reply(w, 200, output)
		return
	}
	replyError(w, api.EInvalidArgs)
}

func RegisterHandlers(mux1 *http.ServeMux, acc Impl) {
	mux := servestk.New(mux1, servestk.SafeHandler, logh.Instance)
	mux.HandleFunc("/user/info", func(w http.ResponseWriter, req *http.Request) { info(w, req, acc) })
	mux.HandleFunc("/admin/user_info", func(w http.ResponseWriter, req *http.Request) { adminInfo(w, req, acc) })
	mux.HandleFunc("/admin/user/info", func(w http.ResponseWriter, req *http.Request) { adminInfo(w, req, acc) })
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, req *http.Request) { oauth2Token(w, req, acc) })
}

func Run(addr string, acc Impl) (err error) {
	mux := http.DefaultServeMux
	RegisterHandlers(mux, acc)
	return http.ListenAndServe(addr, mux)
}
