package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/m3u8"
	"qiniu.com/models"
)

var (
	SegMod *models.SegmentKodoModel
)

func init() {
	SegMod = &models.SegmentKodoModel{}
	SegMod.Init()
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
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, nil)
		return
	}

	bucket, err := GetBucket(xl, getUid(userInfo.uid), params.namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(500, nil)
		return
	}

	mac := qbox.NewMac(userInfo.ak, userInfo.sk)

	segs, _, err := SegMod.GetSegmentTsInfo(xl, params.from, params.to, bucket, params.uaid, 0, "", mac)
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
		realUrl := GetUrlWithDownLoadToken(xl, "http://pdwjeyj6v.bkt.clouddn.com/", filename, total, userInfo)

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
