package qbox

import (
	"net/http"
	"strings"

	"github.com/qiniu/apigate.v1"
	qaccount "qbox.us/account"
	authstub "qiniu.com/auth/authstub.v1"

	. "code.google.com/p/go.net/context"
	. "qiniu.com/apigate.v1/auth/common"
	. "qiniu.com/auth/account.v1"
	"qiniu.com/auth/account.v1/static.v1"
	. "qiniu.com/auth/proto.v1"
	. "qiniu.com/auth/qboxmac.v1"
)

// --------------------------------------------------------------------

var (
	g_acc qaccount.Account
)

func parseAuth(
	acc Interface, ctx Context, req *http.Request, bearAllow bool) (user SudoerInfo, err error) {

	if auth1, ok1 := req.Header["Authorization"]; ok1 {
		auth := auth1[0]
		if strings.HasPrefix(auth, "QBox ") {
			token := auth[5:]
			return ParseNormalAuth(DefaultRequestSigner, acc, ctx, token, req)
		} else if strings.HasPrefix(auth, "QBoxAdmin ") {
			token := auth[10:]
			return ParseAdminAuth(DefaultRequestSigner, acc, ctx, token, req)
		} else if bearAllow && strings.HasPrefix(auth, "Bearer ") {
			token := auth[7:]
			old, err1 := g_acc.ParseAccessToken(token)
			user.Uid = old.Uid
			user.Utype = old.Utype
			user.Appid = uint64(old.Appid)
			err = err1
			return
		}
	}
	err = ErrBadToken
	return
}

// --------------------------------------------------------------------

type qboxAuthStuber struct {
	Acc                  Interface
	AllowFrozenWithAdmin bool
}

func newAuthStuber(pacc Interface) (qboxAuthStuber, qboxAuthStuber) {

	return qboxAuthStuber{pacc, false}, qboxAuthStuber{pacc, true}
}

// QBox <AK>:<MacSign>
// QBoxAdmin <SuInfo>:<AdminAK>:<MacSign>
//
func (p qboxAuthStuber) AuthStub(req *http.Request) (ai apigate.AuthInfo, ok bool, err error) {

	if user, err := parseAuth(p.Acc, Background(), req, false); err == nil {
		authp := authstub.Format(&user)
		req.Header.Set("Authorization", authp)
		return makeAuthInfo(user.Utype, user.Sudoer != 0, p.AllowFrozenWithAdmin), true, nil
	}
	return
}

func makeAuthInfo(utype uint32, su, allowAdminFrozen bool) (ai apigate.AuthInfo) {
	ai.Su = su
	ai.Utype = uint64(utype)
	switch {
	case allowAdminFrozen && su:
	case (ai.Utype & USER_TYPE_DISABLED) == 0:
	default:
		ai.Utype = USER_TYPE_DISABLED
	}
	return
}

// --------------------------------------------------------------------

type bearAuthStuber struct {
	Acc Interface
}

func newBearAuthStuber(pacc Interface) bearAuthStuber {

	return bearAuthStuber{pacc}
}

func (p bearAuthStuber) AuthStub(req *http.Request) (ai apigate.AuthInfo, ok bool, err error) {

	if user, err := parseAuth(p.Acc, Background(), req, true); err == nil {
		authp := authstub.Format(&user)
		req.Header.Set("Authorization", authp)
		return makeAuthInfo(user.Utype, user.Sudoer != 0, true), true, nil
	}
	return
}

// --------------------------------------------------------------------

type qboxMockAuthStuber struct {
}

func (p qboxMockAuthStuber) AuthStub(req *http.Request) (ai apigate.AuthInfo, ok bool, err error) {

	auth := req.Header.Get("Authorization")
	if user, err := authstub.Parse(auth); err == nil {
		return makeAuthInfo(user.Utype, user.Sudoer != 0, true), true, nil
	}
	return
}

// --------------------------------------------------------------------

func InitMock() {

	mock := qboxMockAuthStuber{}
	apigate.RegisterAuthStuber("qbox/mac", mock, mock)
	apigate.RegisterAuthStuber("qbox/bearer", mock, mock)
	apigate.RegisterAuthStuber("qbox/macbearer", mock, mock)
}

func Init(pacc Account, staticAcc static.Account) (err error) {

	qbox, qbox2 := newAuthStuber(pacc)
	apigate.RegisterAuthStuber("qbox/mac", qbox, qbox2)

	qboxStatic, qboxStatic2 := newAuthStuber(staticAcc)
	apigate.RegisterAuthStuber("qbox/static", qboxStatic, qboxStatic2)

	bear := newBearAuthStuber(pacc)
	apigate.RegisterAuthStuber("qbox/bearer", bear, bear)
	apigate.RegisterAuthStuber("qbox/macbearer", bear, bear)
	return nil
}

// --------------------------------------------------------------------
