package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/m3u8"
	"qiniu.com/models"
)

type storeArgs struct {
	// -1: 表示不修改ts文件的expire属性
	// 0:  表示修改ts文件生命周期为永久保存
	// >0: 表示修改ts文件的的生命周期为expireDays
	ExpireDays int `json:"expire"`
}

// sample requset url = /store?from=1532499345&to=1532499345
func MkStore(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = checkParams(xl, params)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
	}

	var storeData storeArgs
	dec := json.NewDecoder(c.Request.Body)
	for {
		if err := dec.Decode(&storeData); err == io.EOF {
			break
		} else if err != nil {
			xl.Errorf("json decode failed %#v", err)
			c.JSON(400, gin.H{
				"error": "json decode failed",
			})
			return
		}
	}

	userInfo, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
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
	namespace := info[0].Namespace
	bucket, _, err := GetBucketAndDomain(xl, userInfo.uid, namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	// get ts list from kodo
	playlist, err, fsize, code := getPlaybackList(xl, params, bucket, userInfo)
	if err != nil {
		xl.Errorf("get playback list error, error = %#v", err.Error())
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	// make m3u8 file name with "uaid + from + end.m3u8" if user not given
	from := strconv.FormatInt(params.from, 10)
	end := strconv.FormatInt(params.to, 10)
	size := strconv.FormatInt(fsize, 10)
	fileName := namespace + "/" + params.uaid + "/" + "store" + "/" + from + "/" + end + "/" + size + ".m3u8"

	// upload new m3u8 file to kodo bucket
	m3u8File := m3u8.Mkm3u8(playlist, xl)
	err = uploadNewFile(fileName, bucket, []byte(m3u8File), userInfo)
	if err != nil {
		xl.Errorf("upload New m3u8 file failed, error = %#v", err.Error())
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}

	var storelist []map[string]interface{}
	m := map[string]interface{}{
		"m3u8": fileName,
	}
	storelist = append(storelist, m)
	go KodoBatch(xl, playlist, storelist, KODO_COMMAND_DELETE_AFTER_DAYS, bucket, storeData.ExpireDays, userInfo)
	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(200, gin.H{
		"key":   fileName,
		"fsize": fsize,
	})
}

// sample requset url = /store?from=1532499345&to=1532499345
func GetStoreList(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = checkParams(xl, params)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
	}

	userInfo, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
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
	namespace := info[0].Namespace
	bucket, _, err := GetBucketAndDomain(xl, userInfo.uid, namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	// get m3u8 list from kodo
	storelist, err, code, nextMarker := getStoreList(xl, params, bucket, params.limit, params.marker, userInfo)
	if err != nil {
		xl.Errorf("get store list error, error = %#v", err.Error())
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(200, gin.H{
		"stores": storelist,
		"marker": nextMarker,
	})
}

func getStoreList(xl *xlog.Logger, params *requestParams, bucket string, limit int, marker string, user *userInfo) ([]map[string]interface{}, error, int, string) {
	segs, next, err := segMod.GetStoreInfo(xl, limit, params.from, params.to, bucket, params.namespace, params.uaid, marker, user.uid, user.ak)
	if err != nil {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		return nil, errors.New("Service Internal Error"), 500, ""
	}
	if len(segs) == 0 {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		return nil, errors.New("can't find stream in this period"), 404, ""
	}
	var storeList []map[string]interface{}

	var total int64
	for _, v := range segs {
		start, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			return nil, errors.New("start time format error"), 500, ""
		}
		end, ok := v[models.SEGMENT_ITEM_END_TIME].(int64)
		if !ok {
			return nil, errors.New("end time format error"), 500, ""
		}
		duration := float64(end-start) / 1000
		total += int64(duration)
		filename, ok := v[models.SEGMENT_ITEM_FILE_NAME].(string)

		if !ok {
			return nil, errors.New("filename format error"), 500, ""

		}
		m := map[string]interface{}{
			"start": v[models.SEGMENT_ITEM_START_TIME].(int64),
			"end":   v[models.SEGMENT_ITEM_END_TIME].(int64),
			"m3u8":  filename,
			"size":  v[models.SEGMENT_ITEM_FSIZE].(int64),
		}
		storeList = append(storeList, m)
	}
	return storeList, nil, 200, next
}

// sample requset url = /store?from=1532499345&to=1532499345&e=1532499345&token=xxxxxx
func DeleteStoreList(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = checkParams(xl, params)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
	}

	userInfo, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
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
	namespace := info[0].Namespace
	bucket, _, err := GetBucketAndDomain(xl, userInfo.uid, namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	// get m3u8 list from kodo
	storelist, err, code, nextMarker := getStoreList(xl, params, bucket, params.limit, params.marker, userInfo)
	if err != nil {
		xl.Errorf("get store list error, error = %#v", err.Error())
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	// get ts list from kodo
	playlist, err, fsize, code := getPlaybackList(xl, params, bucket, userInfo)
	if err != nil {
		xl.Errorf("get playback list error, error = %#v", err.Error())
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	go KodoBatch(xl, playlist, storelist, KODO_COMMAND_DELETE, bucket, 0, userInfo)
	xl.Infof("%v %v", storelist, nextMarker)
	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(200, gin.H{
		"fsize": fsize,
	})
}

// sample requset url = /store?from=1532499345&to=1532499345&e=1532499345&token=xxxxxx
func UpdateStoreList(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = checkParams(xl, params)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
	}

	var storeData storeArgs
	dec := json.NewDecoder(c.Request.Body)
	for {
		if err := dec.Decode(&storeData); err == io.EOF {
			break
		} else if err != nil {
			xl.Errorf("json decode failed %#v", err)
			c.JSON(400, gin.H{
				"error": "json decode failed",
			})
			return
		}
	}

	userInfo, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
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
	namespace := info[0].Namespace
	bucket, _, err := GetBucketAndDomain(xl, userInfo.uid, namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	// get m3u8 list from kodo
	storelist, err, code, _ := getStoreList(xl, params, bucket, 0, "", userInfo)
	if err != nil {
		xl.Errorf("get store list error, error = %#v", err.Error())
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	// get ts list from kodo
	playlist, err, fsize, code := getPlaybackList(xl, params, bucket, userInfo)
	if err != nil {
		xl.Errorf("get playback list error, error = %#v", err.Error())
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	go KodoBatch(xl, playlist, storelist, KODO_COMMAND_DELETE_AFTER_DAYS, bucket, storeData.ExpireDays, userInfo)

	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(200, gin.H{
		"fsize": fsize,
	})
}
