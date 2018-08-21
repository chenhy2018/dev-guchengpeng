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
        assert.Equal(t, tm.Format(time.StampMilli), "Jan  2 12:12:12.033", "they should be equal")

        str = []string{ "2018", "12", "09", "11", "11", "01", "555"}
        err, starttime = TransferTimeToInt64(str)
        assert.Equal(t, err, nil, "they should be equal")
        tm = time.Unix(starttime / 1000 , starttime % 1000 * 1000000)
        assert.Equal(t, tm.Format(time.StampMilli), "Dec  9 11:11:01.555", "they should be equal")

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
        assert.Equal(t, seg[SEGMENT_ITEM_START_TIME], int64(1514952732111), "they should be equal")
        assert.Equal(t, seg[SEGMENT_ITEM_END_TIME], int64(155321111), "they should be equal")
        assert.Equal(t, seg[SEGMENT_ITEM_FRAGMENT_START_TIME], int64(15332111222), "they should be equal")
        assert.Equal(t, seg[SEGMENT_ITEM_FILE_NAME], filename, "they should be equal")
        assert.Equal(t, seg[SEGMENT_ITEM_EXPIRE], int64(123), "they should be equal")

        // Test get 2018/8/16 - 2018/8/17 segment count. It should be 2767.
        xl.Infof("SegmentTsInfo 1")
        infos, err4 := model.GetSegmentTsInfo(xl, int64(1534385684000), int64(1534472084000), "testuid5", "testdeviceid5")
        assert.Equal(t, len(infos), 2767, "they should be equal")
/*
        for i := 0; i < len(infos); i++ {
                assert.Equal(t, infos[i][SEGMENT_ITEM_START_TIME].(int64), 0, "they should be equal")
                assert.Equal(t, infos[i][SEGMENT_ITEM_END_TIME].(int64), 0, "they should be equal")
        }
*/
        assert.Equal(t, err4, nil, "they should be equal")

        xl.Infof("Test CalculatePrefixList")
        arr := CalculatePrefixList(xl, int64(1533783079678), int64(1534893076489))
        for i := 0; i < len(arr); i++ {
                test := fmt.Sprintf("2018/08/%02d", 9+i);
                assert.Equal(t, arr[i], test, "they should be equal")
        }

        infos, err4 = model.GetSegmentTsInfo(xl, int64(1534414484000), int64(1534500884000), "testuid10", "testdeviceid10")
        assert.Equal(t, err4, nil, "they should be equal")
        assert.Equal(t, 12026, len(infos), "they should be equal")
        // filename should be seg/uid/ua_id/yyyy/mm/dd/hh/mm/ss/mmm/endts
        infoF, markF,  errF := model.GetFragmentTsInfo(xl, 0, int64(1534385684000), int64(1534472084000), "testuid5", "testdeviceid5", "")
        assert.Equal(t, errF, nil, "they should be equal")
        assert.Equal(t, len(infoF), 11, "they should be equal")

        infoF, markF,  errF = model.GetFragmentTsInfo(xl, 1, int64(1534385684000), int64(1534472084000), "testuid5", "testdeviceid5", "")
        assert.Equal(t, errF, nil, "they should be equal")
        assert.Equal(t, len(infoF), 1, "they should be equal")
        infoF, markF,  errF = model.GetFragmentTsInfo(xl, 2, int64(1534385684000), int64(1534472084000), "testuid5", "testdeviceid5", markF)
        assert.Equal(t, errF, nil, "they should be equal")
        assert.Equal(t, len(infoF), 2, "they should be equal")

        infoF, markF,  errF = model.GetFragmentTsInfo(xl, 10, int64(1534385684000), int64(1534472084000), "testuid5", "testdeviceid5", markF)
        assert.Equal(t, errF, nil, "they should be equal")
        assert.Equal(t, len(infoF), 8, "they should be equal")

        infoF, markF,  errF = model.GetFragmentTsInfo(xl, 0, int64(1534443284000), int64(1534500884000), "testuid10", "testdeviceid10", "")
        assert.Equal(t, errF, nil, "they should be equal")
        assert.Equal(t, len(infoF), 9, "they should be equal")

        infoF, markF,  errF = model.GetFragmentTsInfo(xl, 1, int64(1534412096000), int64(1534729559000), "testuid10", "testdeviceid10", "")
        assert.Equal(t, errF, nil, "they should be equal")
        assert.Equal(t, len(infoF), 1, "they should be equal")
        xl.Infof("1 0 mark %s", markF)
        infoF, markF,  errF = model.GetFragmentTsInfo(xl, 1, int64(1534412096000), int64(1534729559000), "testuid10", "testdeviceid10", markF)
        assert.Equal(t, errF, nil, "they should be equal")
        assert.Equal(t, len(infoF), 1, "they should be equal")
        assert.Equal(t, infos[0][SEGMENT_ITEM_END_TIME].(int64), int64(1534414490692), "they should be equal")
        xl.Infof("1 1 mark %s", markF)

        infoF, markF,  errF = model.GetFragmentTsInfo(xl, 5, int64(1534412096000), int64(1534729559000), "testuid10", "testdeviceid10", markF)
        assert.Equal(t, errF, nil, "they should be equal")
        assert.Equal(t, len(infoF), 5, "they should be equal")
        assert.Equal(t, infos[0][SEGMENT_ITEM_END_TIME].(int64), int64(1534414490692), "they should be equal")
        xl.Infof("1 2 mark %s", markF)

        infoF, markF,  errF = model.GetFragmentTsInfo(xl, 7, int64(1534412096000), int64(1534729559000), "testuid10", "testdeviceid10", markF)
        assert.Equal(t, errF, nil, "they should be equal")
        assert.Equal(t, len(infoF), 3, "they should be equal")
}
