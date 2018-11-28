package controllers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
	"io"
	"qiniu.com/models"
	"time"
)

type saveasArgs struct {
	Fname    string `json:"fname"`
	Format   string `json:"format"`
	Pipeline string `json:"pipeline"`
	Notify   string `json:"notify"`

	// -1: 表示不修改ts文件的expire属性
	// 0:  表示修改ts文件生命周期为永久保存
	// >0: 表示修改ts文件的的生命周期为expireDays
	ExpireDays int `json:"expire"`
}

// sample requset url = /saveas?from=1532499345&to=1532499345
func Saveas(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	var saveasData saveasArgs
	dec := json.NewDecoder(c.Request.Body)
	for {
		if err := dec.Decode(&saveasData); err == io.EOF {
			break
		} else if err != nil {
			xl.Errorf("json decode failed %#v", err)
			c.JSON(400, gin.H{
				"error": "json decode failed",
			})
			return
		}
	}
	if params.to == 0 || params.from == 0 {
		xl.Errorf("from or to is not correct")
		c.JSON(400, gin.H{
			"error": "from or to is not correct",
		})
		return
	}
	eightMin := int64((8 * time.Minute).Seconds() * 1000)
	if (params.to - params.from) > eightMin {
		xl.Errorf("only support save 8 min stream")
		c.JSON(400, gin.H{
			"error": "only support save 8 min stream",
		})
		return
	}
	userInfo, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	xl.Infof("uid = %v, uaid = %v, from = %v, to = %v", userInfo.uid, params.uaid, params.from, params.to)

	info, err := UaMod.GetUaInfo(xl, userInfo.uid, params.namespace, params.uaid)
	if err != nil || len(info) == 0 {
		xl.Errorf("get ua info failed, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "ua is not correct",
		})
		return
	}
	xl.Infof("info[0].Namespace %v", info[0].Namespace)
	namespace := info[0].Namespace
	bucket, domain, err := GetBucketAndDomain(xl, userInfo.uid, namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	// get ts list from kodo
	saveasList, key, err, code := getSaveasList(xl, params, bucket, domain, userInfo)
	if err != nil {
		xl.Errorf("get playback list error, error = %#v", err.Error())
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(saveasList)
	persistentID, err := fopSaveas(xl, key, saveasList, bucket, saveasData.Fname, saveasData.Pipeline, saveasData.Notify, saveasData.ExpireDays, userInfo)
	if err != nil {
		xl.Errorf("fop failed, error = %#v", err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(200, gin.H{
		"fname":        saveasData.Fname,
		"persistentID": persistentID,
	})
}

func fopSaveas(xl *xlog.Logger, key, saveasList, bucket, fname, pipeline, notify string, expires int, user *userInfo) (string, error) {
	cfg := storage.Config{
		UseHTTPS: false,
	}
	//rpcClient := models.NewRpcClient(user.uid)
	fmt.Printf("ak %s sk %s  bucket %s uid %s\n", user.ak, user.sk, bucket, user.uid)
	mac := qbox.Mac{
		AccessKey: user.ak,
		SecretKey: []byte(user.sk),
	}
	//mac := qbox.NewMac("JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ", "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS")
	zone, err := models.GetZone(user.ak, bucket)
	if err != nil {
		return "", err
	}
	//fmt.Printf("zone %#v rpc %#v \n", zone, rpcClient)
	cfg.Zone = zone
	operationManager := storage.NewOperationManager(&mac, &cfg)
	saveasFile := storage.EncodedEntry(bucket, fname)
	fopCmd := fmt.Sprintf("avconcat/2/format/mp4/index/1%s|saveas/%s", saveasList, saveasFile)

	if expires > 0 {
		fopCmd += fmt.Sprintf("/deleteAfterDays/%d", expires)
	}
	fmt.Println(fopCmd)
	persistentID, err := operationManager.Pfop(bucket, key, fopCmd, "", "", true)
	fmt.Println(persistentID)
	return persistentID, err
}

func getSaveasList(xl *xlog.Logger, params *requestParams, bucket, domain string, user *userInfo) (string, string, error, int) {
	segs, _, err := segMod.GetSegmentTsInfo(xl, params.from, params.to, bucket, params.uaid, 0, "", user.uid, user.ak)
	if err != nil {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		return "", "", errors.New("Service Internal Error"), 500
	}
	if len(segs) == 0 {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		return "", "", errors.New("can't find stream in this period"), 404
	}
	var saveaslist string
	var key string
	var total int64
	for _, v := range segs {
		start, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			return "", "", errors.New("start time format error"), 500
		}
		end, ok := v[models.SEGMENT_ITEM_END_TIME].(int64)
		if !ok {
			return "", "", errors.New("end time format error"), 500
		}
		duration := float64(end-start) / 1000
		total += int64(duration)
		filename, ok := v[models.SEGMENT_ITEM_FILE_NAME].(string)
		if !ok {
			return "", "", errors.New("filename format error"), 500
		}
		if key == "" {
			key = filename
		} else {
			fmt.Println(filename)
			saveaslist += "/" + base64.URLEncoding.EncodeToString([]byte("http://"+domain+"/"+filename))
		}
	}
	return saveaslist, key, nil, 200
}
