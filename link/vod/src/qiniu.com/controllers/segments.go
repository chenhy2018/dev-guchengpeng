package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/models"
)

func GetSegments(c *gin.Context) {

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
	segs, err := segMod.GetFragmentTsInfo(0, 0, params.from*1000, params.to*1000, params.uid, params.deviceid)
	if err != nil {
		xl.Errorf("get segments list error, error =%v", err)
		c.JSON(500, nil)
		return
	}

	c.JSON(200, gin.H{
		"segments": segs,
	})

}
