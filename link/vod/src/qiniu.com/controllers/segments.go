package controllers

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/models"
)

type segInfo struct {
	StartTime int64 `json:"starttime"`
	EndTime   int64 `json:"endtime"`
}

func GetSegments(c *gin.Context) {

	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	if params.to <= params.from {
		xl.Errorf("bad from/to time, from = %v, to = %v", params.from, params.to)
		c.JSON(400, gin.H{
			"error": "bad from/to time, from great or equal to",
		})
		return
	}

	dayInMilliSec := int64((24 * time.Hour).Seconds() * 1000)

	if (params.to - params.from) > dayInMilliSec*7 {
		xl.Errorf("bad from/to time, from = %v, to = %v", params.from, params.to)
		c.JSON(400, gin.H{
			"error": "bad from/to time, currently we only support segments in 7x24 hours",
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

	segs, marker := []segInfo{}, params.marker
	for {
		var ret []map[string]interface{}
		ret, marker, err = SegMod.GetFragmentTsInfo(xl, params.limit-len(segs), params.from-dayInMilliSec, params.to, params.namespace, params.uaid, marker)
		if err != nil {
			xl.Errorf("get segments list error, error =%#v", err)
			c.JSON(500, nil)
			return
		}
		if ret == nil {
			c.JSON(200, gin.H{
				"segments": []string{},
				"marker":   marker,
			})
			return
		}

		segs, err = filterSegs(ret, params)
		if err != nil {
			xl.Error("parse seg start/end failed")
			c.JSON(500, gin.H{
				"error": "parse seg start/end failed",
			})
			return
		}
		if marker == "" || len(segs) == params.limit {
			break
		}
	}
	c.JSON(200, gin.H{
		"segments": segs,
		"marker":   marker,
	})

}

func filterSegs(ret []map[string]interface{}, params *requestParams) (segs []segInfo, err error) {
	for _, v := range ret {
		starttime, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			return []segInfo{}, errors.New("parse starttime error")
		}
		endtime, ok := v[models.SEGMENT_ITEM_END_TIME].(int64)
		if !ok {
			return []segInfo{}, errors.New("parse starttime error")
		}

		// if from to in the middle seg case
		//seg           A1----------A2  B1------B2  C1--------C2
		//url               D1---------------------------D2
		//result            D1------A2  B1------B2  C1---D2

		if params.from > endtime {
			continue
		}
		if (params.from >= starttime) && (params.from <= endtime) {
			starttime = params.from
		}

		if (params.to >= starttime) && (params.to <= endtime) {
			endtime = params.to
		}

		seg := segInfo{StartTime: starttime / 1000,
			EndTime: endtime / 1000}
		segs = append(segs, seg)
	}
	return
}
