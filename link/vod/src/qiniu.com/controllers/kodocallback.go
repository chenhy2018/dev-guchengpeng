package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
	rpc "qiniupkg.com/x/rpc.v7"
	//	redis "gopkg.in/redis.v5"
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
	// key  =  ts/ua_id/start_ts/fragment_start_ts/expiry.ts
	if len(ids) < 5 {
		xl.Errorf("bad file name, file name = %s", key)
		c.JSON(500, gin.H{
			"error": "bad file name",
		})
		return

	}

	// Add namespace&ua check
	err, exprie := GetNameSpaceInfo(xl, kodoData.Bucket, ids[1])
	if err != nil {
		xl.Errorf("error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = HandleTs(xl, kodoData.Bucket, ids, kodoData, c.Request, exprie)
	if err != nil {
		xl.Errorf("error = %#v", err.Error())
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"name":    key,
	})
}

func HandleTs(xl *xlog.Logger, bucket string, ids []string, kodoData kodoCallBack, req *http.Request, dbExprie int) error {
	startTime, err := strconv.ParseInt(ids[2], 10, 64)
	if err != nil {
		return fmt.Errorf("parse ts file name failed, filename = %#v", ids[2])
	}
	duration, err := strconv.ParseFloat(kodoData.Duration, 64)
	if err != nil {
		return fmt.Errorf("parse duration failed, duration = %#v", kodoData.Duration)
	}
	endTime := startTime + int64(duration*1000)
	var tsExpire int
	if dbExprie == 0 {
		expire := strings.Split(ids[4], ".")
		if len(expire) != 2 {
			return fmt.Errorf("bad file name, expire = %#v", ids[4])
		}
		e, err := strconv.ParseInt(expire[0], 10, 32)
		tsExpire = int(e)
		if err != nil {
			return fmt.Errorf("parse ts expire failed tsExpire = %#v", ids[4])
		}
	} else {
		tsExpire = dbExprie
	}
	// key -->ts/ua_id/start_ts/endts/segment_start_ts/expiry.ts
	newFilName := append(ids[:3], append([]string{strconv.FormatInt(endTime, 10)}, ids[3:]...)...)
	xl.Infof("oldFileName = %v, newFileName = %v", kodoData.Key, strings.Join(newFilName[:], "/"))

	user, err := getUserInfo(xl, req)
	if err != nil {
		return fmt.Errorf("get user info falied, erro = %#v", err)
	}
	mac := qbox.NewMac(user.ak, user.sk)
	segPrefix := strings.Join([]string{"seg", ids[1], ids[3]}, "/")
	if err := updateTsName(xl, kodoData.Key, strings.Join(newFilName[:], "/"), bucket, segPrefix, endTime, int(tsExpire), mac); err != nil {
		return fmt.Errorf("ts filename update failed err = %#v", err)
	}
	jpegName := strings.Join([]string{"frame", ids[1], ids[2], ids[3]}, "/")
	if err = fop(strings.Join(newFilName[:], "/"), kodoData.Bucket, jpegName+".jpeg", mac); err != nil {
		return fmt.Errorf("fop operation failed err = %#v", err)
	}
	return nil
}

func updateTsName(xl *xlog.Logger, srcTsKey, destTsKey, bucket, segPrefix string, endTime int64, expire int, mac *qbox.Mac) (err error) {
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

func fop(filename, bucket, jpegName string, mac *qbox.Mac) error {
	cfg := storage.Config{
		UseHTTPS: false,
	}
	operationManager := storage.NewOperationManager(mac, &cfg)

	fopVframe := fmt.Sprintf("vframe/jpg/offset/0|saveas/%s",
		storage.EncodedEntry(bucket, jpegName))
	_, err := operationManager.Pfop(bucket, filename, fopVframe, "", "", true)
	return err
}
