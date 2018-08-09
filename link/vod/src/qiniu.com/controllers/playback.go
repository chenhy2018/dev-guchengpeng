package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/m3u8"
	"qiniu.com/models"
)

// sample requset url = /playback/13764829407/12345?from=1532499345&to=1532499345&e=1532499345&token=xxxxxx
func GetPlayBackm3u8(c *gin.Context) {

	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !VerifyToken(xl, params.expire, params.token, c.Request.URL.String(), params.uid) {
		c.JSON(401, gin.H{
			"error": "bad token",
		})
		return
	}
	xl.Infof("uid= %v, deviceid = %v, from = %v, to = %v", params.uid, params.deviceid, time.Unix(params.from, 0), time.Unix(params.to, 0))

	segMod := &models.SegmentModel{}
	segs, err := segMod.GetSegmentTsInfo(xl, 0, 0, time.Unix(params.from, 0).UnixNano(), time.Unix(params.to, 0).UnixNano(), params.uid, params.deviceid)
	pPlaylist := new(m3u8.MediaPlaylist)
	pPlaylist.Init(32, 32)
	var playlist []map[string]interface{}

	if err == nil {
		for count := 0; count < len(segs); count++ {
			duration := float64(segs[count][models.SEGMENT_ITEM_END_TIME].(int64)-segs[count][models.SEGMENT_ITEM_START_TIME].(int64)) / 1000000000
			realUrl := GetUrlWithDownLoadToken(xl, "http://pcgtsa42m.bkt.clouddn.com/"+segs[count][models.SEGMENT_ITEM_FILE_NAME].(string))
			pPlaylist.AppendSegment(realUrl, duration, params.uid)

			m := map[string]interface{}{
				"duration": duration,
				"url":      realUrl,
			}
			playlist = append(playlist, m)

		}
	}

	c.Header("Content-Type", "application/x-mpegURL")
	c.String(200, pPlaylist.String())
}
