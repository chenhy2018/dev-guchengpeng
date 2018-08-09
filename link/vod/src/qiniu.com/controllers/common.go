package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
)

const (
	accessKey = "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ"
	secretKey = "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS"
)

type requestParams struct {
	uid      string
	deviceid string
	from     int64
	to       int64
	expire   int64
	token    string
}

func VerifyAuth(xl *xlog.Logger, req *http.Request) (bool, error) {

	mac := qbox.NewMac(accessKey, secretKey)
	return mac.VerifyCallback(req)
}

func GetUrlWithDownLoadToken(xl *xlog.Logger, domain, fname string) string {
	mac := qbox.NewMac(accessKey, secretKey)
	expireT := time.Now().Add(time.Hour).Unix()
	realUrl := storage.MakePrivateURL(mac, domain, fname, expireT)
	return realUrl
}

func VerifyToken(xl *xlog.Logger, expire int64, realToken, url, uid string) bool {
	if expire == 0 || realToken == "" {
		return false
	}
	if expire < time.Now().Unix() {
		return false
	}
	tokenIndex := strings.Index(url, "&token=")

	mac := qbox.NewMac(uid, uid)
	token := mac.Sign([]byte(url[0:tokenIndex]))
	return token == realToken
}

func ParseRequest(c *gin.Context, xl *xlog.Logger) (*requestParams, error) {
	uid := c.Param("uid")
	deviceid := c.Param("deviceId")
	from := c.Query("from")
	to := c.Query("to")
	expire := c.Query("e")
	token := c.Query("token")

	if strings.Contains(deviceid, ".m3u8") {
		deviceid = strings.TrimRight(deviceid, ".m3u8")
	}
	fromT, err := strconv.ParseInt(from, 10, 32)
	if err != nil {
		return nil, errors.New("Parse from time failed")
	}
	toT, err := strconv.ParseInt(to, 10, 32)
	if err != nil {
		return nil, errors.New("Parse to time failed")
	}
	expireT, err := strconv.ParseInt(expire, 10, 32)
	if err != nil {
		return nil, errors.New("Parse expire time failed")
	}

	params := &requestParams{
		uid:      uid,
		deviceid: deviceid,
		from:     fromT,
		to:       toT,
		expire:   expireT,
		token:    token,
	}

	return params, nil
}
