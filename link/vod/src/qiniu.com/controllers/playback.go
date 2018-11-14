package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/auth"
	"qiniu.com/m3u8"
	"qiniu.com/models"
	pb "qiniu.com/proto"
)

// sample requset url = /playback/12345.m3u8?from=1532499345&to=1532499345&e=1532499345&token=xxxxxx
func GetPlayBackm3u8(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = checkParams(xl, params)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
	}

	userInfo, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
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
	fileName := params.m3u8FileName
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
	c.Header("Access-Control-Allow-Origin", "*")

	// fast forward case
	if params.speed != 1 {
		if err := getFastForwardStream(xl, params, c, userInfo, bucket, fileName); err != nil {
			xl.Errorf("get fastforward stream error , error = %v", err.Error())
			c.JSON(500, gin.H{"error": "Service Internal Error"})
		}
		return
	}
	c.JSON(200, gin.H{
		"key":    fileName,
		"bucket": bucket,
	})
}

func getPlaybackList(xl *xlog.Logger, params *requestParams, bucket string, user *userInfo) ([]map[string]interface{}, error, int) {
	segs, _, err := segMod.GetSegmentTsInfo(xl, params.from, params.to, bucket, params.uaid, 0, "", user.uid, user.ak)
	if err != nil {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		return nil, errors.New("Service Internal Error"), 500
	}
	if len(segs) == 0 {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		return nil, errors.New("can't find stream in this period"), 404
	}
	var playlist []map[string]interface{}

	var total int64
	for _, v := range segs {
		start, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			return nil, errors.New("start time format error"), 500
		}
		end, ok := v[models.SEGMENT_ITEM_END_TIME].(int64)
		if !ok {
			return nil, errors.New("end time format error"), 500
		}
		duration := float64(end-start) / 1000
		total += int64(duration)
		filename, ok := v[models.SEGMENT_ITEM_FILE_NAME].(string)

		if !ok {
			return nil, errors.New("filename format error"), 500

		}
		m := map[string]interface{}{
			"duration": duration,
			"url":      "/" + filename,
		}
		playlist = append(playlist, m)

	}
	return playlist, nil, 200
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
		return err
	}
	mac, err := getSKByAkFromQconf(xl, user.ak)
	if err != nil {
		return err
	}
	req.Url = getDownUrlWithPm3u8(domain, fileName, mac)
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

func getDownUrlWithPm3u8(domain, fileName string, mac *qbox.Mac) string {
	expireT := time.Now().Add(time.Hour).Unix()
	pm3u8 := "pm3u8/0/expires/86400"
	urlToSign := fmt.Sprintf("http://%s/%s?%s&e=%d", domain, fileName, pm3u8, expireT)
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

// cause admin ak/sk can't sign download token for
// normal user, only get user ak/sk for fastforward
// after new fastforward developed this code will deleted
// from here, please DON'T using this code this other purpose.
func getSKByAkFromQconf(xl *xlog.Logger, ak string) (*qbox.Mac, error) {
	return &qbox.Mac{AccessKey: "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ", SecretKey: []byte("G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS")}, nil
	accessInfo, err := auth.GetUserInfoFromQconf(xl, ak)
	if err != nil {
		return nil, err
	}
	return &qbox.Mac{AccessKey: ak, SecretKey: accessInfo.Secret}, nil

}
