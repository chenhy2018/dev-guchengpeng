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
	UidDevicIdSegId := strings.Split(fileName, "/")
	if len(UidDevicIdSegId) < 5 {
		c.JSON(500, gin.H{
			"error": "bad file name",
		})
		return

	}
	segId, err := strconv.ParseInt(UidDevicIdSegId[3], 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"status": "bad file name"})
		return
	}
	expireAfter, _ := strconv.ParseInt(UidDevicIdSegId[0], 10, 32)

	start, err := strconv.ParseInt(UidDevicIdSegId[3], 10, 32)
	startTime := time.Unix(start, 0)
	d, _ := time.ParseDuration(kodoData.Duration + "s")
	endTime := startTime.Add(d)
	xl.Infof("start = %v\n, end = %v", startTime, endTime, d.Nanoseconds())
        expireAfterSecond := time.Duration(expireAfter * 24 * 60 * 60)
	ts := models.SegmentTsInfo{
		Uid:               UidDevicIdSegId[1],
		UaId:              UidDevicIdSegId[2],
		StartTime:         startTime.UnixNano(),
		FileName:          fileName,
		EndTime:           endTime.UnixNano(),
		Expire:            time.Now().Add(expireAfterSecond*time.Second),
		FragmentStartTime: int64(segId),
	}
	segMod := &models.SegmentModel{}
	segMod.AddSegmentTS(xl, ts)

	c.JSON(200, gin.H{
		"success": true,
		"name":    fileName,
	})
}
