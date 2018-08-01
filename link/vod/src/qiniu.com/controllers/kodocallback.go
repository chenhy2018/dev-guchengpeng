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

type avFormat struct {
	duration string `json:"duration"`
}
type avinfo struct {
	format avFormat `json:"format"`
}
type kodoCallBack struct {
	key    string `json:"key"`
	hash   string `json:"hash"`
	size   int64  `json:"fsize"`
	bucket string `json:"bucket"`
	name   string `json:"name"`
	avInfo avinfo `json:"avinfo"`
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
	xl.Infof("%s", kodoData)
	fileName := kodoData.key

	xl.Infof("upload file = %s", fileName)

	UidDevicIdSegId := strings.Split(fileName, "_")
	if len(UidDevicIdSegId) < 3 {
		c.JSON(500, gin.H{
			"error": "bad file name",
		})
		return

	}
	segId, err := strconv.ParseInt(UidDevicIdSegId[1], 10, 32)

	ts := models.SegmentTsInfo{
		Uuid:              UidDevicIdSegId[0],
		DeviceId:          UidDevicIdSegId[1],
		StartTime:         time.Now().Unix(),
		FileName:          fileName,
		EndTime:           time.Now().Add(time.Minute).Unix(),
		FragmentStartTime: int(segId),
	}
	segMod := &models.SegmentModel{}
	segMod.AddSegmentTS(ts)

	c.JSON(200, gin.H{
		"success": true,
		"name":    fileName,
	})
}
