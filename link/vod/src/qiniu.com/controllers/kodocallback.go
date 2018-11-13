package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/models"
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

	user, _ := getUserInfo(xl, c.Request)
	c.Header("Content-Type", "application/json")

	dec := json.NewDecoder(c.Request.Body)
	var kodoData kodoCallBack
	for {

		if err := dec.Decode(&kodoData); err == io.EOF {
			break
		} else if err != nil {
			xl.Errorf("parse request body failed, body = %#v", c.Request.Body)
			c.JSON(400, gin.H{
				"error": "read callback body failed",
			})
			return
		}
	}

	xl.Infof("%#v", kodoData)

	key := kodoData.Key
	ids := strings.Split(key, "/")
	if ids[0] != "ts" {
		c.JSON(200, "")
		return
	}

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

	err = HandleTs(xl, kodoData.Bucket, ids, kodoData, c.Request, exprie, user)
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

func HandleTs(xl *xlog.Logger, bucket string, ids []string, kodoData kodoCallBack, req *http.Request, dbExprie int, user *userInfo) error {
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
	segPrefix := strings.Join([]string{"seg", ids[1], ids[3]}, "/")
	if err := updateTsName(xl, kodoData.Key, strings.Join(newFilName[:], "/"), bucket, segPrefix, endTime, int(tsExpire), user); err != nil {
		return fmt.Errorf("ts filename update failed err = %#v", err)
	}
	jpegName := strings.Join([]string{"frame", ids[1], ids[2], ids[3]}, "/")
	if err = fop(xl, strings.Join(newFilName[:], "/"), kodoData.Bucket, jpegName+".jpeg", user); err != nil {
		return fmt.Errorf("fop operation failed err = %#v", err)
	}
	return nil
}

func updateTsName(xl *xlog.Logger, srcTsKey, destTsKey, bucket, segPrefix string, endTime int64, expire int, user *userInfo) (err error) {
	bucketManager := models.NewBucketMgtWithEx(xl, user.uid, user.ak, bucket)

	ops := []string{}
	force := true

	delimiter := ""
	marker := ""
	limit := 1000
	// list seg file by segment id
	entries, _, _, _, err := bucketManager.ListFiles(bucket, segPrefix, delimiter, marker, limit)
	if err != nil {
		return
	}
	var segAction string
	newSegName := segPrefix + "/" + strconv.FormatInt(endTime, 10)
	if len(entries) == 0 {
		// create new seg file if doesn't exist
		if err := uploadNewFile(newSegName, bucket, []byte{}, user); err != nil {
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

func fop(xl *xlog.Logger, filename, bucket, jpegName string, user *userInfo) error {
	cfg := storage.Config{
		UseHTTPS: false,
	}
	rpcClient := models.NewRpcClient(user.uid, &defaultUser)
	mac := qbox.Mac{
		AccessKey: defaultUser.AccessKey,
		SecretKey: []byte(defaultUser.SecretKey),
	}
	zone, err := models.GetZone(user.ak, bucket)
	if err != nil {
		return err
	}
	cfg.Zone = zone
	operationManager := storage.NewOperationManagerEx(&mac, &cfg, &rpcClient)

	fopVframe := fmt.Sprintf("vframe/jpg/offset/0|saveas/%s",
		storage.EncodedEntry(bucket, jpegName))
	_, err = operationManager.Pfop(bucket, filename, fopVframe, "", "", true)
	return err

}
