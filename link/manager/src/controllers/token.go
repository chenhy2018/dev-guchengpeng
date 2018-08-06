package controllers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

var (
	accessKey = "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ"
	secretKey = "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS"
	bucket    = "ipcamera"
	host      = "http://39.107.247.14:8088"
)

func GetUnloadToken(c *gin.Context) {

	c.Header("Content-Type", "application/json")
	// 简单上传凭证
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	mac := qbox.NewMac(accessKey, secretKey)

	// 设置上传凭证有效期
	putPolicy = storage.PutPolicy{
		Scope:            bucket,
		CallbackURL:      "http://39.107.247.14:8088/qiniu/upload/callback",
		CallbackBody:     `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)","name":"$(x:name)", "duration":"$(avinfo.format.duration)"}`,
		CallbackBodyType: "application/json",
		DeleteAfterDays:  7,
	}
	putPolicy.Expires = 7200 //示例2小时有效期

	upToken := putPolicy.UploadToken(mac)

	c.JSON(200, gin.H{
		"token":           upToken,
		"deleteAfterdays": 7,
	})
}

func AddTokenAndRedirect(c *gin.Context) {
	uid := c.Param("uid")

	expireT := time.Now().Add(time.Hour).Unix()
	playbackBaseUrl := c.Request.URL.String() + "&e=" + strconv.FormatInt(expireT, 10)
	// using uid password as ak/sk
	mac := qbox.NewMac(uid, uid)
	token := mac.Sign([]byte(playbackBaseUrl))

	realUrl := host + playbackBaseUrl + "&token=" + token
	c.Redirect(301, realUrl)
}
