package controllers

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
)

type format struct {
	Filename string `json:"filename"`
	Duration string `json:"duration"`
}
type fopNotifyInfo struct {
	Format format `json:"formant"`
}

func FopNotify(c *gin.Context) {
	action := c.Param("action")

	xl := xlog.New(c.Writer, c.Request)
	c.Header("Content-Type", "application/json")
	if ok, err := VerifyAuth(xl, c.Request); err == nil && ok == true {
		xl.Infof("verify auth falied %#v", err)
		c.JSON(401, gin.H{
			"error": "verify auth failed",
		})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "read callback body failed",
		})
		return
	}
	xl.Infof("%s", body)
	var notifyData fopNotifyInfo
	err = json.Unmarshal(body, &notifyData)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "decode json failed",
		})
		return
	}
	if action == "avinfo" {
		// update duration
		xl.Info(notifyData.Format.Duration, notifyData.Format.Filename)
	}
}
