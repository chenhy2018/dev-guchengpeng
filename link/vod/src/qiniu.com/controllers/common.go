package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
	"google.golang.org/grpc"
	"gopkg.in/redis.v5"
	"io"
	"net/http"
	"net/url"
	rs "qbox.us/api/rs.v3"
	"qiniu.com/auth"
	qboxmac "qiniu.com/auth/qboxmac.v1"
	"qiniu.com/models"
	pb "qiniu.com/proto"
	"qiniu.com/system"
	"strconv"
	"strings"
	"time"
)

var (
	namespaceMod     *models.NamespaceModel
	segMod           *models.SegmentKodoModel
	fastForwardClint pb.FastForwardClient
	UaMod            *models.UaModel
	defaultUser      system.UserConf
	c                *redis.Client
)

func Init(conf *system.Configuration, client *redis.Client) {
	defaultUser = conf.UserConf
	namespaceMod = &models.NamespaceModel{}
	namespaceMod.Init()
	segMod = &models.SegmentKodoModel{}
	segMod.Init(defaultUser)
	FFGrpcClientInit(&conf.GrpcConf)
	UaMod = &models.UaModel{}
	UaMod.Init()
	c = client
}

func FFGrpcClientInit(conf *system.GrpcConf) {
	conn, err := grpc.Dial(conf.Addr, grpc.WithInsecure())
	if err != nil {
		fmt.Println("Init gprc failed")
	}
	fastForwardClint = pb.NewFastForwardClient(conn)
}

type requestParams struct {
	uaid      string
	from      int64
	to        int64
	limit     int
	expire    int64
	token     string
	marker    string
	namespace string
	prefix    string
	exact     bool
	speed     int32
	fmt       string
	key       string
	reqid     string
}
type userInfo struct {
	uid string
	ak  string
	sk  string
}

func redisGet(key string) string {
	s := c.Get(key).Val()
	return s
}

func redisSet(xl *xlog.Logger, key, value string) error {
	err := c.Set(key, value, time.Hour).Err()
	if err != nil {
		xl.Errorf("Redis set failed %#v", err)
	}
	return err
}

func checkParams(xl *xlog.Logger, params *requestParams) error {
	if params.to <= params.from {
		xl.Errorf("bad from/to time, from = %v, to = %v", params.from, params.to)
		return fmt.Errorf("bad from/to time, from great or equal than to")
	}

	dayInMilliSec := int64((24 * time.Hour).Seconds() * 1000)
	if (params.to - params.from) > dayInMilliSec {
		xl.Errorf("bad from/to time, from = %v, to = %v", params.from, params.to)
		return fmt.Errorf("bad from/to time, currently we only support playback in 24 hours")
	}
	return nil
}

func getUserInfo(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
	authHeader := req.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "QiniuStub ") {
		return nil, errors.New("Parse Authorization header failed")
	}

	auths := authHeader[10:]

	u, err := url.ParseQuery(auths)
	if err != nil {
		return nil, errors.New("Parse Authorization header Failed")
	}
	user := userInfo{
		ak:  u.Get("ak"),
		uid: u.Get("uid"),
	}
	return &user, nil
}

func HandleToken(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	// TODO verify token if private deploy
	c.Next()
}

func GetUrlWithDownLoadToken(xl *xlog.Logger, domain, fname string, tsExpire int64, mac *qbox.Mac) string {
	expireT := time.Now().Add(time.Hour).Unix() + tsExpire
	realUrl := storage.MakePrivateURL(mac, domain, fname, expireT)
	return realUrl
}

func GetBucketAndDomain(xl *xlog.Logger, uid, namespace string) (string, string, error) {
	if system.HaveDb() == false {
		return namespace, "", nil
	}
	namespaceMod = &models.NamespaceModel{}
	info, err := namespaceMod.GetNamespaceInfo(xl, uid, namespace)
	if err != nil {
		return "", "", err
	}
	if len(info) == 0 {
		return "", "", errors.New("can't find namespace")
	}
	return info[0].Bucket, info[0].Domain, nil
}

