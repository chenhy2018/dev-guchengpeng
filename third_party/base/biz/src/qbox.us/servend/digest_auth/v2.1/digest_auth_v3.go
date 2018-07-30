package v21

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

	"qbox.us/api/qconf/akg"
	"qbox.us/api/qconf/uidg"

	qconf "qbox.us/qconf/qconfapi"
)

type Config struct {
	Qconf  *qconf.Client `json:"-"`
	Qconfg qconf.Config  `json:"qconfg"`
}

// ---------------------------------------------------------------------------------------

type Service struct {
	qconfg *qconf.Client
	account.Interface
}

func New(cfg *Config, acc account.Interface) (r *Service) {

	r = &Service{Interface: acc}
	if cfg.Qconf == nil {
		r.qconfg = qconf.New(&cfg.Qconfg)
	} else {
		r.qconfg = cfg.Qconf
	}
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

func (r *Service) getUtype(uid uint32) uint32 {

	utype, err := uidg.Client{r.qconfg}.GetUtype(nil, uid)
	if err != nil {
		log.Warn("digest_auth.getUtype failed:", err)
		return account.USER_TYPE_STDUSER
	}
	return utype
}

func (r *Service) GetSecret(key string) (secret []byte, ok bool) {

	info, err := akg.Client{r.qconfg}.Get(nil, key)
	if err != nil {
		log.Warn("digest_auth.GetSecret failed:", err)
		return
	}
	return info.Secret, true
}

func (r *Service) DigestAuthEx(tempToken string) (user account.UserInfo, data string, err error) {

	parts := strings.SplitN(tempToken, ":", 3)
	if len(parts) != 3 {
		err = errors.Info(api.EBadToken, "DigestAuthEx", tempToken)
		return
	}
	key := parts[0]

	info, err := akg.Client{r.qconfg}.Get(nil, key)
	if err != nil {
		err = errors.Info(api.EBadToken, "DigestAuthEx: accessKey not found").Detail(err)
		return
	}

	h := hmac.New(sha1.New, info.Secret)
	io.WriteString(h, parts[2])
	digest := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if parts[1] != digest {
		err = errors.Info(api.EBadToken, "DigestAuthEx.auth: checksum error")
		return
	}

	user.Appid = uint32(info.Appid)
	user.Uid = info.Uid
	user.Utype = r.getUtype(user.Uid)
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

	info, err := akg.Client{r.qconfg}.Get(nil, key)
	if err != nil {
		err = errors.Info(api.EBadToken, "DigestAuth: accessKey not found").Detail(err)
		return
	}

	sumCalc, err := digest_auth.Checksum(req, info.Secret, incBody(req))
	if err != nil {
		err = errors.Info(err, "DigestAuth").Detail(err)
		return
	}

	sumExp := token[pos+1:]
	if sumCalc != sumExp {
		err = errors.Info(api.EBadToken, "DigestAuth: checksum error", sumCalc, sumExp)
		return
	}

	user.Appid = uint32(info.Appid)
	user.Uid = info.Uid
	user.Utype = r.getUtype(user.Uid)
	return
}

// ---------------------------------------------------------------------------------------
