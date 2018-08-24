package quser

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "code.google.com/p/go.net/context"
	"github.com/qiniu/apigate.v1"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"

	. "qiniu.com/auth/account.v1"
	"qiniu.com/auth/authstub.v1"
	"qiniu.com/auth/authutil.v1"
	. "qiniu.com/auth/proto.v1"
	"qiniu.com/auth/quser.v1"
)

var (
	ErrBadToken = httputil.NewError(401, "bad token")
	// 491 means client should apply a new ak&sk from server
	ErrAKExpire    = httputil.NewError(491, "client access key is expired")
	ErrBadParentAK = httputil.NewError(491, "bad token")
)

type quserStuber struct {
	Acc                  Account
	AllowFrozenWithAdmin bool
}

func (p quserStuber) AuthStub(req *http.Request) (ai apigate.AuthInfo, ok bool, err error) {

	if user, err1 := parseAuth(Background(), p.Acc, req); err1 == nil {
		auths := authstub.Format(&user)
		req.Header.Set("Authorization", auths)
		return makeAuthInfo(user.Utype, user.Sudoer != 0, p.AllowFrozenWithAdmin), true, nil

	} else if err1 == ErrBadParentAK || err1 == ErrAKExpire {

		ok, err = true, err1
		return
	}
	return
}

func parseAuth(ctx Context, acc Account, req *http.Request) (user SudoerInfo, err error) {

	if auth := req.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "QUser ") {
			return parseNormalToken(ctx, acc, auth[6:], req)
		}
		if strings.HasPrefix(auth, "QAdmin ") {
			return parseAdminToken(ctx, acc, auth[7:], req)
		}
	}
	err = ErrBadToken
	return
}

func parseNormalToken(ctx Context, acc Account, token string, req *http.Request) (user SudoerInfo, err error) {

	l := strings.SplitN(token, ":", 2)
	if len(l) != 2 {
		err = ErrBadToken
		return
	}
	qak, signexp := l[0], l[1]

	pak, udata, err := parseQak(qak)
	if err != nil {
		return
	}

	info, qsk, err := getQsk(ctx, acc, pak, qak)
	if err != nil {
		return
	}

	sign, err := quser.SignRequest([]byte(qsk), req, qak)
	if err != nil {
		err = errors.Info(err, "parseNormalToken: SignRequest").Detail(err)
		return
	}

	if base64.URLEncoding.EncodeToString(sign) != signexp {
		err = errors.Info(ErrBadToken, "parseAuth: checksum error")
		return
	}

	user.EndUser = udata
	user.Access = pak
	user.Appid = info.Appid
	user.Uid = info.Uid
	user.Utype, err = acc.GetUtype(ctx, user.Uid)
	return
}

func parseAdminToken(ctx Context, acc Account, token string, req *http.Request) (user SudoerInfo, err error) {

	l := strings.SplitN(token, ":", 3)
	if len(l) != 3 {
		err = ErrBadToken
		return
	}

	suinfo, qak, signexp := l[0], l[1], l[2]
	uid, appid, err := authutil.ParseSuInfo(suinfo)
	if err != nil {
		err = errors.Info(ErrBadToken, "ParseSuInfo: ", suinfo).Detail(err)
		return
	}
	pak, udata, err := parseQak(qak)
	if err != nil {
		return
	}

	info, qsk, err := getQsk(ctx, acc, pak, suinfo+":"+qak)
	if err != nil {
		return
	}

	sign, err := quser.SignAdminRequest([]byte(qsk), req, suinfo, qak)
	if err != nil {
		err = errors.Info(err, "parseAdminToken: SignAdminRequest").Detail(err)
		return
	}

	if base64.URLEncoding.EncodeToString(sign) != signexp {
		err = errors.Info(ErrBadToken, "parseAuth: checksum error")
		return
	}

	utype, err := acc.GetUtype(ctx, uid)
	if err != nil {
		err = errors.Info(err, "parseAdminToken: GetUtype - uid:", uid).Detail(err)
		return
	}

	utypeSu, err := acc.GetUtype(ctx, info.Uid)
	if err != nil {
		err = errors.Info(err, "parseAdminToken: GetUtype(su) - uid:", info.Uid).Detail(err)
		return
	}

	user.EndUser = udata
	user.Appid = appid
	user.Uid = uid
	// 防止su到比自己权限更高的用户上
	user.Utype = (utype &^ USER_TYPE_SUDOERS) | (utype & utypeSu & USER_TYPE_SUDOERS)
	user.UtypeSu = utypeSu
	user.Sudoer = info.Uid
	return

}

func getQsk(ctx Context, acc Account, pak, qak string) (info AccessInfo, qsk string, err error) {

	info, err = acc.GetAccessInfo(ctx, pak)
	if err != nil {
		err = errors.Info(ErrBadParentAK, "parseAuth: GetAccessInfo").Detail(err)
		return
	}

	sign := signData(info.Secret, []byte(qak))
	qsk = base64.URLEncoding.EncodeToString(sign)
	return
}

func signData(sk, data []byte) []byte {

	h := hmac.New(sha1.New, sk)
	h.Write(data)
	return h.Sum(nil)
}

func parseQak(qak string) (pak, udata string, err error) {

	l := strings.SplitN(qak, "/", 3)
	if len(l) != 3 {
		err = ErrBadToken
		return
	}

	expire, err := strconv.Atoi(l[1])
	if err != nil {
		err = ErrBadToken
		return
	}

	if time.Now().Unix() > int64(expire) {
		err = ErrAKExpire
		return
	}

	return l[0], l[2], nil
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

// -----------------------------------------------------------

type quserMockAuthStuber struct {
}

func (p quserMockAuthStuber) AuthStub(req *http.Request) (ai apigate.AuthInfo, ok bool, err error) {

	auth := req.Header.Get("Authorization")
	if user, err := authstub.Parse(auth); err == nil {
		return makeAuthInfo(user.Utype, user.Sudoer != 0, true), true, nil
	}
	return
}

// -----------------------------------------------------------

func InitMock() {

	mock := quserMockAuthStuber{}
	apigate.RegisterAuthStuber("qiniu/user", mock, mock)
}

func Init(pacc Account) (err error) {

	quser, quser2 := &quserStuber{pacc, false}, &quserStuber{pacc, true}
	apigate.RegisterAuthStuber("qiniu/user", quser, quser2)
	return nil
}
