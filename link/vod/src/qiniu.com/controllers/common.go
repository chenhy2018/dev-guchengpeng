package controllers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/auth"
	"qiniu.com/models"
)

type requestParams struct {
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
}
type userInfo struct {
	uid uint32
	ak  string
	sk  string
}

func getUserInfo(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
	reqUrl := req.URL.String()
	if strings.Contains(reqUrl, "token=") {
		tokenIndex := strings.Index(reqUrl, "&token=")
		req.Header.Set("Authorization", "Qbox ak="+strings.Split(reqUrl[tokenIndex+7:], ":")[0])
	}
	authHeader := req.Header.Get("Authorization")
	auths := strings.Split(authHeader, " ")
	if len(auths) != 2 {
		return nil, errors.New("Parse auth header Failed")
	}
	u, err := url.ParseQuery(auths[1])
	if err != nil {
		return nil, errors.New("Parse auth header Failed")
	}
	ak := u["ak"][0]
	user, err := auth.GetUserInfoFromQconf(xl, ak)
	if err != nil {
		return nil, errors.New("get userinfo error")
	}
	uid := user.Uid
	userInfo := userInfo{
		uid: uid,
		ak:  ak,
		sk:  string(user.Secret[:]),
	}
	return &userInfo, nil
}

func GetUrlWithDownLoadToken(xl *xlog.Logger, domain, fname string, tsExpire int64, userInfo *userInfo) string {
	mac := qbox.NewMac(userInfo.ak, userInfo.sk)
	expireT := time.Now().Add(time.Hour).Unix() + tsExpire
	realUrl := storage.MakePrivateURL(mac, domain, fname, expireT)
	return realUrl
}

func IsAutoCreateUa(xl *xlog.Logger, uid, namespace string) (bool, error) {
	namespaceMod = &models.NamespaceModel{}
	info, err := namespaceMod.GetNamespaceInfo(xl, uid, namespace)
	if err != nil {
		return false, err
	}
	if len(info) == 0 {
		return false, errors.New("can't find namespace")
	}
	return info[0].AutoCreateUa, nil
}

func VerifyToken(xl *xlog.Logger, expire int64, realToken string, req *http.Request) bool {
	if expire == 0 || realToken == "" {
		return false
	}
	if expire < time.Now().Unix() {
		return false
	}
	userInfo, err := getUserInfo(xl, req)
	if err != nil {
		return false
	}
	xl.Infof("info = %v", userInfo)
	url := "http://" + req.Host + req.URL.String()
	tokenIndex := strings.Index(url, "&token=")
	mac := qbox.NewMac(userInfo.ak, userInfo.sk)
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

	params := &requestParams{
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
	}

	return params, nil
}
