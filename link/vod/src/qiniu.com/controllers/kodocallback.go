package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
	//	redis "gopkg.in/redis.v5"
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
	// key  =  ts/uid/ua_id/yyyy/mm/dd/hh/mm/ss/mmm/fragment_start_ts/expiry.ts
	if len(ids) < 12 {
		xl.Errorf("bad file name, file name = %s", key)
		c.JSON(500, gin.H{
			"error": "bad file name",
		})
		return

	}
	err, startTime := models.TransferTimeToInt64(ids[3:10])
	if err != nil {
		xl.Errorf("parse ts file name failed, filename = %#v", ids[3:10])
		c.JSON(500, gin.H{
			"error": "parse ts file name failed",
		})
		return
	}
	duration, err := strconv.ParseFloat(kodoData.Duration, 64)
	if err != nil {
		xl.Errorf("parse duration failed, duration = %#v", kodoData.Duration)
		c.JSON(500, gin.H{
			"error": "parse duration failed",
		})
		return
	}
	endTime := startTime + int64(duration*1000)
	segStart, err := strconv.ParseInt(ids[10], 10, 64)
	tsExpire, _ := strconv.ParseInt(ids[11], 10, 32)

	if err != nil {
		xl.Errorf("parse segment start failed, body = %#v", ids[10])
		c.JSON(500, gin.H{
			"error": "parse segment start failed",
		})
		return
	}
	// key -->ts/uid/ua_id/yyyy/mm/dd/hh/mm/ss/mmm/endts/segment_start_ts/expiry.ts
	newFilName := append(ids[:10], append([]string{strconv.FormatInt(endTime, 10)}, ids[10:]...)...)
	xl.Infof("oldFileName = %v, newFileName = %v", kodoData.Key, strings.Join(newFilName[:], "/"))

	segPrefix := strings.Join([]string{"seg", ids[1], ids[2], models.TransferTimeToString(segStart)}, "/")

	if err := updateTsName(key, strings.Join(newFilName[:], "/"), kodoData.Bucket, segPrefix, endTime, int(tsExpire)); err != nil {
		xl.Errorf("ts filename update failed err = %#v", err)
		c.JSON(500, gin.H{
			"error": "update ts file name failed",
		})
		return
	}
	/*
		auth := strings.TrimRight(c.Request.Header.Get("Authorization"), "QBox ")
		ak := strings.Split(auth, ":")[0]
		segStartTime, err := strconv.ParseInt(ids[10], 10, 64)
		segKey := []string{"seg", ids[1], ids[2], models.TransferTimeToString(segStartTime), strconv.FormatInt(endTime, 10)}


			if err := redisdb.Set(strings.Join(segKey[:], "/")+"_ts", []string{strconv.FormatInt(endTime, 10), ak, kodoData.Bucket}, 0).Err(); err != nil {
				xl.Errorf("insert to redis failed, error = %#v", err)
			}
	*/
	c.JSON(200, gin.H{
		"success": true,
		"name":    key,
	})
}

func updateTsName(srcTsKey, destTsKey, bucket, segPrefix string, endTime int64, expire int) (err error) {
	mac := qbox.NewMac(accessKey, secretKey)

	cfg := storage.Config{
		UseHTTPS: false,
	}
	bucketManager := storage.NewBucketManager(mac, &cfg)

	limit := 1
	delimiter := ""
	marker := ""
	force := true

	beforeList := time.Now().UnixNano()
	entries, _, _, _, err := bucketManager.ListFiles(bucket, segPrefix, delimiter, marker, limit)
	if err != nil {
		return
	}
	afterList := time.Now().UnixNano()

	ops := make([]string, 0, 3)
	var segAction string
	if len(entries) == 0 {
		// create new seg file if doesn't exist
		segAction = storage.URICopy(bucket, "tmpsegfile", bucket, segPrefix+"/"+strconv.FormatInt(endTime, 10), force)
	} else {
		segAction = storage.URIMove(bucket, entries[0].Key, bucket, segPrefix+"/"+strconv.FormatInt(endTime, 10), force)
	}

	// udpate seg file expire time

	storage.URIDeleteAfterDays(bucket, segPrefix+"/"+strconv.FormatInt(endTime, 10), expire)
	ops = append(ops, segAction)
	ops = append(ops, storage.URIMove(bucket, srcTsKey, bucket, destTsKey, force))
	ops = append(ops, storage.URIDeleteAfterDays(bucket, segPrefix+"/"+strconv.FormatInt(endTime, 10), expire))
	_, err = bucketManager.Batch(ops)
	afterBatch := time.Now().UnixNano()
	fmt.Printf("list spend = %dms, batch spend = %dms\n", (afterList-beforeList)/1000000, (afterBatch-afterList)/1000000)
	return
}
