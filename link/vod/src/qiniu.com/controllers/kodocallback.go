package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/xlog.v1"
	"qiniu.com/models"
	"qiniupkg.com/api.v7/auth/qbox"
)

type kodoCallBack struct {
	Key    string `json:"key"`
	Hash   string `json:"hash"`
	Fsize  int64  `json:"fsize"`
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

const (
	accessKey = "kevidUP5vchk8Qs9f9cjKo1dH3nscIkQSaVBjYx7"
	secretKey = "KG9zawEhR4axJT0Kgn_VX_046LZxkUZBhcgURAC0"
)

// sample requst see: https://developer.qiniu.com/kodo/manual/1653/callback
func UploadTs(c *gin.Context) {

	xl := xlog.New(c.Writer, c.Request)

	c.Header("Content-Type", "application/json")
	if ok, err := verifyAuth(xl, c.Request); err == nil && ok == true {
		xl.Infof("verify auth falied %#v", err)
		c.JSON(401, gin.H{
			"error": "verify auth falied",
		})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "read callback body falied",
		})
		return
	}
	xl.Infof("%s", body)
	var kodoData kodoCallBack
	err = json.Unmarshal(body, &kodoData)
	xl.Infof("%s", kodoData)
	fileName := kodoData.Name

	xl.Infof("upload file = %s", fileName)
	// filename = uid_devciceid_segmentid.ts
	// models.addDevice(xl, fileName[0], fileName[1], fileName[2], time.Now().Unix())
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

func verifyAuth(xl *xlog.Logger, req *http.Request) (bool, error) {

	mac := qbox.NewMac(accessKey, secretKey)
	return mac.VerifyCallback(req)
}
