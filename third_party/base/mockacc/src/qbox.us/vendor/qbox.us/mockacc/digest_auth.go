package mockacc

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"qbox.us/api"
	"qbox.us/digest_auth"
	"qbox.us/errors"
	"qbox.us/servend/account"
)

// ---------------------------------------------------------------------------------------

type authInfo struct {
	secret []byte
	appId  uint32
	uid    uint32
	utype  uint32
}

type Account struct {
	auth map[string]*authInfo
	Base
}

var Instance Account
var Parser = account.OldParserEx{Instance}

// ---------------------------------------------------------------------------------------

func New(sa SimpleAccount) (r Account) {

	r.auth = make(map[string]*authInfo)
	for _, ui := range sa {
		r.auth[ui.AccessKey] = &authInfo{
			secret: []byte(ui.SecretKey),
			appId:  ui.Appid,
			uid:    ui.Uid,
			utype:  ui.Utype,
		}
	}
	return
}

func NewParser(sa SimpleAccount) account.OldParserEx {

	return account.OldParserEx{New(sa)}
}

var defaultImpl = New(GetSa())

// ---------------------------------------------------------------------------------------

func incBody(req *http.Request) bool {

	if req.ContentLength == 0 {
		return false
	}

	if ct, ok := req.Header["Content-Type"]; ok {
		switch ct[0] {
		case "application/x-www-form-urlencoded":
			return true
		}
	}
	return false
}

func (r Account) GetSecret(key string) (secret []byte, ok bool) {

	if r.auth == nil {
		r = defaultImpl
	}

	info, ok := r.auth[key]
	if ok {
		secret = info.secret
	}
	return
}

func (r Account) DigestAuthEx(tempToken string) (user account.UserInfo, data string, err error) {

	if r.auth == nil {
		r = defaultImpl
	}

	parts := strings.SplitN(tempToken, ":", 3)
	if len(parts) != 3 {
		err = errors.Info(api.EBadToken, "DigestAuthEx", tempToken)
		return
	}
	key := parts[0]

	info, ok := r.auth[key]
	if !ok {
		err = errors.Info(api.EBadToken, "DigestAuthEx: accessKey not found")
		return
	}

	h := hmac.New(sha1.New, info.secret)
	io.WriteString(h, parts[2])
	digest := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if parts[1] != digest {
		err = errors.Info(api.EBadToken, "DigestAuthEx.auth: checksum error")
		return
	}

	user.Appid = info.appId
	user.Uid = info.uid
	user.Utype = info.utype
	data = parts[2]
	return
}

func (r Account) DigestAuth(token string, req *http.Request) (user account.UserInfo, err error) {

	if r.auth == nil {
		r = defaultImpl
	}

	pos := strings.Index(token, ":")
	if pos == -1 {
		err = errors.Info(api.EBadToken, "DigestAuth", token)
		return
	}

	key := token[:pos]

	info, ok := r.auth[key]
	if !ok {
		err = errors.Info(api.EBadToken, "DigestAuth: accessKey not found")
		return
	}

	sumCalc, err := digest_auth.Checksum(req, info.secret, incBody(req))
	if err != nil {
		err = errors.Info(err, "DigestAuth").Detail(err)
		return
	}

	sumExp := token[pos+1:]
	if sumCalc != sumExp {
		err = errors.Info(api.EBadToken, "DigestAuth: checksum error", sumCalc, sumExp)
		return
	}

	user.Appid = info.appId
	user.Uid = info.uid
	user.Utype = info.utype
	return
}

// ---------------------------------------------------------------------------------------
