package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/m3u8"
	"qiniu.com/models"
	"strconv"
	"strings"
)

func mkkey(user *userInfo, params *requestParams) string {
	return fmt.Sprintf("%s%s%s%d", user.uid, params.namespace, params.uaid, params.expire)
}

// sample requset url = /live/12345.m3u8?from=1532499345&to=1532499345&e=1532499345&token=xxxxxx
func GetLivem3u8(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	if params.from == 0 {
		xl.Errorf("parse request falied from = %#v", params.from)
		c.JSON(400, gin.H{
			"error": "params.from == 0",
		})
		return
	}
	userInfo, err, code := getUserInfoByAk(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed %v", err)
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	if !verifyToken(xl, params.expire, params.token, c.Request, userInfo) {
		xl.Errorf("verify token falied")
		c.JSON(401, gin.H{
			"error": "bad token",
		})
		return
	}
	xl.Infof("uaid = %v, from = %v, namespace = %v", params.uaid, params.from, params.namespace)
	key := params.reqid
	value := redisGet(key)
	sub := strings.Split(value, "/")
	sequeue := int64(1)
	mark := ""
	if len(sub) > 1 {
		sequeue, err = strconv.ParseInt(sub[0], 10, 64)
		mark = sub[1]
	}
	info, err := UaMod.GetUaInfo(xl, userInfo.uid, params.namespace, params.uaid)
	if err != nil || len(info) == 0 {
		xl.Errorf("get ua info failed, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "ua is not correct",
		})
		return
	}
	namespace := params.namespace
	bucket, domain, err := GetBucketAndDomain(xl, userInfo.uid, namespace)

	if err != nil {
		xl.Errorf("get bucket error, error =  %#v", err)
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	mac := &qbox.Mac{AccessKey: userInfo.ak, SecretKey: []byte(userInfo.sk)}

	playlist, err := getLiveList(xl, sequeue, mac, params, bucket, "http://"+domain, mark, key, userInfo)
	if err != nil {
		xl.Errorf("get playback list error, error = %#v", err.Error())
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}
	c.Header("Content-Type", "application/x-mpegURL")
	c.String(200, m3u8.MkLivem3u8(playlist, uint64(sequeue+1), xl))
}

func getLiveList(xl *xlog.Logger, sequeue int64, mac *qbox.Mac, params *requestParams, bucket, domain, mark, key string, user *userInfo) ([]map[string]interface{}, error) {
	segs, _, err := segMod.GetSegmentTsInfo(xl, params.from, int64(^uint64(0)>>1), bucket, params.uaid, 5, mark, user.uid, user.ak)
	if err != nil {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		return nil, errors.New("Service Internal Error")
	}
	if len(segs) == 0 {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		return nil, errors.New("can't find stream in this period")
	}

	if len(segs) > 1 {
		next, ok := segs[0][models.SEGMENT_ITEM_MARK].(string)
		if mark != next && ok {
			redisSet(xl, key, fmt.Sprintf("%d/%s", sequeue+1, next))
		}
	}
	var playlist []map[string]interface{}

	var total int64
	for _, v := range segs {
		start, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			return nil, errors.New("start time format error")
		}
		end, ok := v[models.SEGMENT_ITEM_END_TIME].(int64)
		if !ok {
			return nil, errors.New("end time format error")
		}
		duration := float64(end-start) / 1000
		total += int64(duration)
		filename, ok := v[models.SEGMENT_ITEM_FILE_NAME].(string)

		if !ok {
			return nil, errors.New("filename format error")

		}
		realUrl := GetUrlWithDownLoadToken(xl, domain, filename, total, mac)
		m := map[string]interface{}{
			"duration": duration,
			"url":      realUrl,
		}
		playlist = append(playlist, m)

	}
	return playlist, nil
}

func getNewToken(origin string, expire int64, user *userInfo) string {
	playbackBaseUrl := origin + "&e=" + strconv.FormatInt(expire, 10)
	// using uid password as ak/sk
	mac := qbox.NewMac(user.ak, user.sk)
	token := mac.Sign([]byte(playbackBaseUrl))
	return playbackBaseUrl + "&token=" + token
}
