package controllers

import (
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	xlog "qiniu.com/cdnbase/xlog.v1"
)

// sample requst see: https://developer.qiniu.com/kodo/manual/1653/callback
func UploadTs(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")

	c.Header("Content-Type", "text/plain")
	xl := xlog.New(c.Writer, c.Request)

	if !verifyAuth(xl, authHeader) {
		c.JSON(401, gin.H{
			"error": "verify auth falied",
		})
		return
	}
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "read callback body falied",
		})
		return
	}
	values, err := url.ParseQuery(string(body))
	if err != nil || len(values["name"]) == 0 {
		c.JSON(500, gin.H{
			"error": "read callback body falied",
		})
		return
	}

	fileName := strings.Split(values["name"][0], "_")
	if len(fileName) < 3 {
		c.JSON(500, gin.H{
			"error": "bad file name",
		})
		return
	}
	xl.Infof("upload file = %s", values["name"][0])
	// filename = uid_devciceid_segmentid.ts
	// models.addDevice(xl, fileName[0], fileName[1], fileName[2], time.Now().Unix())
	c.JSON(200, gin.H{
		"success": true,
		"name":    values["name"][0],
	})
}

func verifyAuth(xl *xlog.Logger, authHeader string) bool {

	return true
}
