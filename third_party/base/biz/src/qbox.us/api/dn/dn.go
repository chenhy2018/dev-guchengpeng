package dn

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	gourl "net/url"
	"strconv"
	"time"

	"github.com/qiniu/api/auth/digest"
	. "github.com/qiniu/api/conf"
)

type AuthPolicy struct {
	Scope    string `json:"S"`
	Deadline int64  `json:"E"`
}

func MakeAuthTokenString(key, secret string, auth *AuthPolicy) string {
	b, _ := json.Marshal(auth)
	mac := &digest.Mac{key, []byte(secret)}
	return mac.SignWithData(b)
}

type GetPolicyEx struct {
	Expires int64
}

func (r GetPolicyEx) MakeRequest(rawURL string) (string, error) {
	if r.Expires == 0 {
		r.Expires = 3600
	}
	r.Expires += time.Now().Unix()

	u, err := gourl.Parse(rawURL)
	if err != nil {
		return "", err
	}
	rawURL = u.String()

	if u.RawQuery == "" {
		rawURL += "?"
	} else {
		rawURL += "&"
	}
	rawURL += "e=" + strconv.FormatInt(r.Expires, 10)

	// sign for url
	h := hmac.New(sha1.New, []byte(SECRET_KEY))
	io.WriteString(h, rawURL)
	digest := h.Sum(nil)
	token := ACCESS_KEY + ":" + base64.URLEncoding.EncodeToString(digest)

	rawURL += "&token=" + token
	return rawURL, nil
}
