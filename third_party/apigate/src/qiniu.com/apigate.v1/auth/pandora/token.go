package pandora

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"

	. "code.google.com/p/go.net/context"
	. "qiniu.com/apigate.v1/auth/common"
	. "qiniu.com/auth/pandoramac.v1"
	. "qiniu.com/auth/proto.v1"
)

var (
	ErrTokenExpired         = httputil.NewError(401, "The token contains in the request is expired")
	ErrUnmatchedMethod      = httputil.NewError(401, "The method is different to the one specified in token")
	ErrUnmatchedHeaders     = httputil.NewError(401, "The X-Qiniu- headers is different to the one specified in token")
	ErrUnmatchedResource    = httputil.NewError(401, "The resource is different to the one specified in token")
	ErrUnmatchedContentType = httputil.NewError(401, "The contentType is different to the one specified in token")
	ErrUnmatchedContentMD5  = httputil.NewError(401, "The contentMD5 is different to the one specified in token")
)

type token struct {
	Resource    string `json:"resource"`
	Expires     int64  `json:"expires"`
	ContentType string `json:"contentType"`
	ContentMD5  string `json:"contentMD5"`
	Method      string `json:"method"`
	Headers     string `json:"headers"`
}

func (t *token) String() string {
	return fmt.Sprintf("resource=%s, expires=%d, contentType=%s, contentMD5=%s, method=%s, headers=%s",
		t.Resource,
		t.Expires,
		t.ContentType,
		t.ContentMD5,
		t.Method,
		t.Headers)
}

func parseTokenAuth(acc Interface, ctx Context, auth string, req *http.Request) (user SudoerInfo, err error) {
	auths := strings.Split(auth, ":")
	if len(auths) != 3 {
		err = ErrBadToken
		log.Errorf("split auth to 3 pieces fail: %s", auth)
		return
	}
	user, data, err := ParseDataAuth(acc, ctx, auth)
	if err != nil {
		log.Error("ParseDataAuth fail: ", err)
		return
	}
	buf, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		err = ErrBadToken
		return
	}

	var t token
	err = json.Unmarshal(buf, &t)
	if err != nil {
		log.Errorf("unmarshal fail, data: %s, buf: %s, err: %s", data, string(buf), err)
		err = ErrBadToken
		return
	}
	err = checkToken(&t, req)
	return
}

func checkToken(t *token, req *http.Request) (err error) {
	defer func() {
		if err != nil {
			log.Errorf("token: %s, err: %v", t.String(), err)
		}
	}()

	if time.Now().Unix() > t.Expires {
		return ErrTokenExpired
	}
	if req.Method != t.Method {
		return ErrUnmatchedMethod
	}
	if req.Header.Get("Content-Type") != t.ContentType {
		return ErrUnmatchedContentType
	}
	if req.Header.Get("Content-MD5") != t.ContentMD5 {
		return ErrUnmatchedContentMD5
	}

	if SignQiniuHeaderValues(req.Header) != t.Headers {
		return ErrUnmatchedHeaders
	}
	if SignQiniuResourceValues(req.URL) != t.Resource {
		return ErrUnmatchedResource
	}

	return
}
