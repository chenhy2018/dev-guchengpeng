package models
  
import (
        "fmt"
        "strconv"
        "context"
        "github.com/qiniu/xlog.v1"
        "github.com/qiniu/api.v7/auth/qbox"
        "github.com/qiniu/api.v7/storage"
        "time"
        "strings"
)

var (
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
//                        SEGMENT_ITEM_FILE_NAME : s,
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

// fragment filename should be 7/uid/ua_id/start_ts/end_ts.ts
func GetOldInfoFromFilename(s, sep string) (error, map[string]interface{}) {
        sub := strings.Split(s, sep)
        var info map[string]interface{}
        //uid := sub[1]
        //uaid := sub[2]
        starttime, err1 := strconv.ParseInt(sub[3], 10, 64)
        if (err1 != nil) {
                 return err1, info
        }
        end := strings.Split(sub[4], ".")
        endtime, err1 := strconv.ParseInt(end[0], 10, 64)
        if (err1 != nil) {
                 return err1, info
        }
        info = map[string]interface{} {
                        SEGMENT_ITEM_FRAGMENT_START_TIME : starttime,
                        SEGMENT_ITEM_START_TIME : starttime,
                        SEGMENT_ITEM_END_TIME : endtime,
                        SEGMENT_ITEM_FILE_NAME : s,
        }
        return nil, info
}

// Calculate prefix starttime list.
// Return []yyyy/mm/dd, if same day and same hour, return [1]yyyy/mm/dd/hh
func CalculatePrefixList(xl *xlog.Logger, starttime,endtime int64) ([]string) {
        var str []string
        starttm := time.Unix(starttime / 1000 , starttime % 1000 * 1000000)
        
        endtm := time.Unix(endtime / 1000 , endtime % 1000 * 1000000)
        fmt.Printf("end time %04d/%02d/%02d  %d\n", endtm.Year(), endtm.Month(), endtm.Day(), endtime)
        var prefix string
        if (starttm.Year() == endtm.Year() &&  starttm.Month() == endtm.Month() && starttm.Day() == endtm.Day()) {
                if (starttm.Hour() == endtm.Hour()) {
                        prefix = fmt.Sprintf("%04d/%02d/%02d/%02d", starttm.Year(), starttm.Month(), starttm.Day(), starttm.Hour())
                } else {
                        prefix = fmt.Sprintf("%04d/%02d/%02d", starttm.Year(), starttm.Month(), starttm.Day())
                }
                str = append(str, prefix)
                return str
        } else {
                for {
                        prefix = fmt.Sprintf("%04d/%02d/%02d", starttm.Year(), starttm.Month(), starttm.Day())
                        fmt.Printf("%04d/%02d/%02d \n", starttm.Year(), starttm.Month(), starttm.Day())
                        str = append(str, prefix)
                        if (starttm.Year() == endtm.Year() &&  starttm.Month() == endtm.Month() && starttm.Day() == endtm.Day()) {
                                return str
                        }
                        
                        starttm = starttm.AddDate(0, 0, 1)
                }
        }
        return str
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
        marker := ""
        prefix := "ts/" + uid + "/" + uaid + "/"
        prefix1 := CalculatePrefixList(xl, starttime, endtime)
        for count := 0; count < len(prefix1); count++ {
                ctx, cancelFunc := context.WithCancel(context.Background())
                xl.Infof("GetSegmentTsInfo prefix %s \n", prefix + prefix1[count])
                entries, err := bucketManager.ListBucketContext(ctx, bucket, prefix + prefix1[count], delimiter, marker)
                if err != nil {
                        xl.Errorf("GetSegmentTsInfo ListBucketContext %#v", err)
                        info := err.(*storage.ErrorInfo)
                        if (info.Code == 200) {
                                continue
                        } else {
                                return r, err
                        }
                }

                for listItem1 := range entries {
                        err, info := GetInfoFromFilename(listItem1.Item.Key, "/")
                        if err != nil {
                                cancelFunc()
                                return r, err
                        }
		        if len(info) == 0 {
			        continue
		        }
                        if (info[SEGMENT_ITEM_END_TIME].(int64) > endtime) {
                                cancelFunc()
                                break;
                        }
                        if (info[SEGMENT_ITEM_START_TIME].(int64) > starttime) {
                                r = append(r, info)
                        }
                }
        }
        xl.Infof("find segment need %d ms", (time.Now().UnixNano() - pre) / 1000000)
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
        num := 0
        marker := ""
        total := 0
        sub := strings.Split(mark, "/")
        if (len(sub) == 2) {
                index, err := strconv.Atoi(sub[0])
                if (err != nil) {
                        xl.Errorf("parse marker error, marke = %#v, err = %#v", mark, err)
                        num = 0
                        marker = ""
                } else {
                        num = index
                        marker = sub[1]
                }
        } else if (len(sub) == 1) {
                index, err := strconv.Atoi(sub[0])
                if (err != nil) {
                        num = 0
                        marker = ""
                } else {
                        num = index
                }
        }

        prefix := "seg/" + uid + "/" + uaid + "/"
        prefix1 := CalculatePrefixList(xl, starttime, endtime)

        for ; num < len(prefix1); num++ {
                xl.Infof("GetFragmentTsInfo prefix  %s \n", prefix + prefix1[num])
                ctx, cancelFunc := context.WithCancel(context.Background())
                entries, err := bucketManager.ListBucketContext(ctx, bucket, prefix + prefix1[num], delimiter, marker)
                marker = ""
                if err != nil {
                        xl.Errorf("GetFragmentTsInfo ListBucketContext %#v", err)
                        info := err.(*storage.ErrorInfo)
                        if (info.Code == 200) {
                                continue
                        } else {
                                return r, "", err
                        }
                }
                for listItem1 := range entries {

                        err, info := GetInfoFromFilename(listItem1.Item.Key, "/")
                        if err != nil {
                                cancelFunc()
                                return r, "", err
                        }
                        if (info[SEGMENT_ITEM_END_TIME].(int64) > endtime) {
                                cancelFunc()
                                break;
                        }
                        if (info[SEGMENT_ITEM_START_TIME].(int64) > starttime) {
                                r = append(r, info)
                                total++
                        }
                        if (total >= count && count != 0) {
                                if (listItem1.Marker == "") {
                                        num++
                                }
                                next := fmt.Sprintf("%d/%s", num, listItem1.Marker)
                                return r, next, nil
                        } 
                }
        }
        xl.Infof("find fragment need %d ms\n", (time.Now().UnixNano() - pre) / 1000000)
        return r, "",  nil
}
