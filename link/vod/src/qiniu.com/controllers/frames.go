package controllers

import (
	"errors"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/models"
)

type frameInfo struct {
	Key       string `json:"key"`
	Timestamp int64  `json:"timestamp"`
}

func GetFrames(c *gin.Context) {

	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request parameter falied, error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	err = checkParams(xl, params)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	xl.Infof("uaid = %v, from = %v, to = %v", params.uaid, params.from, params.to)

	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user info error, error = %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	namespace := params.namespace
	info, err := UaMod.GetUaInfo(xl, user.uid, namespace, params.uaid)
	if err != nil && len(info) == 0 {
		xl.Errorf("get ua info failed, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "ua is not correct",
		})
		return
	}

	bucket, _, err := GetBucketAndDomain(xl, user.uid, namespace)
	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	frames, err := segMod.GetFrameInfo(xl, params.from, params.to, bucket, params.uaid, user.uid, user.ak)
	if err != nil {
		xl.Errorf("get FrameInfo falied, error = %#v", err)
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}

	if frames == nil {
		c.JSON(200, gin.H{"frames": []string{}})
		return
	}
	formatedFrames, err := formatFramesData(xl, frames)
	if err != nil {
		xl.Errorf("Parse kodo file name failed, error = %#v", err)
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}
	c.JSON(200, gin.H{"frames": formatedFrames})
}

func formatFramesData(xl *xlog.Logger, frames []map[string]interface{}) ([]frameInfo, error) {

	newFrames := make([]frameInfo, 0, len(frames))
	for _, v := range frames {
		fileName, ok := v[models.SEGMENT_ITEM_FILE_NAME].(string)
		if !ok {
			xl.Errorf("filename format error %#v", v)
			return []frameInfo{}, errors.New("filename format error")

		}
		starttime, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			xl.Errorf("segment start format error %#v", v)
			return []frameInfo{}, errors.New("segment start format error")

		}
		frame := frameInfo{
			Key:       fileName,
			Timestamp: starttime / 1000,
		}

		newFrames = append(newFrames, frame)

	}

	return newFrames, nil
}
