package controllers

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/xlog.v1"
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
	if ok, err := VerifyAuth(xl, c.Request); err == nil && ok == true {
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
	UidDevicIdSegId := strings.Split(fileName, "_")
	if len(UidDevicIdSegId) < 3 {
		c.JSON(500, gin.H{
			"error": "bad file name",
		})
		return

	}
	segId, err := strconv.ParseInt(UidDevicIdSegId[2], 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"status": "bad file name"})
		return
	}

	endTime := time.Now()
	d, _ := time.ParseDuration(kodoData.Duration + "s")
	startTime := endTime.Add(-d)
	xl.Infof("start = %v\n, end = %v", startTime, endTime, d.Nanoseconds())
	ts := models.SegmentTsInfo{
		Uuid:              UidDevicIdSegId[0],
		DeviceId:          UidDevicIdSegId[1],
		StartTime:         startTime.UnixNano(),
		FileName:          fileName,
		EndTime:           endTime.UnixNano(),
		FragmentStartTime: int(segId),
	}
	segMod := &models.SegmentModel{}
	segMod.AddSegmentTS(ts)

	c.JSON(200, gin.H{
		"success": true,
		"name":    fileName,
	})
}
