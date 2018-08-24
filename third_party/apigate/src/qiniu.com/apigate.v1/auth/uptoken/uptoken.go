package uptoken

import (
	"net/http"
	"strings"

	"github.com/qiniu/apigate.v1"
	"qiniu.com/auth/authstub.v1"

	. "code.google.com/p/go.net/context"
	. "qiniu.com/apigate.v1/auth/common"
	. "qiniu.com/auth/account.v1"
	. "qiniu.com/auth/proto.v1"
)

func parseAuth(acc Account, ctx Context, req *http.Request) (user SudoerInfo, data string, err error) {

	if auth1, ok1 := req.Header["Authorization"]; ok1 {
		auth := auth1[0]
		if strings.HasPrefix(auth, "UpToken ") {
			token := auth[8:]
			return ParseDataAuth(acc, ctx, token)
		}
	}
	err = ErrBadToken
	return
}

type uptokenStuber struct {
	Acc                  Account
	AllowFrozenWithAdmin bool
}

func newAuthStuber(pacc Account) (uptokenStuber, uptokenStuber) {

	return uptokenStuber{pacc, false}, uptokenStuber{pacc, true}
}

func (p uptokenStuber) AuthStub(req *http.Request) (ai apigate.AuthInfo, ok bool, err error) {

	if user, data, err := parseAuth(p.Acc, Background(), req); err == nil {
		authp := authstub.FormatStubData(&user, data)
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

// -----------------------------------------------------

type mockAuthStuber struct{}

func (p mockAuthStuber) AuthStub(req *http.Request) (ai apigate.AuthInfo, ok bool, err error) {

	auth := req.Header.Get("Authorization")
	if user, _, err := authstub.ParseStubData(auth); err == nil {
		return makeAuthInfo(user.Utype, user.Sudoer != 0, true), true, nil
	}
	return
}

// -----------------------------------------------------

func InitMock() {

	mock := mockAuthStuber{}
	apigate.RegisterAuthStuber("qbox/uptoken", mock, mock)
}

func Init(pacc Account) (err error) {

	uptoken, uptoken2 := newAuthStuber(pacc)
	apigate.RegisterAuthStuber("qbox/uptoken", uptoken, uptoken2)
	return nil
}