func IsAutoCreateUa(xl *xlog.Logger, uid, bucket string) (bool, []models.NamespaceInfo, error) {
	if system.HaveDb() == false {
		return true, []models.NamespaceInfo{}, nil
	}

	namespaceMod = &models.NamespaceModel{}
	info, err := namespaceMod.GetNamespaceByBucket(xl, uid, bucket)
	if err != nil {
		return false, []models.NamespaceInfo{}, err
	}
	if len(info) == 0 {
		return false, []models.NamespaceInfo{}, errors.New("can't find namespace")
	}
	return info[0].AutoCreateUa, info, nil
}
func ParseRequest(c *gin.Context, xl *xlog.Logger) (*requestParams, error) {
	uaid := c.Param("uaid")
	namespace := c.Param("namespace")
	from := c.DefaultQuery("from", "0")
	to := c.DefaultQuery("to", "0")
	expire := c.DefaultQuery("e", "0")
	token := c.Query("token")
	limit := c.DefaultQuery("limit", "1000")
	marker := c.DefaultQuery("marker", "")
	prefix := c.DefaultQuery("prefix", "")
	exact := c.DefaultQuery("exact", "false")
	speed := c.DefaultQuery("speed", "1")
	m3u8Name := c.DefaultQuery("key", "")
	reqid := c.DefaultQuery("reqid", "")
	fmt := c.Query("fmt")

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
	if fmt != "fmp4" && fmt != "flv" && fmt != "" {
		return nil, errors.New("fmt error, it should be flv or fmp4")
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
		prefix:    prefix,
		exact:     exactT,
		speed:     int32(speedT),
		fmt:       fmt,
		key:       m3u8Name,
		reqid:     reqid,
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

func GetNameSpaceInfo(xl *xlog.Logger, bucket, uaid, uid string) (error, int) {

	if system.HaveDb() == false {
		return nil, 0

	}
	isAuto, info, err := IsAutoCreateUa(xl, uid, bucket)
	if err != nil {
		return err, 0
	}

	if isAuto == false {
		model := models.UaModel{}
		r, err := model.GetUaInfo(xl, info[0].Uid, info[0].Space, uaid)
		if err != nil {
			return err, 0
		}
		if len(r) == 0 {
			return fmt.Errorf("Can't find ua info"), 0
		}
	}
	return nil, info[0].Expire
}

func newRsService(user *userInfo, bucket string) (*rs.Service, error) {
	mac := qboxmac.Mac{AccessKey: defaultUser.AccessKey, SecretKey: []byte(defaultUser.SecretKey)}
	var tr http.RoundTripper
	if defaultUser.IsAdmin {
		tr = qboxmac.NewAdminTransport(&mac, user.uid+"/0", nil)
	} else {
		tr = qboxmac.NewTransport(&mac, nil)
	}
	zone, err := models.GetZone(user.ak, bucket)
	if err != nil {
		return nil, err
	}
	upHost := "http://" + zone.SrcUpHosts[0]
	return rs.NewService(tr, zone.GetRsHost(false), zone.GetRsfHost(false), upHost), nil
}

func uploadNewFile(filename, bucket string, data []byte, user *userInfo) error {
	rsService, err := newRsService(user, bucket)
	if err != nil {
		return err
	}
	entry := bucket + ":" + filename
	_, _, err = rsService.Put(entry, "", bytes.NewReader(data), int64(len(data)), "", "", "")
	return err
}

func verifyToken(xl *xlog.Logger, expire int64, realToken string, req *http.Request, user *userInfo) bool {
	if expire == 0 || realToken == "" {
		return false
	}
	if expire < time.Now().Unix() {
		return false
	}
	url := "http://" + req.Host + req.URL.String()
	tokenIndex := strings.Index(url, "&token=")
	mac := &qbox.Mac{AccessKey: user.ak, SecretKey: []byte(user.sk)}
	token := mac.Sign([]byte(url[0:tokenIndex]))
	return token == realToken
}

func getUserInfoByAk(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) {
	reqUrl := req.URL.String()
	if !strings.Contains(reqUrl, "token=") {
		return nil, errors.New("bad url, should contain a token in url"), 401
	}
	token := strings.Split(reqUrl, "&token=")[1]
	ak := strings.Split(token, ":")[0]
	accessInfo, err := auth.GetUserInfoFromQconf(xl, ak)
	if err != nil {
		xl.Errorf("get user info from Qconf failed, err = %#v", err)
		return nil, errors.New("get user info from Qconf failed"), 500
	}
	user := &userInfo{ak: ak,
		sk:  string(accessInfo.Secret[:]),
		uid: fmt.Sprint(accessInfo.Uid)}
	return user, nil, 200
}

func getDomain(xl *xlog.Logger, bucket string, user *userInfo) ([]string, error) {
	zone, err := models.GetZone(user.ak, bucket)
	host := zone.GetApiHost(false)
	url := fmt.Sprintf("%s%s", host+"/v6/domain/list?tbl=", bucket)
	var domain []string
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		xl.Errorf("%#v", err)
		return domain, err
	}
	rpcClient := models.NewRpcClient(user.uid)
	resp, err := rpcClient.Do(context.Background(), request)
	if err != nil {
		return domain, err

	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	for {
		if err := dec.Decode(&domain); err == io.EOF {
			break

		} else if err != nil {
			return domain, err
		}

	}
	if len(domain) == 0 {
		return domain, nil
	}

	return domain, nil
}
