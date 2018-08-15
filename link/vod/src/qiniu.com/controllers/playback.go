package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/m3u8"
	"qiniu.com/models"
)

// sample requset url = /playback/12345.m3u8?from=1532499345&to=1532499345&e=1532499345&token=xxxxxx
func GetPlayBackm3u8(c *gin.Context) {

	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(500, gin.H{
			"error": err.Error(),
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
	xl.Infof("uid= %v, uaid = %v, from = %v, to = %v", params.uid, params.uaid, time.Unix(params.from, 0), time.Unix(params.to, 0))

	c.Header("Content-Type", "application/x-mpegURL")
	segMod := &models.SegmentKodoModel{}
	segs, err := segMod.GetSegmentTsInfo(xl, 0, 0, params.from*1000, params.to*1000, params.uid, params.uaid)
	if len(segs) == 0 {
		c.JSON(200, nil)
		return
	}

	pPlaylist := new(m3u8.MediaPlaylist)
	pPlaylist.Init(32, 32)
	var playlist []map[string]interface{}

	if err == nil {
		var total int64
		for _, v := range segs {
			duration := float64(v[models.SEGMENT_ITEM_END_TIME].(int64)-v[models.SEGMENT_ITEM_START_TIME].(int64)) / 1000
			total += int64(duration)
			realUrl := GetUrlWithDownLoadToken(xl, "http://pcgtsa42m.bkt.clouddn.com/", v[models.SEGMENT_ITEM_FILE_NAME].(string), total)
			pPlaylist.AppendSegment(realUrl, duration, params.uid)

			m := map[string]interface{}{
				"duration": duration,
				"url":      realUrl,
			}
			playlist = append(playlist, m)

		}
	}

	c.String(200, m3u8.Mkm3u8(playlist, xl))
}
