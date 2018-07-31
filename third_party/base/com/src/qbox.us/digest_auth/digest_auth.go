package digest_auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/http"
	"qbox.us/errors"
	"qbox.us/net/httputil/bytes"
)

// ---------------------------------------------------------------------------------------

func Checksum(req *http.Request, key []byte, incbody bool) (digest string, err error) {

	h := hmac.New(sha1.New, key)

	u := req.URL
	data := u.Path
	if u.RawQuery != "" {
		data += "?" + u.RawQuery
	}
	io.WriteString(h, data+"\n")

	if incbody {
		seekable, err2 := bytes.Seekable(req)
		if err2 != nil {
			err = errors.Info(err2, "digest_auth.Checksum").Detail(err)
			return
		}
		h.Write(seekable.Bytes())
	}
	digest = base64.URLEncoding.EncodeToString(h.Sum(nil))
	return
}

// ---------------------------------------------------------------------------------------

// Transport implements http.RoundTripper. When configured with a valid
// Config and Token it can be used to make authenticated HTTP requests.
//
//	c := NewClient(token, nil)
//	r, _, err := c.Get("http://example.org/url/requiring/auth")
//
type Transport struct {
	key    string
	secret []byte

	// Transport is the HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	// (It should never be an oauth.Transport.)
	transport http.RoundTripper
}

func incBody(req *http.Request) bool {
	if req.Body == nil {
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

// RoundTrip executes a single HTTP transaction using the Transport's
// Token as authorization headers.
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	digest, err := Checksum(req, t.secret, incBody(req))
	if err != nil {
		err = errors.Info(err, "digest_auth.Transport.RoundTrip").Detail(err)
		return
	}
	token := t.key + ":" + digest
	req.Header.Set("Authorization", "QBox "+token)
	return t.transport.RoundTrip(req)
}

func (t *Transport) NestedObject() interface{} {
	return t.transport
}

func NewTransport(key, secret string, transport http.RoundTripper) *Transport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &Transport{key, []byte(secret), transport}
}

func NewClient(key, secret string, transport http.RoundTripper) *http.Client {
	t := NewTransport(key, secret, transport)
	return &http.Client{Transport: t}
}

// ---------------------------------------------------------------------------------------
