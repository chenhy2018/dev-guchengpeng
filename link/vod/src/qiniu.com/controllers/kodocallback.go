package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
	rpc "qiniupkg.com/x/rpc.v7"
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

	if err := updateTsName(xl, key, strings.Join(newFilName[:], "/"), kodoData.Bucket, segPrefix, endTime, int(tsExpire)); err != nil {
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

func updateTsName(xl *xlog.Logger, srcTsKey, destTsKey, bucket, segPrefix string, endTime int64, expire int) (err error) {
	mac := qbox.NewMac(accessKey, secretKey)

	cfg := storage.Config{
		UseHTTPS: false,
	}
	bucketManager := storage.NewBucketManager(mac, &cfg)

	delimiter := ""
	marker := ""
	force := true
	limit := 1000
	// list seg file by segment id
	entries, _, _, _, err := bucketManager.ListFiles(bucket, segPrefix, delimiter, marker, limit)
	if err != nil {
		return
	}

	ops := []string{}
	var segAction string
	newSegName := segPrefix + "/" + strconv.FormatInt(endTime, 10)
	if len(entries) == 0 {
		// create new seg file if doesn't exist
		if err := createNewsegFile(newSegName, bucket, mac); err != nil {
			xl.Errorf("create seg file error, err = %#v", err)
		}
		xl.Info("create new seg file file name =%#v", newSegName)
	} else {
		// in some bad case if first two ts files arrived at same time, we
		// may create two seg file. we should make sure only one seg file
		// exist for current segment

		// update one seg file endtime
		ops = append(ops, storage.URIMove(bucket, entries[0].Key, bucket, newSegName, force))
		// udpate seg file expire time
		ops = append(ops, storage.URIDeleteAfterDays(bucket, segPrefix+"/"+strconv.FormatInt(endTime, 10), expire))
		xl.Info(entries[0].Key, "---->", newSegName)

		// delete other files
		for i := 1; i < len(entries); i++ {
			segAction = storage.URIDelete(bucket, entries[i].Key)
			ops = append(ops, segAction)
			xl.Info("delete file = ", entries[i])
		}
	}

	// add endtime for ts file
	ops = append(ops, storage.URIMove(bucket, srcTsKey, bucket, destTsKey, force))

	rets, err := bucketManager.Batch(ops)
	// check batch error
	if err != nil {
		if _, ok := err.(*rpc.ErrorInfo); ok {
			for _, ret := range rets {
				if ret.Code != 200 {
					xl.Error(ret.Data.Error)
				}
			}

		} else {
			xl.Errorf("batch error, %#v", err)
		}
	}
	return
}

func createNewsegFile(filename, bucket string, mac *qbox.Mac) error {

	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}

	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{}
	// 空间对应的机房
	//cfg.Zone = &storage.ZoneHuadong
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}

	data := []byte{}
	return formUploader.Put(context.Background(), &ret, upToken, filename, bytes.NewReader(data), 0, &putExtra)
}
