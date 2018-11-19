package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/auth"
	"qiniu.com/m3u8"
	"qiniu.com/models"
	pb "qiniu.com/proto"
)

// temp solution for fastforward, after i-frame verison
// this code will delete from here
func GetFastForward(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	userInfo, err, code := getUserInfoByAk(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed %v", err)
		c.JSON(code, gin.H{"error": err.Error()})
		return

	}
	if !verifyToken(xl, params.expire, params.token, c.Request, userInfo) {
		xl.Errorf("verify token falied")
		c.JSON(401, gin.H{
			"error": "bad token",
		})
		return
	}

	err = checkParams(xl, params)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
	}

	xl.Infof("uid = %v, uaid = %v, from = %v, to = %v", userInfo.uid, params.uaid, params.from, params.to)

	info, err := UaMod.GetUaInfo(xl, userInfo.uid, params.namespace, params.uaid)
	if err != nil || len(info) == 0 {
		xl.Errorf("get ua info failed, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "ua is not correct",
		})
		return
	}
	xl.Infof("info[0].Namespace %v", info[0].Namespace)
	namespace := info[0].Namespace
	bucket, err := GetBucket(xl, userInfo.uid, namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	// get ts list from kodo
	playlist, err, code := getPlaybackList(xl, params, bucket, userInfo)
	if err != nil {
		xl.Errorf("get playback list error, error = %#v", err.Error())
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	// make m3u8 file name with "uaid + from + end.m3u8" if user not given
	fileName := params.key
	if fileName == "" {
		from := strconv.FormatInt(params.from, 10)
		end := strconv.FormatInt(params.to, 10)
		fileName = params.uaid + from + end + ".m3u8"
	}

	// upload new m3u8 file to kodo bucket
	m3u8File := m3u8.Mkm3u8(playlist, xl)
	err = uploadNewFile(fileName, bucket, []byte(m3u8File), userInfo)
	if err != nil {
		xl.Errorf("uplaod New m3u8 file failed, error = %#v", err.Error())
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}

	if err = getFastForwardStream(xl, params, c, userInfo, bucket, fileName); err != nil {
		xl.Errorf("get fastforward stream error , error = %v", err.Error())
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}
}

func getFastForwardStream(xl *xlog.Logger, params *requestParams, c *gin.Context, user *userInfo, bucket, fileName string) error {
	// remove speed fmt from url
	url := c.Request.URL
	query := url.Query()
	query.Del("fmt")
	query.Del("speed")
	query.Del("token")
	query.Del("e")
	req := new(pb.FastForwardInfo)
	domain, err := getDomain(xl, bucket, user)
	if err != nil {
		xl.Errorf("get domain error, err = %#v", err)
		return err
	}

	req.Url = getDownUrlWithPm3u8(domain, fileName, user)
	req.Speed = params.speed
	req.Fmt = params.fmt
	ctx, cancel := context.WithCancel(context.Background())
	r, err := fastForwardClint.GetTsStream(ctx, req)
	defer cancel()
	if err != nil {
		xl.Errorf("get TsStream error, errr =%#v", err)
		return errors.New("get TsStream error")
	}
	if params.fmt == "fmp4" {
		c.Header("Content-Type", "video/mp4")
	} else {
		c.Header("Content-Type", "video/flv")
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.Stream(func(w io.Writer) bool {
		if ret, err := r.Recv(); err == nil {
			w.Write(ret.Stream)
			return true
		}
		return false
	})
	return nil
}

func getDownUrlWithPm3u8(domain, fileName string, user *userInfo) string {
	expireT := time.Now().Add(time.Hour).Unix()
	pm3u8 := "pm3u8/0/expires/86400"
	urlToSign := fmt.Sprintf("http://%s/%s?%s&e=%d", domain, fileName, pm3u8, expireT)
	mac := &qbox.Mac{AccessKey: user.ak, SecretKey: []byte(user.sk)}
	token := mac.Sign([]byte(urlToSign))
	return fmt.Sprintf("%s&token=%s", urlToSign, token)
}

func getDomain(xl *xlog.Logger, bucket string, user *userInfo) (string, error) {
	zone, err := models.GetZone(user.ak, bucket)
	host := zone.GetApiHost(false)
	url := fmt.Sprintf("%s%s", host+"/v6/domain/list?tbl=", bucket)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		xl.Errorf("%#v", err)
		return "", err
	}
	rpcClient := models.NewRpcClient(user.uid)
	resp, err := rpcClient.Do(context.Background(), request)
	if err != nil {
		return "", err

	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var domain []string
	for {
		if err := dec.Decode(&domain); err == io.EOF {
			break

		} else if err != nil {
			return "", err

		}

	}
	if len(domain) == 0 {
		return "", nil
	}

	return domain[0], nil
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
