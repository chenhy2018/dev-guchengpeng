package v7

import (
	"net/http"

	"github.com/qiniu/http/httputil.v1"
	"qbox.us/servend/account"
	"qbox.us/servend/proxy_auth"
)

var ErrBadToken = httputil.NewError(401, "bad token")
var ErrUserDisabled = httputil.NewError(401, "user disabled")

// ---------------------------------------------------------------------------

type Env struct {
	W   http.ResponseWriter
	Req *http.Request
	account.UserInfo
}

func (p *Env) OpenEnv(rcvr interface{}, w *http.ResponseWriter, req *http.Request) (err error) {

	return p.openEnv(rcvr, w, req, account.USER_TYPE_SUDOERS|account.USER_TYPE_USERS)
}

func (p *Env) openEnv(
	rcvr interface{}, w *http.ResponseWriter, req *http.Request, userTypeAllows uint32) (err error) {

	p.UserInfo, err = proxy_auth.ParseAuth(req)
	if err != nil || (p.UserInfo.Utype&userTypeAllows) == 0 {
		if (p.UserInfo.Utype & account.USER_TYPE_DISABLED) != 0 {
			return ErrUserDisabled
		}
		return ErrBadToken
	}
	p.W, p.Req = *w, req
	return nil
}

func (p *Env) CloseEnv() {
}

// ---------------------------------------------------------------------------

type AdminEnv struct {
	Env
}

func (p *AdminEnv) OpenEnv(rcvr interface{}, w *http.ResponseWriter, req *http.Request) (err error) {
	return p.Env.openEnv(rcvr, w, req, account.USER_TYPE_ADMIN)
}

func (p *AdminEnv) CloseEnv() {
}

// ---------------------------------------------------------------------------

type SudoerEnv struct {
	Env
}

func (p *SudoerEnv) OpenEnv(rcvr interface{}, w *http.ResponseWriter, req *http.Request) (err error) {
	return p.Env.openEnv(rcvr, w, req, account.USER_TYPE_SUDOERS)
}

func (p *SudoerEnv) CloseEnv() {
}

// ---------------------------------------------------------------------------

type ParentUserEnv struct {
	Env
}

func (p *ParentUserEnv) OpenEnv(rcvr interface{}, w *http.ResponseWriter, req *http.Request) (err error) {
	return p.Env.openEnv(rcvr, w, req, account.USER_TYPE_PARENTUSER)
}

func (p *ParentUserEnv) CloseEnv() {
}

// ---------------------------------------------------------------------------

