package controllers

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/models"
)

type kodoCallBack struct {
	Key      string `json:"key"`
	Hash     string `json:"hash"`
	Size     int64  `json:"fsize"`
	Bucket   string `json:"bucket"`
	Name     string `json:"name"`
	Duration string `json:"duration"`
}

// sample requst see: https://developer.qiniu.com/kodo/manual/1653/callback
func UploadTs(c *gin.Context) {

	xl := xlog.New(c.Writer, c.Request)

	c.Header("Content-Type", "application/json")
	if ok, err := VerifyAuth(xl, c.Request); err != nil || ok != true {
		xl.Infof("verify auth failed %#v", err)
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
	var kodoData kodoCallBack
	err = json.Unmarshal(body, &kodoData)
	xl.Infof("%#v", kodoData)
	fileName := kodoData.Key
	ids := strings.Split(fileName, "/")
	if len(ids) < 5 {
		c.JSON(500, gin.H{
			"error": "bad file name",
		})
		return

	}
	segId, err := strconv.ParseInt(ids[3], 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"status": "bad file name"})
		return
	}
	expireAfter, _ := strconv.ParseInt(ids[0], 10, 32)

	startTime, err := strconv.ParseInt(strings.TrimRight(ids[4], ".ts"), 10, 64)
	duration, err := strconv.ParseFloat(kodoData.Duration, 64)
	endTime := startTime + int64(duration*1000)
	xl.Infof("start = %v\n, end = %v", startTime, endTime, duration)
	expireAfterSecond := time.Duration(expireAfter * 24 * 60 * 60)
	ts := models.SegmentTsInfo{
		Uid:               ids[1],
		UaId:              ids[2],
		StartTime:         startTime,
		FileName:          fileName,
		EndTime:           endTime,
		Expire:            time.Now().Add(expireAfterSecond * time.Second),
		FragmentStartTime: int64(segId),
	}
	segMod := &models.SegmentModel{}
	segMod.AddSegmentTS(ts)

	c.JSON(200, gin.H{
		"success": true,
		"name":    fileName,
	})
}
