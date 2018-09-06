package models

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"github.com/qiniu/xlog.v1"
)

var (
	// test code, TODO change to token api
	accessKey = "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ"
	secretKey = "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS"
	bucket    = "ipcamera"
	cfg       storage.Config
)

type Segment interface {
	GetSegmentTsInfo(xl *xlog.Logger, index, rows int, starttime, endtime int64, bucketurl, uaid, mark string) ([]map[string]interface{}, error)
	GetFragmentTsInfo(xl *xlog.Logger, index, rows int, starttime, endtime int64, bucketurl, uaid, mark string) ([]map[string]interface{}, error)
}

const (
	SEGMENT_FILENAME_SUB_LEN  = 6
	FRAGMENT_FILENAME_SUB_LEN = 4
	FRAME_FILENAME_SUB_LEN    = 4
	MAX_SEGMENT_TS_TIME_STAMP = 20
)

type SegmentKodoModel struct {
}

//TODO AKSK should be get in packet
func (m *SegmentKodoModel) Init() error {

	cfg = storage.Config{
		// 是否使用https域名进行资源管理
		UseHTTPS: false,
	}
	return nil
}

// segment filename should be ts/uaid/startts/endts/fragment_start_ts/expiry.ts
// fragment filename should be seg/uaid/startts/seg_end_ts
func GetInfoFromFilename(s, sep string) (error, map[string]interface{}) {
	sub := strings.Split(s, sep)
	var info map[string]interface{}

	// some file just upload by IPC but not add endtime by controller, we just skip this file
	if (sub[0] == "ts" && len(sub) != SEGMENT_FILENAME_SUB_LEN) || (sub[0] == "seg" && len(sub) != FRAGMENT_FILENAME_SUB_LEN) || (sub[0] == "frame" && len(sub) != FRAME_FILENAME_SUB_LEN) {
		return nil, info
	}

	//uaid := sub[1]
	if len(sub) == FRAGMENT_FILENAME_SUB_LEN && sub[0] == "seg" {
		starttime, err := strconv.ParseInt(sub[2], 10, 64)
		if err != nil {
			return err, info
		}
		endtime, err1 := strconv.ParseInt(sub[3], 10, 64)
		if err1 != nil {
			return err1, info
		}
		info = map[string]interface{}{
			SEGMENT_ITEM_START_TIME: starttime,
			SEGMENT_ITEM_END_TIME:   endtime,
		}
	} else if len(sub) == SEGMENT_FILENAME_SUB_LEN && sub[0] == "ts" {
		starttime, err := strconv.ParseInt(sub[2], 10, 64)
		if err != nil {
			return err, info
		}
		endtime, err1 := strconv.ParseInt(sub[3], 10, 64)
		if err1 != nil {
			return err1, info
		}
		fragmentStartTime, err2 := strconv.ParseInt(sub[4], 10, 64)
		if err2 != nil {
			return err2, info
		}
		expriy := strings.Split(sub[5], ".")
		if len(expriy) != 2 {
			return fmt.Errorf("the filename is error [%s]", sub[5]), info
		}
		exprie, err3 := strconv.ParseInt(expriy[0], 10, 64)
		if err3 != nil {
			return err3, info
		}
		info = map[string]interface{}{
			SEGMENT_ITEM_FRAGMENT_START_TIME: fragmentStartTime,
			SEGMENT_ITEM_START_TIME:          starttime,
			SEGMENT_ITEM_END_TIME:            endtime,
			SEGMENT_ITEM_FILE_NAME:           s,
			SEGMENT_ITEM_EXPIRE:              exprie,
		}
	} else if len(sub) == FRAME_FILENAME_SUB_LEN && sub[0] == "frame" {
		starttime, err := strconv.ParseInt(sub[2], 10, 64)
		if err != nil {
			return err, info
		}
		r := strings.Split(sub[3], ".")
		if len(r) != 2 {
			return fmt.Errorf("the filename is error [%s]", sub[3]), info
		}
		fragmentStartTime, err2 := strconv.ParseInt(r[0], 10, 64)
		if err2 != nil {
			return err2, info
		}
		info = map[string]interface{}{
			SEGMENT_ITEM_FRAGMENT_START_TIME: fragmentStartTime,
			SEGMENT_ITEM_START_TIME:          starttime,
			SEGMENT_ITEM_FILE_NAME:           s,
		}
	}
	return nil, info
}

