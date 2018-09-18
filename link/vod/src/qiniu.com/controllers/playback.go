package controllers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	xlog "github.com/qiniu/xlog.v1"
	"google.golang.org/grpc"
	"qiniu.com/m3u8"
	"qiniu.com/models"
	pb "qiniu.com/proto"
	"qiniu.com/system"
)

var (
	SegMod           *models.SegmentKodoModel
	fastForwardClint pb.FastForwardClient
)

func Init(conf *system.GrpcConf) {
	SegMod = &models.SegmentKodoModel{}
	SegMod.Init()
	FFGrpcClientInit(conf)

}
func FFGrpcClientInit(conf *system.GrpcConf) {
	conn, err := grpc.Dial(conf.Addr, grpc.WithInsecure())
	if err != nil {
		fmt.Println("Init gprc failed")
	}
	fastForwardClint = pb.NewFastForwardClient(conn)
}

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
	if params.to <= params.from {
		xl.Errorf("bad from/to time, from = %v, to = %v", params.from, params.to)
		c.JSON(400, gin.H{
			"error": "bad from/to time, from great or equal than to",
		})
		return
	}

	if !VerifyToken(xl, params.expire, params.token, c.Request) {
		xl.Errorf("verify token falied")
		c.JSON(401, gin.H{
			"error": "bad token",
		})
		return
	}
	xl.Infof("uaid = %v, from = %v, to = %v, namespace = %v", params.uaid, params.from, params.to, params.namespace)
	dayInMilliSec := int64((24 * time.Hour).Seconds() * 1000)
	if (params.to - params.from) > dayInMilliSec {
		xl.Errorf("bad from/to time, from = %v, to = %v", params.from, params.to)
		c.JSON(400, gin.H{
			"error": "bad from/to time, currently we only support playback in 24 hours",
		})
		return
	}
	userInfo, err := getUserInfo(xl, c.Request)
	if params.speed != 1 {
		if err := getFastForwardStream(xl, params, c, userInfo); err != nil {
			xl.Errorf("get fastforward stream error , error = %v", err.Error())
			c.JSON(500, gin.H{"error": "Service Internal Error"})
		}
		return
	}
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}

	bucket, err := GetBucket(xl, getUid(userInfo.uid), params.namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}

	mac := qbox.NewMac(userInfo.ak, userInfo.sk)

	segs, _, err := SegMod.GetSegmentTsInfo(xl, params.from, params.to, bucket, params.uaid, 0, "", mac)
	if err != nil {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}

	if len(segs) == 0 {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		c.JSON(404, gin.H{"error": "can't find stream in this period"})
		return
	}
	playlist, err := getPlaybackList(xl, segs, userInfo)
	if err != nil {
		xl.Errorf("get playback list error, error = %#v", err.Error())
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}
	c.Header("Content-Type", "application/x-mpegURL")
	c.Header("Access-Control-Allow-Origin", "*")
	c.String(200, m3u8.Mkm3u8(playlist, xl))
}
func getPlaybackList(xl *xlog.Logger, segs []map[string]interface{}, user *userInfo) ([]map[string]interface{}, error) {
	var playlist []map[string]interface{}

	var total int64
	for _, v := range segs {
		start, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			return nil, errors.New("start time format error")
		}
		end, ok := v[models.SEGMENT_ITEM_END_TIME].(int64)
		if !ok {
			return nil, errors.New("end time format error")
		}
		duration := float64(end-start) / 1000
		total += int64(duration)
		filename, ok := v[models.SEGMENT_ITEM_FILE_NAME].(string)

		if !ok {
			return nil, errors.New("filename format error")

		}
		realUrl := GetUrlWithDownLoadToken(xl, "http://pdwjeyj6v.bkt.clouddn.com/", filename, total, user)

		m := map[string]interface{}{
			"duration": duration,
			"url":      realUrl,
		}
		playlist = append(playlist, m)

	}
	return playlist, nil
}
func getFastForwardStream(xl *xlog.Logger, params *requestParams, c *gin.Context, user *userInfo) error {
	url := c.Request.URL.String()
	fullUrl := "http://" + c.Request.Host + url

	req := new(pb.FastForwardInfo)
	expire := time.Now().Add(time.Hour).Unix()
	req.Url = getNewToken(fullUrl, expire, user)
	req.Speed = params.speed
	ctx, cancel := context.WithCancel(context.Background())
	r, err := fastForwardClint.GetTsStream(ctx, req)
	defer cancel()
	if err != nil {
		xl.Errorf("get TsStream error, errr =%#v", err)
		return errors.New("get TsStream error")
	}
	c.Header("Content-Type", "video/mp4")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Content-Disposition", "attachment;filename="+params.uaid+".ts")
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
	prefix := strings.Split(origin, "&speed")[0]
	playbackBaseUrl := prefix + "&e=" + strconv.FormatInt(expire, 10)
	// using uid password as ak/sk
	mac := qbox.NewMac(user.ak, user.sk)
	token := mac.Sign([]byte(playbackBaseUrl))
	return playbackBaseUrl + "&token=" + token
}
