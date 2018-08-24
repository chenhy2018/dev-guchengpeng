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
	if ok, err := VerifyAuth(xl, c.Request); err != nil || ok != true {
		xl.Errorf("verify auth failed %#v", err)
		c.JSON(401, gin.H{
			"error": "verify auth failed",
		})
		return
	}

	xl.Infof("uid= %v, deviceid = %v, from = %v, to = %v, limit = %v, marker = %v, namespace = %v", params.uid, params.uaid, params.from, params.to, params.limit, params.marker, params.namespace)

	segs, marker, err := SegMod.GetFragmentTsInfo(xl, params.limit, params.from, params.to, params.namespace, params.uaid, params.marker)
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