// Calculate mark.
// Return []yyyy/mm/dd, if same day and same hour, return [1]yyyy/mm/dd/hh
func calculateMark(xl *xlog.Logger, starttime int64, uaid, head string) string {
	k2 := fmt.Sprintf("%s/%s/%d", head, uaid, starttime/1000 - MAX_SEGMENT_TS_TIME_STAMP)
	xl.Infof("CalculateMark k %s", k2)

	m := map[string]interface{}{
		"k": k2,
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	encodeString := base64.StdEncoding.EncodeToString(b)
	return encodeString
}

// Get Segment TS info List.
func (m *SegmentKodoModel) GetSegmentTsInfo(xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, mark string) ([]map[string]interface{}, string, error) {
	//todo change to get aksk
	mac := qbox.NewMac(accessKey, secretKey)
	// 指定空间所在的区域，如果不指定将自动探测
	// 如果没有特殊需求，默认不需要指定
	//cfg.Zone=&storage.ZoneHuabei
	bucketManager := storage.NewBucketManager(mac, &cfg)
	pre := time.Now().UnixNano()
	var r []map[string]interface{}
	delimiter := ""
	marker := ""
	nextMarker := ""
	total := 0
	if mark != "" {
		marker = mark
	} else {
		marker = calculateMark(xl, starttime, uaid, "ts")
	}

	prefix := "ts/" + uaid + "/"
	ctx, cancelFunc := context.WithCancel(context.Background())
	xl.Infof("GetSegmentTsInfo prefix ********* %s \n", prefix)
	entries, err := bucketManager.ListBucketContext(ctx, bucket, prefix, delimiter, marker)
	if err != nil {
                info := err.(*storage.ErrorInfo)
                if info.Code == 200 {
                        return r, "", nil
                } else {
                        return r, "", err
                }
	}

	for listItem1 := range entries {
		err, info := GetInfoFromFilename(listItem1.Item.Key, "/")
		if err != nil {
			fmt.Println(err)
			cancelFunc()
			break
		}
		if len(info) == 0 {
			continue
		}
		if info[SEGMENT_ITEM_START_TIME].(int64) / 1000 > endtime / 1000 {
			cancelFunc()
			break
		}
		if info[SEGMENT_ITEM_END_TIME].(int64) > starttime {
			xl.Infof("GetTsInfo info[SEGMENT_ITEM_START_TIME] %d \n", info[SEGMENT_ITEM_START_TIME].(int64))
			r = append(r, info)
			total++
		}
		if total >= limit && limit != 0 {
			nextMarker = listItem1.Marker
			break
		}
	}
	xl.Infof("find segment need %d ms ******", (time.Now().UnixNano()-pre)/1000000)
	return r, nextMarker, nil
}

// Get Fragment Ts info List.
func (m *SegmentKodoModel) GetFragmentTsInfo(xl *xlog.Logger, count int, starttime, endtime int64, bucketurl, uaid, mark string) ([]map[string]interface{}, string, error) {
	pre := time.Now().UnixNano()
	//todo change to get aksk
	mac := qbox.NewMac(accessKey, secretKey)
	// 指定空间所在的区域，如果不指定将自动探测
	// 如果没有特殊需求，默认不需要指定
	//cfg.Zone=&storage.ZoneHuabei
	bucketManager := storage.NewBucketManager(mac, &cfg)
	var r []map[string]interface{}
	delimiter := ""
	nextMarker := ""
	marker := ""
	total := 0
	if mark != "" {
		marker = mark
	} else {
		marker = calculateMark(xl, starttime, uaid, "seg")
	}
	prefix := "seg/" + uaid + "/"
	xl.Infof("GetFragmentTsInfo prefix  %s \n", prefix)
	ctx, cancelFunc := context.WithCancel(context.Background())
	entries, err := bucketManager.ListBucketContext(ctx, bucket, prefix, delimiter, marker)
	if err != nil {
		xl.Errorf("GetFragmentTsInfo ListBucketContext %#v", err)
		info := err.(*storage.ErrorInfo)
		if info.Code == 200 {
			return r, "", nil
		} else {
			return r, "", err
		}
	}
	for listItem1 := range entries {

		err, info := GetInfoFromFilename(listItem1.Item.Key, "/")
		if err != nil {
			// if one file is not correct, continue to next
			continue
		}

		if info[SEGMENT_ITEM_START_TIME].(int64) / 1000 > endtime / 1000 {
			cancelFunc()
			break
		}
                if info[SEGMENT_ITEM_END_TIME].(int64) > starttime {
                        xl.Infof("GetFragmentTsInfo info[SEGMENT_ITEM_START_TIME] %d \n", info[SEGMENT_ITEM_START_TIME].(int64))
                        xl.Infof("GetFragmentTsInfo info[SEGMENT_ITEM_END_TIME] %d \n", info[SEGMENT_ITEM_END_TIME].(int64))
                        r = append(r, info)
                        total++
                } 
		if total >= count && count != 0 {
			nextMarker = listItem1.Marker
			break
		}
	}
	xl.Infof("find fragment need %d ms\n", (time.Now().UnixNano()-pre)/1000000)
	return r, nextMarker, err
}

// Get Frame info List.
func (m *SegmentKodoModel) GetFrameInfo(xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string) ([]map[string]interface{}, error) {
	pre := time.Now().UnixNano()
	//todo change to get aksk
	mac := qbox.NewMac(accessKey, secretKey)
	// 指定空间所在的区域，如果不指定将自动探测
	// 如果没有特殊需求，默认不需要指定
	//cfg.Zone=&storage.ZoneHuabei
	bucketManager := storage.NewBucketManager(mac, &cfg)
	var r []map[string]interface{}
	delimiter := ""
	total := 0
	marker := calculateMark(xl, starttime, uaid, "frame")
	prefix := "frame/" + uaid + "/"
	xl.Infof("GetFragmentTsInfo prefix  %s \n", prefix)
	ctx, cancelFunc := context.WithCancel(context.Background())
	entries, err := bucketManager.ListBucketContext(ctx, bucket, prefix, delimiter, marker)
	if err != nil {
		xl.Errorf("GetFrameTsInfo ListBucketContext %#v", err)
		info := err.(*storage.ErrorInfo)
		if info.Code == 200 {
			return r, nil
		} else {
			return r, err
		}
	}
	for listItem1 := range entries {

		err, info := GetInfoFromFilename(listItem1.Item.Key, "/")
		if err != nil {
			// if one file is not correct, continue to next
			xl.Infof("GetFrameInfo err  %s \n", err)
			continue
		}
		if info[SEGMENT_ITEM_START_TIME].(int64) / 1000 > endtime / 1000 {
			cancelFunc()
			break
		}
		if info[SEGMENT_ITEM_START_TIME].(int64) >= starttime {
			r = append(r, info)
			total++
		}
	}
	xl.Infof("find frame need %d ms\n", (time.Now().UnixNano()-pre)/1000000)
	return r, err
}
