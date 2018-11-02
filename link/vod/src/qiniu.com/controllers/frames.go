package controllers

import (
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
	err = checkParams(xl, params)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
	}

	xl.Infof("uaid = %v, from = %v, to = %v", params.uaid, params.from, params.to)

	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user info error, error = %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	info, err := UaMod.GetUaInfo(xl, getUid(user.uid), params.uaid)
	if err != nil && len(info) == 0 {
		xl.Errorf("get ua info failed, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "ua is not correct",
		})
		return
	}
	namespace := info[0].Namespace

	bucket, err := GetBucket(xl, getUid(user.uid), namespace)
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
		realUrl := GetUrlWithDownLoadToken(xl, domain, filename, 0, mac)
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
