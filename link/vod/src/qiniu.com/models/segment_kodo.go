package models
  
import (
        "encoding/base64"
        "fmt"
        "strconv"
        "context"
        "github.com/qiniu/xlog.v1"
        "github.com/qiniu/api.v7/auth/qbox"
        "github.com/qiniu/api.v7/storage"
        "time"
        "strings"
        "encoding/json"
)

var (
        // test code, TODO change to token api
        accessKey = "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ"
        secretKey = "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS"
        bucket    = "ipcamera"
        cfg storage.Config
)

type Segment interface {
        GetSegmentTsInfo(xl *xlog.Logger, index, rows int, starttime,endtime int64, uid,uaid string) ([]map[string]interface{}, error)
        GetFragmentTsInfo(xl *xlog.Logger, index, rows int, starttime,endtime int64, uid,uaid string) ([]map[string]interface{}, error)
}

const (
        SEGMENT_FILENAME_SUB_LEN = 13
        FRAGMENT_FILENAME_SUB_LEN = 11
)

type SegmentKodoModel struct {
}

//TODO AKSK should be get in packet
func (m *SegmentKodoModel) Init() error {

        cfg = storage.Config{
                // 是否使用https域名进行资源管理
                UseHTTPS: false,
        }
        return nil;
}

// time should be int64 to []string "yyyy/mm/dd/hh/mm/ss/mmm"
func TransferTimeToString(date int64) (string) {
        //location, _ := time.LoadLocation("Asia/Shanghai")
        tm := time.Unix(date / 1000 , date % 1000 * 1000000)
        //tm = tm.In(location)
        return fmt.Sprintf("%04d/%02d/%02d/%02d/%02d/%02d/%03d", tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second(), tm.Nanosecond() /1000000)
}

// s should be []string "yyyy" "mm" "dd" "hh" "mm" "ss" "mmm" to int64
func TransferTimeToInt64(s []string) (error, int64) {
        if (len(s) != 7) {
                return fmt.Errorf("start time: format is error [%s]", s), int64(0)
        }
        year, err := strconv.ParseInt(s[0], 10, 32)
        if (err != nil) {
                 return fmt.Errorf("start time:  parser year is error [%s]", s), int64(0)
        }
        month, err1 := strconv.ParseInt(s[1], 10, 32)
        if (err1 != nil || (month > 12 || month < 1)) {
                 return fmt.Errorf("start time:  parser month is error [%s]", s), int64(0)
        }
        day, err2 := strconv.ParseInt(s[2], 10, 32)
        if (err2 != nil) {
                 return fmt.Errorf("start time:  parser day is error [%s]", s), int64(0)
        }
        hour, err3 := strconv.ParseInt(s[3], 10, 32)
        if (err3 != nil) {
                 return fmt.Errorf("start time:  parser hour is error [%s]", s), int64(0)
        }
        minute, err4 := strconv.ParseInt(s[4], 10, 32)
        if (err4 != nil) {
                 return fmt.Errorf("start time:  parser minute is error [%s]", s), int64(0)
        }
        second, err5 := strconv.ParseInt(s[5], 10, 32)
        if (err5 != nil) {
                 return fmt.Errorf("start time:  parser second is error [%s]", s), int64(0)
        }
        millisecond, err6 := strconv.ParseInt(s[6], 10, 32)
        if (err6 != nil) {
                 return fmt.Errorf("start time:  parser millisecond is error [%s]", s), int64(0)
        }
        location, _ := time.LoadLocation("Asia/Shanghai")
        t := time.Date( int(year), time.Month(month), int(day), int(hour), int(minute), int(second), int(millisecond*1000000), location)
        return nil, t.UnixNano() / 1000000
}

// segment filename should be ts/uid/ua_id/yyyy/mm/dd/hh/mm/ss/mmm/endts/fragment_start_ts/expiry.ts
// fragment filename should be seg/uid/ua_id/yyyy/mm/dd/hh/mm/ss/mmm/seg_end_ts
func GetInfoFromFilename(s, sep string) (error, map[string]interface{}) {
        sub := strings.Split(s, sep)
        var info map[string]interface{}
	
	// some file just upload by IPC but not add endtime by controller, we just skip this file
        if ((sub[0] == "ts" && len(sub) != SEGMENT_FILENAME_SUB_LEN) || (sub[0] == "seg" && len(sub) != FRAGMENT_FILENAME_SUB_LEN)) {
		return nil, info
        }
        //uid := sub[1]
        //uaid := sub[2]
        err, starttime := TransferTimeToInt64(sub[3:10])
        if (err != nil) {
                 return err, info
        }
        endtime, err1 := strconv.ParseInt(sub[10], 10, 64)
        if (err1 != nil) {
                 return err1, info
        }
        if (len(sub) == 11) {
                info = map[string]interface{} {
                        SEGMENT_ITEM_START_TIME : starttime,
                        SEGMENT_ITEM_END_TIME : endtime,
                }
        } else {
                fragmentStartTime, err2 := strconv.ParseInt(sub[11], 10, 64)
                if (err2 != nil) {
                        return err2, info
                }
                expriy := strings.Split(sub[12], ".")
                if (len(expriy) != 2) {
                        return fmt.Errorf("the filename is error [%s]", sub[12]), info
                }
                exprie, err3 := strconv.ParseInt(expriy[0], 10, 64)
                if (err3 != nil) {
                        return err3, info
                }
                info = map[string]interface{} {
                        SEGMENT_ITEM_FRAGMENT_START_TIME : fragmentStartTime,
                        SEGMENT_ITEM_START_TIME : starttime,
                        SEGMENT_ITEM_END_TIME : endtime,
                        SEGMENT_ITEM_FILE_NAME : s,
                        SEGMENT_ITEM_EXPIRE : exprie,
                }
        }
        return nil, info
}

