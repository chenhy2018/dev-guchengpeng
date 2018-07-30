// 这个包已经迁移到qiniu.com/auth/account.v1/static.v1
package admin

import (
	"io"
	"strings"

	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"

	"github.com/qiniu/log.v1"
	"qbox.us/api"
	"qbox.us/digest_auth"
	"qbox.us/errors"
	"qbox.us/servend/account"
)

type Info struct {
	Access string `json:"access"`
	Secret string `json:"secret"`
	Uid    uint32 `json:"uid"`
	Utype  uint32 `json:"utype"`
}

type Config struct {
	Users []Info `json:"users"`
}

// ---------------------------------------------------------------------------------------

type Service struct {
	users map[string]Info // access -> Info
	account.Interface
}

func New(cfg *Config, acc account.Interface) (r *Service) {

	users := make(map[string]Info, len(cfg.Users))
	for _, u := range cfg.Users {
		users[u.Access] = u
	}
	r = &Service{users: users, Interface: acc}
	return
}

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

func (r *Service) GetSecret(key string) (secret []byte, ok bool) {

	log.Debug("GetSecret with", key)
	info, ok := r.users[key]
	if !ok {
		log.Warn("digest_auth.GetSecret failed:", key)
		return
	}
	return []byte(info.Secret), true
}

func (r *Service) DigestAuthEx(tempToken string) (user account.UserInfo, data string, err error) {

	parts := strings.SplitN(tempToken, ":", 3)
	if len(parts) != 3 {
		err = errors.Info(api.EBadToken, "DigestAuthEx", tempToken)
		return
	}
	key := parts[0]

	log.Debug("DigestAuthEx with", key)
	info, ok := r.users[key]
	if !ok {
		err = errors.Info(api.EBadToken, "DigestAuthEx: accessKey not found", key)
		return
	}

	h := hmac.New(sha1.New, []byte(info.Secret))
	io.WriteString(h, parts[2])
	digest := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if parts[1] != digest {
		err = errors.Info(api.EBadToken, "DigestAuthEx.auth: checksum error")
		return
	}

	user.Uid = info.Uid
	user.Utype = info.Utype
	data = parts[2]
	return
}

func (r *Service) DigestAuth(token string, req *http.Request) (user account.UserInfo, err error) {

	pos := strings.Index(token, ":")
	if pos == -1 {
		err = errors.Info(api.EBadToken, "DigestAuth", token)
		return
	}

	key := token[:pos]

	log.Debug("DigestAuth with", key)
	info, ok := r.users[key]
	if !ok {
		err = errors.Info(api.EBadToken, "DigestAuth: accessKey not found", key)
		return
	}

	sumCalc, err := digest_auth.Checksum(req, []byte(info.Secret), incBody(req))
	if err != nil {
		err = errors.Info(err, "DigestAuth").Detail(err)
		return
	}

	sumExp := token[pos+1:]
	if sumCalc != sumExp {
		err = errors.Info(api.EBadToken, "DigestAuth: checksum error", sumCalc, sumExp)
		return
	}

	user.Uid = info.Uid
	user.Utype = info.Utype
	return
}

// ---------------------------------------------------------------------------------------
