package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
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
			"error": "bad from/to time, from great or equal to",
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

	if ok, err := VerifyAuth(xl, c.Request); err != nil || ok != true {
		xl.Errorf("verify auth failed %#v", err)
		c.JSON(401, gin.H{
			"error": "verify auth failed",
		})
		return
	}

	xl.Infof("uid= %v, uaid = %v, from = %v, to = %v, namespace = %v", params.uid, params.uaid, params.from, params.to, params.namespace)

	frames, err := SegMod.GetFrameInfo(xl, params.from, params.to, params.namespace, params.uaid)
	if err != nil {
		xl.Errorf("get FrameInfo falied, error = %#v", err)
		c.JSON(500, nil)
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
			c.JSON(500, nil)
			return
		}
		realUrl := GetUrlWithDownLoadToken(xl, "http://pdwjeyj6v.bkt.clouddn.com/", filename, 0)
		starttime, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			xl.Errorf("segment start format error %#v", v)
			c.JSON(500, nil)
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
