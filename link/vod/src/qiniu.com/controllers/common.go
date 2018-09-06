package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/models"
)

const (
	accessKey = "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ"
	secretKey = "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS"
)

type requestParams struct {
	url       string
	uid       string
	uaid      string
	from      int64
	to        int64
	expire    int64
	token     string
	limit     int
	marker    string
	namespace string
	regex     string
	exact     bool
	speed     int32
}

func VerifyAuth(xl *xlog.Logger, req *http.Request) (bool, error) {
	mac := qbox.NewMac(accessKey, secretKey)
	return mac.VerifyCallback(req)
}

func GetUrlWithDownLoadToken(xl *xlog.Logger, domain, fname string, tsExpire int64) string {
	mac := qbox.NewMac(accessKey, secretKey)
	expireT := time.Now().Add(time.Hour).Unix() + tsExpire
	realUrl := storage.MakePrivateURL(mac, domain, fname, expireT)
	fmt.Println(realUrl)
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

	mac := qbox.NewMac(accessKey, secretKey)
	token := mac.Sign([]byte(url[0:tokenIndex]))
	return token == realToken
}

func ParseRequest(c *gin.Context, xl *xlog.Logger) (*requestParams, error) {
	/* uaidT, err := base64.StdEncoding.DecodeString(c.Param("uaid"))
	   if err != nil {
	           return nil, errors.New("decode uaid error")
	   }
	   uaid := string(uaidT)
	*/
	uaid := c.Param("uaid")
	namespace := c.Param("namespace")
	// TODO use ak or body uid.
	uid := c.DefaultQuery("uid", "link")
	from := c.DefaultQuery("from", "0")
	to := c.DefaultQuery("to", "0")
	expire := c.DefaultQuery("e", "0")
	token := c.Query("token")
	limit := c.DefaultQuery("limit", "1000")
	marker := c.DefaultQuery("marker", "")
	regex := c.DefaultQuery("regex", "")
	exact := c.DefaultQuery("exact", "false")
	speed := c.DefaultQuery("speed", "1")

	if strings.Contains(uaid, ".m3u8") {
		uaid = strings.Split(uaid, ".")[0]
	}
	fromT, err := strconv.ParseInt(from, 10, 64)
	if err != nil {
		return nil, errors.New("Parse from time failed")
	}
	toT, err := strconv.ParseInt(to, 10, 64)
	if err != nil {
		return nil, errors.New("Parse to time failed")
	}
	expireT, err := strconv.ParseInt(expire, 10, 64)
	if err != nil {
		return nil, errors.New("Parse expire time failed")
	}
	limitT, err := strconv.ParseInt(limit, 10, 32)
	if err != nil {
		return nil, errors.New("Parse limit failed")
	}
	if limitT > 1000 || limitT <= 0 {
		limitT = 1000
	}
	exactT, err := strconv.ParseBool(exact)
	if err != nil {
		return nil, errors.New("Parse exact failed")
	}

	speedT, err := strconv.ParseInt(speed, 10, 32)
	if err != nil || !validspeed(speedT) {
		return nil, errors.New("Parse speed failed")
	}

	params := &requestParams{
		url:       c.Request.URL.Path,
		uid:       uid,
		uaid:      uaid,
		from:      fromT * 1000,
		to:        toT * 1000,
		expire:    expireT * 1000,
		token:     token,
		limit:     int(limitT),
		marker:    marker,
		namespace: namespace,
		regex:     regex,
		exact:     exactT,
		speed:     int32(speedT),
	}

	return params, nil
}
func validspeed(speed int64) bool {
	s := []int64{1, 2, 4, 8, 16, 32}
	for _, v := range s {
		if speed == v {
			return true
		}
	}
	return false
}
