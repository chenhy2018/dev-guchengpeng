package models

import (
	"testing"

	"github.com/qiniu/api.v7/storage"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qiniu.com/system"
)

func TestKodoSegment(t *testing.T) {
	type listFilesRet2 struct {
		Marker string           `json:"marker"`
		Item   storage.ListItem `json:"item"`
		Dir    string           `json:"dir"`
	}
	xl := xlog.NewDummy()
	xl.Infof("Test kodo segment")
	model := SegmentKodoModel{}
	model.Init(system.UserConf{})
	// filename should be ts/ua_id/startts/endts/fragment_start_ts/expiry.ts
	filename := "ts/ua1/1514952732111/155321111/15332111222/123.ts"
	_, seg := GetInfoFromFilename(filename, "/")

	assert.Equal(t, seg[SEGMENT_ITEM_START_TIME], int64(1514952732111), "they should be equal")
	assert.Equal(t, seg[SEGMENT_ITEM_END_TIME], int64(155321111), "they should be equal")
	assert.Equal(t, seg[SEGMENT_ITEM_FRAGMENT_START_TIME], int64(15332111222), "they should be equal")
	assert.Equal(t, seg[SEGMENT_ITEM_FILE_NAME], filename, "they should be equal")
	assert.Equal(t, seg[SEGMENT_ITEM_EXPIRE], int64(123), "they should be equal")
}
