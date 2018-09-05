package pandora

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/qiniu/apigate.v1"
	"qiniu.com/auth/account.v1/static.v1"
	"qiniu.com/auth/authstub.v1"

	. "code.google.com/p/go.net/context"
	. "qiniu.com/apigate.v1/auth/common"
	. "qiniu.com/auth/account.v1"
	. "qiniu.com/auth/pandoramac.v1"
	. "qiniu.com/auth/proto.v1"
)

func parseAuth(acc Interface, ctx Context, req *http.Request) (user SudoerInfo, err error) {
	if auth1, ok1 := req.Header["Authorization"]; ok1 {
		defer func() {
			if err == nil {
				req.Header.Set("X-Appid", fmt.Sprintf("%d", user.Uid))
			}
		}()
		auth := auth1[0]
		if strings.HasPrefix(auth, "Pandora ") && strings.Count(auth, ":") == 1 {
			token := auth[8:]
			return ParseNormalAuth(DefaultRequestSigner, acc, ctx, token, req)
		} else if strings.HasPrefix(auth, "Pandora ") && strings.Count(auth, ":") == 2 {
			token := auth[8:]
			return parseTokenAuth(acc, ctx, token, req)
		} else if strings.HasPrefix(auth, "PandoraAdmin ") {
			token := auth[13:]
			return ParseAdminAuth(DefaultRequestSigner, acc, ctx, token, req)
		}
	}
	err = ErrBadToken
	return
}

type qiniuAuthStuber struct {
	Acc                  Interface
	AllowFrozenWithAdmin bool
}

func newAuthStuber(pacc Interface) (qiniuAuthStuber, qiniuAuthStuber) {
	return qiniuAuthStuber{pacc, false}, qiniuAuthStuber{pacc, true}
}

// Qiniu <AK>:<MacSign>
// Qiniu <AK>:<Token>:<UnsignedToken>
// QiniuAdmin <SuInfo>:<AdminAK>:<MacSign>
//
func (p qiniuAuthStuber) AuthStub(req *http.Request) (ai apigate.AuthInfo, ok bool, err error) {
	var user SudoerInfo
	if user, err = parseAuth(p.Acc, Background(), req); err == nil {
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

type mockAuthStuber struct {
}

func (p mockAuthStuber) AuthStub(req *http.Request) (ai apigate.AuthInfo, ok bool, err error) {
	auth := req.Header.Get("Authorization")
	if user, err := authstub.Parse(auth); err == nil {
		return makeAuthInfo(user.Utype, user.Sudoer != 0, true), true, nil
	}
	return
}

func InitMock() {
	mock := mockAuthStuber{}
	apigate.RegisterAuthStuber("pandora/mac", mock, mock)
}

func Init(pacc Account, staticAcc static.Account) (err error) {
	qbox, qbox2 := newAuthStuber(pacc)
	apigate.RegisterAuthStuber("pandora/mac", qbox, qbox2)
	qboxStatic, qboxStatic2 := newAuthStuber(staticAcc)
	apigate.RegisterAuthStuber("pandora/static", qboxStatic, qboxStatic2)

	return nil
}
