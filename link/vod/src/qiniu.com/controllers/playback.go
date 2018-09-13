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
)

var (
	SegMod           *models.SegmentKodoModel
	fastForwardClint pb.FastForwardClient
)

func init() {
	SegMod = &models.SegmentKodoModel{}
	SegMod.Init()
	getFFGrpcClient()

}
func getFFGrpcClient() pb.FastForwardClient {
	if fastForwardClint != nil {
		return fastForwardClint
	}
	conn, err := grpc.Dial("47.105.118.51:50051", grpc.WithInsecure())
	//conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())

	if err != nil {
		fmt.Println("Init gprc failedgrpcgrpc")

	}

	fastForwardClint = pb.NewFastForwardClient(conn)
	return fastForwardClint
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

	fullUrl := "http://" + c.Request.Host + c.Request.URL.String()
	if !VerifyToken(xl, params.expire, params.token, fullUrl, params.uid) {
		xl.Errorf("verify token falied")
		c.JSON(401, gin.H{
			"error": "bad token",
		})
		return
	}
	fmt.Println("speed = ", params.speed)
	xl.Infof("uid= %v, uaid = %v, from = %v, to = %v, namespace = %v", params.uid, params.uaid, params.from, params.to, params.namespace)

	dayInMilliSec := int64((24 * time.Hour).Seconds() * 1000)
	if (params.to - params.from) > dayInMilliSec {
		xl.Errorf("bad from/to time, from = %v, to = %v", params.from, params.to)
		c.JSON(400, gin.H{
			"error": "bad from/to time, currently we only support playback in 24 hours",
		})
		return
	}
	if params.speed != 1 {
		if err := getFastForwardStream(params, c); err != nil {
			xl.Errorf("get fastforward stream error , error = %v", err.Error())
			c.JSON(500, nil)
		}
		return
	}
	segs, _, err := SegMod.GetSegmentTsInfo(xl, params.from, params.to, params.namespace, params.uaid, 0, "")
	if err != nil {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		c.JSON(500, nil)
		return
	}

	if len(segs) == 0 {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		c.JSON(404, gin.H{"error": "can't find stream in this period"})
		return
	}
	var playlist []map[string]interface{}

	var total int64
	for _, v := range segs {
		start, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			xl.Errorf("start time format error %#v", v)
			c.JSON(500, nil)
			return
		}
		end, ok := v[models.SEGMENT_ITEM_END_TIME].(int64)
		if !ok {
			xl.Errorf("end time format error %#v", v)
			c.JSON(500, nil)
			return
		}
		duration := float64(end-start) / 1000
		total += int64(duration)
		filename, ok := v[models.SEGMENT_ITEM_FILE_NAME].(string)

		if !ok {
			xl.Errorf("filename format error %#v", v)
			c.JSON(500, nil)
			return
		}
		realUrl := GetUrlWithDownLoadToken(xl, "http://pdwjeyj6v.bkt.clouddn.com/", filename, total)

		m := map[string]interface{}{
			"duration": duration,
			"url":      realUrl,
		}
		playlist = append(playlist, m)

	}
	c.Header("Content-Type", "application/x-mpegURL")
	c.Header("Access-Control-Allow-Origin", "*")
	c.String(200, m3u8.Mkm3u8(playlist, xl))
}

func getFastForwardStream(params *requestParams, c *gin.Context) error {
	url := c.Request.URL.String()
	fullUrl := "http://" + c.Request.Host + url

	req := new(pb.FastForwardInfo)
	expire := time.Now().Add(time.Hour).Unix()
	req.Url = getNewToken(fullUrl, expire)
	req.Speed = params.speed
	fmt.Println(req.Url)
	ffGrpcClient := getFFGrpcClient()
	if ffGrpcClient == nil {
		return errors.New("grpc client error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	r, err := ffGrpcClient.GetTsStream(ctx, req)
	defer cancel()
	if err != nil {
		fmt.Println(err)
	}
	if r == nil {
		fmt.Println("get ts file error")
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

func getNewToken(origin string, expire int64) string {
	prefix := strings.Split(origin, "&speed")[0]
	playbackBaseUrl := prefix + "&e=" + strconv.FormatInt(expire, 10)
	// using uid password as ak/sk
	mac := qbox.NewMac(accessKey, secretKey)
	token := mac.Sign([]byte(playbackBaseUrl))
	return playbackBaseUrl + "&token=" + token
}
