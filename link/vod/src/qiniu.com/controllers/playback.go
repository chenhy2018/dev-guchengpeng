package controllers

import (
	"context"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	xlog "github.com/qiniu/xlog.v1"
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

	if !VerifyToken(xl, params.expire, params.token, c.Request, userInfo) {
		xl.Errorf("verify token falied")
		c.JSON(401, gin.H{
			"error": "bad token",
		})
		return
	}
	xl.Infof("uaid = %v, from = %v, to = %v", params.uaid, params.from, params.to)

	// fast forward case
	if params.speed != 1 {
		if err := getFastForwardStream(xl, params, c, userInfo); err != nil {
			xl.Errorf("get fastforward stream error , error = %v", err.Error())
			c.JSON(500, gin.H{"error": "Service Internal Error"})
		}
		return
	}
	info, err := UaMod.GetUaInfo(xl, getUid(userInfo.uid), params.uaid)
	if err != nil && len(info) == 0 {
		xl.Errorf("get ua info failed, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "ua is not correct",
		})
		return
	}
	xl.Infof("info[0].Namespace %v", info[0].Namespace)
	namespace := info[0].Namespace
	bucket, err := GetBucket(xl, getUid(userInfo.uid), namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	mac := qbox.NewMac(userInfo.ak, userInfo.sk)

	// get ts list from kodo
	playlist, err, code := getPlaybackList(xl, mac, params, bucket)
	if err != nil {
		xl.Errorf("get playback list error, error = %#v", err.Error())
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	// make m3u8 file name with "uaid + from + end.m3u8" if user not given
	fileName := params.m3u8FileName
	if fileName == "" {
		from := strconv.FormatInt(params.from, 10)
		end := strconv.FormatInt(params.from, 10)
		fileName = params.uaid + from + end + ".m3u8"
	}

	// upload new m3u8 file to kodo bucket
	m3u8File := m3u8.Mkm3u8(playlist, xl)
	err = uploadNewFile(fileName, bucket, []byte(m3u8File), mac)
	if err != nil {
		if err.Error() != "file exists" {
			xl.Errorf("uplaod New m3u8 file failed, error = %#v", err.Error())
			c.JSON(500, gin.H{"error": "Service Internal Error"})
			return
		}
	}
	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(200, gin.H{
		"key":    fileName,
		"bucket": bucket,
	})
}

func getPlaybackList(xl *xlog.Logger, mac *qbox.Mac, params *requestParams, bucket string) ([]map[string]interface{}, error, int) {
	segs, _, err := segMod.GetSegmentTsInfo(xl, params.from, params.to, bucket, params.uaid, 0, "", mac)
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
func getFastForwardStream(xl *xlog.Logger, params *requestParams, c *gin.Context, user *userInfo) error {
	// remove speed fmt from url
	url := c.Request.URL
	query := url.Query()
	query.Del("fmt")
	query.Del("speed")
	query.Del("token")
	query.Del("e")
	url.RawQuery = query.Encode()
	fullUrl := "http://" + c.Request.Host + url.String()
	req := new(pb.FastForwardInfo)
	expire := time.Now().Add(time.Hour).Unix()
	req.Url = getNewToken(fullUrl, expire, user)
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

func getNewToken(origin string, expire int64, user *userInfo) string {
	playbackBaseUrl := origin + "&e=" + strconv.FormatInt(expire, 10)
	// using uid password as ak/sk
	mac := qbox.NewMac(user.ak, user.sk)
	token := mac.Sign([]byte(playbackBaseUrl))
	return playbackBaseUrl + "&token=" + token
}
