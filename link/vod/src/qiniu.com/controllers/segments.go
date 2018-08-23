package controllers

import (
	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
)

func GetSegments(c *gin.Context) {

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
		xl.Errorf("verify token failed")
		c.JSON(401, gin.H{
			"error": "bad token",
		})
		return
	}
	xl.Infof("uid= %v, deviceid = %v, from = %v, to = %v, limit = %v, marker = %v", params.uid, params.uaid, params.from, params.to, params.limit, params.marker)

	segs, marker, err := SegMod.GetFragmentTsInfo(xl, params.limit, params.from*1000, params.to*1000, "ipcamera", params.uaid, params.marker)
	if err != nil {
		xl.Errorf("get segments list error, error =%#v", err)
		c.JSON(500, nil)
		return
	}
	if segs == nil {
		c.JSON(200, gin.H{
			"segments": []string{},
			"marker":   marker,
		})
		return
	}

	c.JSON(200, gin.H{
		"segments": segs,
		"marker":   marker,
	})

}
