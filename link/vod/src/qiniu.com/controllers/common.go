package controllers

import (
	"errors"
	"fmt"
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
	"qiniu.com/system"
)

type requestParams struct {
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
type userInfo struct {
	uid uint32
	ak  string
	sk  string
}

const TOKEN_HEADER_LEN = 5

var (
	localInfo userInfo
)

func SetUserInfo(ak, sk string) {
	localInfo.uid = 0
	localInfo.ak = ak
	localInfo.sk = sk
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
	var info userInfo
	if system.HaveQconf() == true {
		user, err := auth.GetUserInfoFromQconf(xl, ak)
		if err != nil {
			return nil, errors.New("get userinfo error")
		}
		uid := user.Uid
		info = userInfo{
			uid: uid,
			ak:  ak,
			sk:  string(user.Secret[:]),
		}
	} else {
		if ak == localInfo.ak {
			info = userInfo{
				uid: localInfo.uid,
				ak:  localInfo.ak,
				sk:  localInfo.sk,
			}
		}
	}
	xl.Infof("info = %v", info)
	return &info, nil
}

func HandleToken(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	token := c.Request.Header.Get("Authorization")
	if len(token) <= TOKEN_HEADER_LEN && !strings.Contains(c.Request.URL.String(), "playback") {
		xl.Errorf("bad token")
		c.AbortWithStatusJSON(401, gin.H{
			"error": "bad token",
		})
		return
	}
	if !strings.Contains(c.Request.URL.String(), "playback") {
		c.Request.Header.Set("Authorization", "Qbox ak="+strings.Split(token[TOKEN_HEADER_LEN:], ":")[0])
		fmt.Println(strings.Split(token[TOKEN_HEADER_LEN:], ":")[0])
	}
	c.Next()
}

func getUid(uid uint32) string {
	return strconv.Itoa(int(uid))
}

func GetUrlWithDownLoadToken(xl *xlog.Logger, domain, fname string, tsExpire int64, userInfo *userInfo) string {
	mac := qbox.NewMac(userInfo.ak, userInfo.sk)
	expireT := time.Now().Add(time.Hour).Unix() + tsExpire
	realUrl := storage.MakePrivateURL(mac, domain, fname, expireT)
	return realUrl
}

func GetBucket(xl *xlog.Logger, uid, namespace string) (string, error) {
	if system.HaveDb() == false {
		return namespace, nil
	}
	namespaceMod = &models.NamespaceModel{}
	info, err := namespaceMod.GetNamespaceInfo(xl, uid, namespace)
	if err != nil {
		return "", err
	}
	if len(info) == 0 {
		return "", errors.New("can't find namespace")
	}
	return info[0].Bucket, nil
}

func IsAutoCreateUa(xl *xlog.Logger, bucket string) (bool, []models.NamespaceInfo, error) {
	if system.HaveDb() == false {
		return true, []models.NamespaceInfo{}, nil
	}

	namespaceMod = &models.NamespaceModel{}
	info, err := namespaceMod.GetNamespaceByBucket(xl, bucket)
	if err != nil {
		return false, []models.NamespaceInfo{}, err
	}
	if len(info) == 0 {
		return false, []models.NamespaceInfo{}, errors.New("can't find namespace")
	}
	return info[0].AutoCreateUa, info, nil
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
	if err != nil || !isValidSpeed(speedT) {
		return nil, errors.New("Parse speed failed")
	}

	params := &requestParams{
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
func isValidSpeed(speed int64) bool {
	s := []int64{1, 2, 4, 8, 16, 32}
	for _, v := range s {
		if speed == v {
			return true
		}
	}
	return false
}

func GetNameSpaceInfo(xl *xlog.Logger, bucket, uaid string) (error, int) {

	if system.HaveDb() == false {
		return nil, 0
	}

	isAuto, info, err := IsAutoCreateUa(xl, bucket)
	if err != nil {
		return err, 0
	}

	if isAuto == false {
		model := models.UaModel{}
		r, err := model.GetUaInfo(xl, info[0].Space, uaid)
		if err != nil {
			return err, 0
		}
		if len(r) == 0 {
			return fmt.Errorf("Can't find ua info"), 0
		}
	}
	return nil, info[0].Expire
}
