package controllers

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/xlog.v1"
)

type kodoCallBack struct {
	Key    string `json:"key"`
	Hash   string `json:"hash"`
	Fsize  int64  `json:"fsize"`
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

// sample requst see: https://developer.qiniu.com/kodo/manual/1653/callback
func UploadTs(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")

	c.Header("Content-Type", "application/json")
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
	xl.Infof("%s", body)
	var kodoData kodoCallBack
	err = json.Unmarshal(body, &kodoData)
	xl.Infof("%s", kodoData)
	fileName := kodoData.Name

	xl.Infof("upload file = %s", fileName)
	// filename = uid_devciceid_segmentid.ts
	// models.addDevice(xl, fileName[0], fileName[1], fileName[2], time.Now().Unix())
	c.JSON(200, gin.H{
		"success": true,
		"name":    fileName,
	})
}

func verifyAuth(xl *xlog.Logger, authHeader string) bool {

	return true
}
