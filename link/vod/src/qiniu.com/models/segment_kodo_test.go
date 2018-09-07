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
	infos, _, err4 := model.GetSegmentTsInfo(xl, int64(1535874867854), int64(1535878468023), bucket, "testdeviceid99", 0, "")
	assert.Equal(t, len(infos), 11, "they should be equal")
	/*
	   for i := 0; i < len(infos); i++ {
	           assert.Equal(t, infos[i][SEGMENT_ITEM_START_TIME].(int64), 0, "they should be equal")
	           assert.Equal(t, infos[i][SEGMENT_ITEM_END_TIME].(int64), 0, "they should be equal")
	   }
	*/
	assert.Equal(t, err4, nil, "they should be equal")

        // Reduce accuracy to seconds. also can find ts.
        // Include in 1535874867000-1535874874000
        infos, _, err4 = model.GetSegmentTsInfo(xl, int64(1535874867000), int64(1535874871000), bucket, "testdeviceid99", 0, "")
        assert.Equal(t, len(infos), 1, "they should be equal")

        infos, _, err4 = model.GetSegmentTsInfo(xl, int64(1535874869000), int64(1535874871000), bucket, "testdeviceid99", 0, "")
        assert.Equal(t, len(infos), 1, "they should be equal")

	infos, _, err4 = model.GetSegmentTsInfo(xl, int64(1535874869000), int64(1535874874000), bucket, "testdeviceid99", 0, "")
        assert.Equal(t, len(infos), 2, "they should be equal")

	// filename should be seg/ua_id/yyyy/mm/dd/hh/mm/ss/mmm/endts
	infoF, markF, errF := model.GetFragmentTsInfo(xl, 0, int64(1535874867854), int64(1535878468023), bucket, "testdeviceid99", "")
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 2, "they should be equal")
	infoF, markF, errF = model.GetFragmentTsInfo(xl, 1, int64(1535874867854), int64(1535878468023), bucket, "testdeviceid99", "")
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 1, "they should be equal")
	infoF, markF, errF = model.GetFragmentTsInfo(xl, 2, int64(1535874867854), int64(1535878468023), bucket, "testdeviceid99", markF)
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 1, "they should be equal")


	infoF, markF, errF = model.GetFragmentTsInfo(xl, 5, int64(1535874867854), int64(1535878468023), bucket, "testdeviceid99", markF)
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 2, "they should be equal")
	assert.Equal(t, infos[0][SEGMENT_ITEM_END_TIME].(int64), int64(1535874874060), "they should be equal")
	xl.Infof("1 2 mark %s", markF)


	infoF, errF = model.GetFrameInfo(xl, int64(1535874867854), int64(1535878468023), bucket, "testdeviceid99")
	assert.Equal(t, errF, nil, "they should be equal")
	assert.Equal(t, len(infoF), 11, "they should be equal")
}