// Calculate mark.
// Return []yyyy/mm/dd, if same day and same hour, return [1]yyyy/mm/dd/hh
func calculateMark(xl *xlog.Logger, starttime int64, uid, uaid, head string) (string) {
        starttm := time.Unix(starttime / 1000 , starttime % 1000 * 1000000)
        k2 := fmt.Sprintf("%s/%s/%s/%04d/%02d/%02d/%02d/%02d/%02d", head, uid, uaid, starttm.Year(), starttm.Month(), starttm.Day(), starttm.Hour(), starttm.Minute(), starttm.Second() -1)
        xl.Infof("CalculateMark k %s", k2)

        m := map[string]interface{}{
                 "k" : k2,
        }
        b, err := json.Marshal(m)
        if err != nil {
            return ""
        }
        encodeString := base64.StdEncoding.EncodeToString(b)
        xl.Infof("CalculateMark %s", encodeString)
        return encodeString
}

// Get Segment Ts info List.
func (m *SegmentKodoModel) GetSegmentTsInfo(xl *xlog.Logger, starttime,endtime int64, uid,uaid string) ([]map[string]interface{}, error) {
        //todo change to get aksk
        mac := qbox.NewMac(accessKey, secretKey)
        // 指定空间所在的区域，如果不指定将自动探测
        // 如果没有特殊需求，默认不需要指定
        //cfg.Zone=&storage.ZoneHuabei
        bucketManager := storage.NewBucketManager(mac, &cfg)
        pre := time.Now().UnixNano()
        var r []map[string]interface{}
        delimiter := ""
        marker := calculateMark(xl, starttime, uid, uaid, "ts")
        prefix := "ts/" + uid + "/" + uaid + "/"
        ctx, cancelFunc := context.WithCancel(context.Background())
        xl.Infof("GetSegmentTsInfo prefix ********* %s \n", prefix)
        entries, err := bucketManager.ListBucketContext(ctx, bucket, prefix, delimiter, marker)
        if err != nil {
                xl.Errorf("GetSegmentTsInfo ListBucketContext %#v", err)
                return r, err
        }

        for listItem1 := range entries {
                err, info := GetInfoFromFilename(listItem1.Item.Key, "/")
                if err != nil {
                        cancelFunc()
                        break
                }
                if (len(info) == 0) {
                        continue
                }
                if (info[SEGMENT_ITEM_END_TIME].(int64) > endtime) {
                        cancelFunc()
                        break
                }
                if (info[SEGMENT_ITEM_START_TIME].(int64) > starttime) {
                        r = append(r, info)
                }
        }
        xl.Infof("find segment need %d ms ******", (time.Now().UnixNano() - pre) / 1000000)
        return r, nil
}

// Get Fragment Ts info List.
func (m *SegmentKodoModel) GetFragmentTsInfo(xl *xlog.Logger, count int, starttime,endtime int64, uid,uaid,mark string) ([]map[string]interface{},string, error) {
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
        if (mark != "") {
                marker = mark
        } else {
                marker = calculateMark(xl, starttime, uid,uaid, "seg")
        }
        prefix := "seg/" + uid + "/" + uaid + "/"
        xl.Infof("GetFragmentTsInfo prefix  %s \n", prefix)
        ctx, cancelFunc := context.WithCancel(context.Background())
        entries, err := bucketManager.ListBucketContext(ctx, bucket, prefix, delimiter, marker)
        if err != nil {
                xl.Errorf("GetFragmentTsInfo ListBucketContext %#v", err)
                info := err.(*storage.ErrorInfo)
                if (info.Code == 200) {
                       return r, "", nil
                } else {
                       return r, "", err
                }
        }
        for listItem1 := range entries {

                err, info := GetInfoFromFilename(listItem1.Item.Key, "/")
                if err != nil {
                        // if one file is not correct, continue to next
                        continue;
                }
                if (info[SEGMENT_ITEM_END_TIME].(int64) > endtime) {
                        cancelFunc()
                        break
                }
                if (info[SEGMENT_ITEM_START_TIME].(int64) > starttime) {
                        xl.Infof("GetFragmentTsInfo info[SEGMENT_ITEM_START_TIME] %d \n", info[SEGMENT_ITEM_START_TIME].(int64))
                        xl.Infof("GetFragmentTsInfo info[SEGMENT_ITEM_END_TIME] %d \n", info[SEGMENT_ITEM_END_TIME].(int64))
                        r = append(r, info)
                        total++
                }
                if (total >= count && count != 0) {
                        nextMarker = listItem1.Marker
                        break
                } 
        }
        xl.Infof("find fragment need %d ms\n", (time.Now().UnixNano() - pre) / 1000000)
        return r, nextMarker, err
}
