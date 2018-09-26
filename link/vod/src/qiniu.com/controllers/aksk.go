package controllers

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/system"
)

type akskInfo struct {
	Ak string `json:"accesskey"`
	Sk string `json:"secretkey"`
}

func SetPrivateAkSk(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	if system.HaveQconf() == true {
		xl.Errorf("this api is only for private")
		c.JSON(403, gin.H{
			"error": "This interface is forbidden",
		})
		return
	}
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		xl.Errorf("parse request body failed, body = %#v", body)
		c.JSON(400, gin.H{
			"error": "read callback body failed",
		})
		return
	}
	var aksk akskInfo
	err = json.Unmarshal(body, &aksk)

	if err != nil {
		xl.Errorf("parse request body failed, body = %#v", body)
		c.JSON(400, gin.H{
			"error": "read callback body failed",
		})
		return
	}
	SetUserInfo(aksk.Ak, aksk.Sk)
	c.JSON(200, gin.H{"success": "true"})
}
