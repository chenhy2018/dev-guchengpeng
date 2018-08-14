package controllers

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
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
		xl.Errorf("verify auth failed %#v", err)
		c.JSON(401, gin.H{
			"error": "verify auth failed",
		})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		xl.Errorf("parse request body failed, body = %#v", body)
		c.JSON(500, gin.H{
			"error": "read callback body failed",
		})
		return
	}
	xl.Infof("%s", body)
	var kodoData kodoCallBack
	err = json.Unmarshal(body, &kodoData)
	xl.Infof("%#v", kodoData)
	key := kodoData.Key
	ids := strings.Split(key, "/")
	if len(ids) < 12 {
		xl.Errorf("bad file name, file name = %s", key)
		c.JSON(500, gin.H{
			"error": "bad file name",
		})
		return

	}
	year, _ := strconv.ParseInt(ids[3], 10, 32)
	month, _ := strconv.ParseInt(ids[4], 10, 32)
	day, _ := strconv.ParseInt(ids[5], 10, 32)
	hour, _ := strconv.ParseInt(ids[6], 10, 32)
	minute, _ := strconv.ParseInt(ids[7], 10, 32)
	second, _ := strconv.ParseInt(ids[8], 10, 32)
	milsecond, _ := strconv.ParseInt(ids[9], 10, 32)

	startTime := time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), int(milsecond)*1000000, nil)
	duration, err := strconv.ParseFloat(kodoData.Duration, 64)
	endTime := startTime.UnixNano()/1000000 + int64(duration*1000)

	newFilName := append(ids[:10], append([]string{string(endTime)}, ids[10:]...)...)
	xl.Infof("oldFileName = %v\n, newFileName = %v", kodoData.Key, strings.Join(newFilName[:], "/"))

	if err := updateTsName(kodoData.Bucket, key, kodoData.Bucket, strings.Join(newFilName[:], "/")); err != nil {
		xl.Errorf("ts filename update failed err = %#v", err)
		c.JSON(500, gin.H{
			"error": "update ts file name failed",
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"name":    key,
	})
}

func updateTsName(srcBucket, srcKey, destBucket, destKey string) error {
	mac := qbox.NewMac(accessKey, secretKey)

	cfg := storage.Config{
		UseHTTPS: false,
	}

	bucketManager := storage.NewBucketManager(mac, &cfg)

	force := false
	return bucketManager.Move(srcBucket, srcKey, destBucket, destKey, force)
}
