package qauth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
        "github.com/gin-gonic/gin"
	"qiniu.com/models"
	"net/http"
	"strings"
)

type UserRet struct {
	uid    string
	secret []byte
}

//------------------------------------------------------
// to check ak if ak is username then return is pwd,
// if ak is qiniu accesskey then return user info.
//------------------------------------------------------

func VerifyOldAccessKey(ak string, ret *UserRet) error {
	pwd, err := models.GetPwdByUID(ak)
	if err == nil {
		ret.uid = ak
		ret.secret = []byte(pwd)
		return nil
	}
	return err
}

func VerifyNewAccessKey(ak string, ret *UserRet) error {
	info, err := GetUserInfoByAccessKey(ak)
	if err == nil {
		ret.secret = info.Secret
		acc, err := GetUserInfoByUid(info.Uid)
		if err != nil {
			return err
		}
		if err = models.ValidateUid(acc.Email); err != nil {
			return err
		}
		ret.uid = acc.Email
		return nil
	}
	return err
}

func AccessHandler(ak string) (UserRet, error) {
	resp := UserRet{}
	err1 := VerifyNewAccessKey(ak, &resp)
	if err1 == nil {
		return resp, nil
	}

	err2 := VerifyOldAccessKey(ak, &resp)
	if err2 == nil {
		return resp, nil
	}

	return resp, fmt.Errorf("user are not registered: get user info failed: %v; get pwd failed: %v", err1, err2)
}

func GetUID(ctx *gin.Context) (uid string, err error) {
	val, ok := ctx.Get("uid")
	if !ok {
		return "", fmt.Errorf("uid not set")
	}
        uid = val.(string)
	return uid, nil
}

func SetUID(ctx *gin.Context, uid interface{}) {
	if u, ok := uid.(string); ok {
		ctx.Set("uid", u)
	}
}

func ParsePrefix(tokenStr string) ([]string, error) {

	fields := strings.Fields(tokenStr)
	// token with no prefix
        if len(fields) == 1 {
                return strings.Split(tokenStr, ":"), nil
		// token with prefix
	} else if len(fields) == 2 {
		return strings.Split(fields[1], ":"), nil
	}
	return nil, fmt.Errorf("token prefix format undefined")
}

func ParseRequest(r *http.Request) ([]string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return nil, fmt.Errorf("token doesn't exist")
	}
	parts, err := ParsePrefix(token)
	if err != nil {
		return nil, fmt.Errorf("token string parse failed: %v", err)
	}
	// ak:sig
	if len(parts) != 2 {
		return nil, fmt.Errorf("token body format illegal")
	}
	return parts, nil
}

func sign(req *http.Request, key []byte) string {

	h := hmac.New(sha1.New, key)

	u := req.URL
	data := u.Path
	if u.RawQuery != "" {
		data += "?" + u.RawQuery
	}
	io.WriteString(h, data+"\n")

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func Authenticate(ctx *gin.Context) (err error) {
	token, err := ParseRequest(ctx.Request)
	if err != nil {
		return err
	}
	ak := token[0]
	sig := token[1]

	ret, err := AccessHandler(ak)
	if err != nil {
		return err
	}

	if sig != sign(ctx.Request, ret.secret) {
		return fmt.Errorf("no auth, signature incorrect!")
	}

	SetUID(ctx, ret.uid)
	return nil
}

func ValidateLoginForOldAccount(sign, uid, pwd string) error {
	return models.ValidateLogin(uid, pwd)
}

func ValidateLoginForNewAccount(uid, pwd string) error {
	if err := models.ValidateUid(uid); err != nil {
		return err
	}
	if _, err := GetUserInfo(uid, pwd); err != nil {
		return err
	}
	return nil
}

func ValidateLogin(sign, uid, pwd string) error {
	err1 := ValidateLoginForNewAccount(uid, pwd)
	if err1 == nil {
		return nil
	}
	err2 := ValidateLoginForOldAccount(sign, uid, pwd)
	if err2 == nil {
		return nil
	}
	return fmt.Errorf("login failed: get qiniu account failed: %v; get db info failed: %v", err1, err2)
}
