package controllers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/m3u8"
	"qiniu.com/models"
)

// sample requset url = /playback/12345.m3u8?from=1532499345&to=1532499345&e=1532499345&token=xxxxxx
func GetPlayBackm3u8(c *gin.Context) {
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
	xl.Infof("info[0].Namespace %v", info[0].Namespace)
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
	playlist, err, _, code := getPlaybackList(xl, params, bucket, userInfo)
	if err != nil {
		xl.Errorf("get playback list error, error = %#v", err.Error())
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	// make m3u8 file name with "uaid + from + end.m3u8" if user not given
	fileName := params.key
	if fileName == "" {
		from := strconv.FormatInt(params.from, 10)
		end := strconv.FormatInt(params.to, 10)
		fileName = params.uaid + from + end + ".m3u8"
	}

	// upload new m3u8 file to kodo bucket
	m3u8File := m3u8.Mkm3u8(playlist, xl)
	err = uploadNewFile(fileName, bucket, []byte(m3u8File), userInfo)
	if err != nil {
		xl.Errorf("uplaod New m3u8 file failed, error = %#v", err.Error())
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}
	var storelist []map[string]interface{}
	m := map[string]interface{}{
		"m3u8": m3u8File,
	}
	storelist = append(storelist, m)
	var list []map[string]interface{}
	err = KodoBatch(xl, list, storelist, KODO_COMMAND_DELETE_AFTER_DAYS, bucket, 1, userInfo)
	if err != nil {
		xl.Error(err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(200, gin.H{
		"key":    fileName,
		"bucket": bucket,
	})
}

func getPlaybackList(xl *xlog.Logger, params *requestParams, bucket string, user *userInfo) ([]map[string]interface{}, error, int64, int) {
	segs, _, err := segMod.GetSegmentTsInfo(xl, params.from, params.to, bucket, params.uaid, 0, "", user.uid, user.ak)
	if err != nil {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		return nil, errors.New("Service Internal Error"), 0, 500
	}
	if len(segs) == 0 {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		return nil, errors.New("can't find stream in this period"), 0, 404
	}
	var playlist []map[string]interface{}

	var total int64
	var fsize int64
	for _, v := range segs {
		start, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			return nil, errors.New("start time format error"), 0, 500
		}
		end, ok := v[models.SEGMENT_ITEM_END_TIME].(int64)
		if !ok {
			return nil, errors.New("end time format error"), 0, 500
		}
		duration := float64(end-start) / 1000
		total += int64(duration)
		filename, ok := v[models.SEGMENT_ITEM_FILE_NAME].(string)

		if !ok {
			return nil, errors.New("filename format error"), 0, 500

		}
		m := map[string]interface{}{
			"duration": duration,
			"url":      "/" + filename,
		}
		playlist = append(playlist, m)
		size, ok := v[models.SEGMENT_ITEM_FSIZE].(int64)
		if !ok {
			return nil, errors.New("size format error"), 0, 500
		}
		fsize += size
	}
	return playlist, nil, fsize, 200
}
