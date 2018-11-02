package controllers

import (
	"encoding/json"
	"io"

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
	dec := json.NewDecoder(c.Request.Body)
	var aksk akskInfo
	for {
		if err := dec.Decode(&aksk); err == io.EOF {
			break
		} else if err != nil {
			xl.Errorf("json decode failed")
			c.JSON(400, gin.H{
				"error": "json decode failed",
			})
			return
		}
	}

	SetUserInfo(aksk.Ak, aksk.Sk)
	c.JSON(200, gin.H{"success": "true"})
}
