package models
  
import (
        "fmt"
        //"strconv"
        "testing"
        "github.com/qiniu/xlog.v1"
        "github.com/stretchr/testify/assert"
        "time"
        //"strings"
)

func TestKodoSegment(t *testing.T) {
        xl := xlog.NewDummy()
        xl.Infof("Test kodo segment")
        model := SegmentKodoModel{}
        model.Init()

        xl.Infof("Test TransferTimeToInt64") 
        str := []string{ "2018", "01", "02", "12", "12", "12", "033"}
        err, starttime := TransferTimeToInt64(str)
        assert.Equal(t, err, nil, "they should be equal")
        tm := time.Unix(starttime / 1000 , starttime % 1000 * 1000000)
        assert.Equal(t, tm.Format(time.StampMilli), "Jan  2 20:12:12.033", "they should be equal")

        str = []string{ "2018", "12", "09", "11", "11", "01", "555"}
        err, starttime = TransferTimeToInt64(str)
        assert.Equal(t, err, nil, "they should be equal")
        tm = time.Unix(starttime / 1000 , starttime % 1000 * 1000000)
        assert.Equal(t, tm.Format(time.StampMilli), "Dec  9 19:11:01.555", "they should be equal")

        str = []string{ "2018", "13", "09", "11", "11", "01", "555"}
        err, starttime = TransferTimeToInt64(str)
        assert.EqualError(t, err, "start time:  parser month is error [[2018 13 09 11 11 01 555]]", "they should be equal")
        //tm = time.Unix(starttime / 1000 , starttime % 1000 * 1000000)
        //assert.Equal(t, tm.Format(time.StampMilli), "Dec  9 19:11:01.555", "they should be equal")

        xl.Infof("Test TransferTimeToString")
        sstarttime := TransferTimeToString(int64(1533783079679))
        assert.Equal(t, sstarttime, "2018/08/09/10/51/19/679", "they should be equal")
 
        sstarttime = TransferTimeToString(int64(1533783076490))
        assert.Equal(t, sstarttime, "2018/08/09/10/51/16/490", "they should be equal")

        sstarttime = TransferTimeToString(int64(1533783076509))
        assert.Equal(t, sstarttime, "2018/08/09/10/51/16/509", "they should be equal")

        // filename should be ts/uid/ua_id/yyyy/mm/dd/hh/mm/ss/mmm/endts/fragment_start_ts/expiry.ts
        filename := "ts/ua/ua1/2018/01/3/12/12/12/111/155321111/15332111222/123.ts"
        err3, seg := GetInfoFromFilename(filename, "/")
        assert.Equal(t, err3, nil, "they should be equal")
        assert.Equal(t, seg[SEGMENT_ITEM_START_TIME], int64(1514981532111), "they should be equal")
        assert.Equal(t, seg[SEGMENT_ITEM_END_TIME], int64(155321111), "they should be equal")
        assert.Equal(t, seg[SEGMENT_ITEM_FRAGMENT_START_TIME], int64(15332111222), "they should be equal")
        assert.Equal(t, seg[SEGMENT_ITEM_FILE_NAME], filename, "they should be equal")
        assert.Equal(t, seg[SEGMENT_ITEM_EXPIRE], int64(123), "they should be equal")

        infos, err4 := model.GetSegmentTsInfo(xl, 0, 0, int64(1533783079678), int64(1533783076489), "testuid5", "testdeviceid5")
        assert.Equal(t, len(infos), 20, "they should be equal")
/*
        for i := 0; i < len(infos); i++ {
                assert.Equal(t, infos[i][SEGMENT_ITEM_START_TIME].(int64), 0, "they should be equal")
                assert.Equal(t, infos[i][SEGMENT_ITEM_END_TIME].(int64), 0, "they should be equal")
        }
*/
        assert.Equal(t, err4, nil, "they should be equal")

        xl.Infof("Test CalculatePrefixList")
        arr := CalculatePrefixList(int64(1533783079678), int64(1534893076489))
        for i := 0; i < len(arr); i++ {
                test := fmt.Sprintf("2018/08/%02d", 9+i);
                assert.Equal(t, arr[i], test, "they should be equal")
        }
}
