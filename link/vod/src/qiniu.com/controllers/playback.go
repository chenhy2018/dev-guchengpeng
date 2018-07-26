package controllers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/xlog.v1"
)

// sample requset url = /playback/13764829407/12345?from=1532499345&to=1532499345&e=1532499345&token=xxxxxx
func PlayBackGetm3u8(c *gin.Context) {
	Uid := c.Param("uid")
	DeviceId := c.Param("deviceId")
	from := c.DefaultQuery("from", string(time.Now().Unix()))
	to := c.DefaultQuery("to", string(time.Now().Unix()))
	expire := c.DefaultQuery("e", "")
	token := c.DefaultQuery("token", "")

	fromT, err := strconv.ParseInt(from, 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"status": "Parse from time failed"})
		return
	}
	toT, err := strconv.ParseInt(to, 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"status": "Parse to time failed"})
		return
	}
	expireT, err := strconv.ParseInt(expire, 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"status": "Parse expire time failed"})
		return
	}
	xl := xlog.New(c.Writer, c.Request)
	xl.Infof("uid= %v, deviceid = %v, from = %v, to = %v, expire = %v, token=%v", Uid, DeviceId, time.Unix(fromT, 0), time.Unix(toT, 0), time.Unix(expireT, 0), token)

	if !verifyToken(xl, expire, token, c.Request.URL.String()) {
		c.JSON(401, gin.H{"status": "bad token"})
		return
	}

	// get ts list from models
	// models.GetTsByStartEnd(xl, Uid, DeviceId, time.Unix(fromT, 0), time.Unix(toT, 0))

	// makem3u8
	m3u8 := "#EXTM3U\n" +
		"#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=200000\n" +
		"gear1/prog_index.m3u8\n" +
		"#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=311111\n" +
		"gear2/prog_index.m3u8\n" +
		"#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=484444\n" +
		"gear3/prog_index.m3u8\n" +
		"#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=737777\n" +
		"gear4/prog_index.m3u8"
	c.Header("Content-Type", "text/plain")
	c.String(200, m3u8)
}

func verifyToken(xl *xlog.Logger, expire, token, url string) bool {
	if expire == "" || token == "" {
		return false
	}
	tokenIndex := strings.Index(url, "&token=")
	xl.Info(url[0:tokenIndex])
	return true
}
