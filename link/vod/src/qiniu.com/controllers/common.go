package controllers

import (
	"net/http"

	xlog "github.com/qiniu/xlog.v1"
	"qiniupkg.com/api.v7/auth/qbox"
)

const (
	accessKey = "kevidUP5vchk8Qs9f9cjKo1dH3nscIkQSaVBjYx7"
	secretKey = "KG9zawEhR4axJT0Kgn_VX_046LZxkUZBhcgURAC0"
)

func VerifyAuth(xl *xlog.Logger, req *http.Request) (bool, error) {

	mac := qbox.NewMac(accessKey, secretKey)
	return mac.VerifyCallback(req)
}
