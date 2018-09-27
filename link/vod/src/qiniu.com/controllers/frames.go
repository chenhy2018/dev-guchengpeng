package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/models"
)

type FrameInfo struct {
	DownloadUr string `json:"download_url"`
	Timestamp  int64  `json:"timestamp"`
}

func GetFrames(c *gin.Context) {

	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request parameter falied, error = %#v", err.Error())
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

	dayInMilliSec := int64((24 * time.Hour).Seconds() * 1000)
	if (params.to - params.from) > dayInMilliSec {
		xl.Errorf("bad from/to time, from = %v, to = %v", params.from, params.to)
		c.JSON(400, gin.H{
			"error": "bad from/to time, currently we only support playback in 24 hours",
		})
		return
	}

	xl.Infof("uaid = %v, from = %v, to = %v, namespace = %v", params.uaid, params.from, params.to, params.namespace)

	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user info error, error = %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	bucket, err := GetBucket(xl, getUid(user.uid), params.namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}
	domain, err := getDomain(xl, bucket, user)
	if err != nil {
		xl.Errorf("getDomain error, error =  %#v", err)
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}
	if domain == "" {
		xl.Errorf("bucket is not correct, err = %#v", err)
		c.JSON(403, gin.H{
			"error": "bucket is not correct",
		})
		return
	}

	domain = "http://" + domain
	mac := qbox.NewMac(user.ak, user.sk)
	frames, err := segMod.GetFrameInfo(xl, params.from, params.to, bucket, params.uaid, mac)
	if err != nil {
		xl.Errorf("get FrameInfo falied, error = %#v", err)
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}
	if frames == nil {
		c.JSON(200, gin.H{
			"frames": []string{},
		})
		return
	}

	framesWithToken := make([]FrameInfo, 0, len(frames))
	for _, v := range frames {
		filename, ok := v[models.SEGMENT_ITEM_FILE_NAME].(string)
		if !ok {
			xl.Errorf("filename format error %#v", v)
			c.JSON(500, gin.H{"error": "Service Internal Error"})
			return
		}
		realUrl := GetUrlWithDownLoadToken(xl, domain, filename, 0, user)
		starttime, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			xl.Errorf("segment start format error %#v", v)
			c.JSON(500, gin.H{"error": "Service Internal Error"})
			return
		}
		frame := FrameInfo{DownloadUr: realUrl,
			Timestamp: starttime / 1000}
		framesWithToken = append(framesWithToken, frame)
	}

	c.JSON(200, gin.H{
		"frames": framesWithToken,
	})

}
