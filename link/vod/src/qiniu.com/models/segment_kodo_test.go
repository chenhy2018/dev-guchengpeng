package models

import (
	"testing"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func TestKodoSegment(t *testing.T) {
	xl := xlog.NewDummy()
	xl.Infof("Test kodo segment")
	model := SegmentKodoModel{}
	model.Init()
	bucket := "ipcamera"
	// filename should be ts/ua_id/startts/endts/fragment_start_ts/expiry.ts
	filename := "ts/ua1/1514952732111/155321111/15332111222/123.ts"
	_, seg := GetInfoFromFilename(filename, "/")

	assert.Equal(t, seg[SEGMENT_ITEM_START_TIME], int64(1514952732111), "they should be equal")
	assert.Equal(t, seg[SEGMENT_ITEM_END_TIME], int64(155321111), "they should be equal")
	assert.Equal(t, seg[SEGMENT_ITEM_FRAGMENT_START_TIME], int64(15332111222), "they should be equal")
	assert.Equal(t, seg[SEGMENT_ITEM_FILE_NAME], filename, "they should be equal")
	assert.Equal(t, seg[SEGMENT_ITEM_EXPIRE], int64(123), "they should be equal")

	// Test get 2018/8/16 - 2018/8/17 segment count. It should be 2767.
	xl.Infof("SegmentTsInfo 1")
	infos, _, err4 := model.GetSegmentTsInfo(xl, int64(1534385684000), int64(1534472084000), bucket, "testdeviceid5", 0, "")
	assert.Equal(t, len(infos), 2767, "they should be equal")
	/*
	   for i := 0; i < len(infos); i++ {
	           assert.Equal(t, infos[i][SEGMENT_ITEM_START_TIME].(int64), 0, "they should be equal")
	           assert.Equal(t, infos[i][SEGMENT_ITEM_END_TIME].(int64), 0, "they should be equal")
	   }
	*/
	assert.Equal(t, err4, nil, "they should be equal")

	infos, _, err4 = model.GetSegmentTsInfo(xl, int64(1534848606000), int64(1534859606000), bucket, "testdeviceid10", 0, "")

	//infos, err4 = model.GetSegmentTsInfo(xl, int64(1534414484000), int64(1534500884000), bucket,"testdeviceid10")
	//assert.Equal(t, err4, nil, "they should be equal")
	//assert.Equal(t, 12026, len(infos), "they should be equal")

	// filename should be seg/ua_id/yyyy/mm/dd/hh/mm/ss/mmm/endts
	infoF, markF, errF := model.GetFragmentTsInfo(xl, 0, int64(1534385684000), int64(1534472084000), bucket, "testdeviceid5", "")
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 11, "they should be equal")
	infoF, markF, errF = model.GetFragmentTsInfo(xl, 1, int64(1534385684000), int64(1534472084000), bucket, "testdeviceid5", "")
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 1, "they should be equal")
	infoF, markF, errF = model.GetFragmentTsInfo(xl, 2, int64(1534385684000), int64(1534472084000), bucket, "testdeviceid5", markF)
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 2, "they should be equal")

	infoF, markF, errF = model.GetFragmentTsInfo(xl, 10, int64(1534385684000), int64(1534472084000), bucket, "testdeviceid5", markF)
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 8, "they should be equal")

	infoF, markF, errF = model.GetFragmentTsInfo(xl, 0, int64(1534443284000), int64(1534500884000), bucket, "testdeviceid10", "")
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 9, "they should be equal")

	infoF, markF, errF = model.GetFragmentTsInfo(xl, 1, int64(1534412096000), int64(1534729559000), bucket, "testdeviceid10", "")
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 1, "they should be equal")
	xl.Infof("1 0 mark %s", markF)
	infoF, markF, errF = model.GetFragmentTsInfo(xl, 1, int64(1534412096000), int64(1534729559000), bucket, "testdeviceid10", markF)
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 1, "they should be equal")
	//assert.Equal(t, infos[0][SEGMENT_ITEM_END_TIME].(int64), int64(1534414490692), "they should be equal")
	xl.Infof("1 1 mark %s", markF)

	infoF, markF, errF = model.GetFragmentTsInfo(xl, 5, int64(1534412096000), int64(1534729559000), bucket, "testdeviceid10", markF)
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 5, "they should be equal")
	//assert.Equal(t, infos[0][SEGMENT_ITEM_END_TIME].(int64), int64(1534414490692), "they should be equal")
	xl.Infof("1 2 mark %s", markF)

	infoF, markF, errF = model.GetFragmentTsInfo(xl, 7, int64(1534412096000), int64(1534729559000), bucket, "testdeviceid10", markF)
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 3, "they should be equal")

	infoF, errF = model.GetFrameInfo(xl, int64(1535358330000), int64(1535359621000), bucket, "testdeviceid6")
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 3, "they should be equal")
}
