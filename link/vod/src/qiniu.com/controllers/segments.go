package controllers

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/models"
)

type segInfo struct {
	StartTime int64  `json:"starttime"`
	EndTime   int64  `json:"endtime"`
	Snapshot  string `json:"snapshot"`
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
			"error": "bad from/to time, from great or equal than to",
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

	xl.Infof("deviceid = %v, from = %v, to = %v, limit = %v, marker = %v, namespace = %v", params.uaid, params.from, params.to, params.limit, params.marker, params.namespace)

	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user info error, error = %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	info, err := UaMod.GetUaInfo(xl, user.uid, params.namespace, params.uaid)
	if err != nil || len(info) == 0 {
		xl.Errorf("get ua info failed, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "ua is not correct",
		})
		return
	}

	namespace := info[0].Namespace

	bucket, _, err := GetBucketAndDomain(xl, user.uid, namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{"error": "namespace is not correct"})
		return
	}
	newSegFrom, tsStart := getFristTsAfterFrom(xl, params.from, params.to, bucket, params.uaid, user)
	ret, marker, err := segMod.GetFragmentTsInfo(xl, params.limit, newSegFrom, params.to, bucket, params.uaid, params.marker, user.uid, user.ak)
	if err != nil {
		xl.Errorf("get segments list error, error =%#v", err)
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}
	if ret == nil {
		c.JSON(200, gin.H{
			"segments": []string{},
			"marker":   marker,
		})
		return
	}

	segs, err := filterSegs(xl, ret, params, tsStart)
	if err != nil {
		xl.Error("parse seg start/end failed")
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	}

	c.JSON(200, gin.H{
		"segments": segs,
		"marker":   marker,
	})

}
func getFristTsAfterFrom(xl *xlog.Logger, from, to int64, bucket, uaid string, user *userInfo) (int64, int64) {

	segs, _, err := segMod.GetSegmentTsInfo(xl, from, to, bucket, uaid, 1, "", user.uid, user.ak)
	if err != nil || segs == nil {
		return from, from
	}
	newSegFrom, ok := segs[0][models.SEGMENT_ITEM_FRAGMENT_START_TIME].(int64)
	if !ok {
		return from, from
	}

	newTsStart, ok := segs[0][models.SEGMENT_ITEM_START_TIME].(int64)
	if !ok {
		return from, from
	}
	return newSegFrom, newTsStart
}

func filterSegs(xl *xlog.Logger, ret []map[string]interface{}, params *requestParams, tsStart int64) ([]segInfo, error) {
	segs := []segInfo{}
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
		frameStart := starttime

		// if seg starttime is great than url.from then using first frame after
		// url.from as snapshot
		if params.from > frameStart {
			frameStart = tsStart
		}
		filename := "frame" + "/" + params.uaid + "/" + strconv.FormatInt(frameStart, 10) + "/" + strconv.FormatInt(starttime, 10) + ".jpeg"
		if (params.from >= starttime) && (params.from <= endtime) {
			starttime = params.from

		}

		if (params.to >= starttime) && (params.to <= endtime) {
			endtime = params.to

		}

		seg := segInfo{
			StartTime: starttime / 1000,
			EndTime:   endtime / 1000,
			Snapshot:  filename,
		}
		segs = append(segs, seg)

	}
	return segs, nil

}
