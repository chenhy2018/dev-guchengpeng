package models

import (
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qiniu.com/linking/vod.v1/db"
	"testing"
	"time"
)

func TestSegment(t *testing.T) {
	xl := xlog.NewDummy()
	xl.Infof("Test segment")
	url := "mongodb://127.0.0.1:27017"
	dbName := "vod"
	config := db.MgoConfig{
		Host:     url,
		DB:       dbName,
		Mode:     "",
		Username: "",
		Password: "",
		AuthDB:   "",
		Proxies:  nil,
	}
	db.InitDb(&config)
	assert.Equal(t, 0, 0, "they should be equal")
	model := SegmentModel{}
	model.Init()
	model.DeleteSegmentTS(xl, "UserTest", "dev001", 0, 200)
	// Add first frament, count size 100, start time 0 to end time 100.
	for count := 0; count < 100; count++ {
		p := SegmentTsInfo{
			Uid:               "UserTest",
			UaId:              "dev001",
			FragmentStartTime: 0,
			StartTime:         int64(count),
			EndTime:           int64(count + 1),
			FileName:          "test1",
			Expire:            time.Now().Add(3600 * time.Second),
		}
		err := model.AddSegmentTS(xl, p)
		assert.Equal(t, err, nil, "they should be equal")
	}
	// Add first frament, count size 100, start time 100 to end time 200.
	for count := 100; count < 200; count++ {
		p := SegmentTsInfo{
			Uid:               "UserTest",
			UaId:              "dev001",
			FragmentStartTime: 100,
			StartTime:         int64(count),
			EndTime:           int64(count + 1),
			FileName:          "test1",
			Expire:            time.Now().Add(3600 * time.Second),
		}
		err := model.AddSegmentTS(xl, p)
		assert.Equal(t, err, nil, "they should be equal")
	}
	last, err_0 := model.GetLastSegmentTsInfo(xl, "UserTest", "dev001")
	assert.Equal(t, err_0, nil, "they should be equal")
	assert.Equal(t, last["starttime"], int64(199), "they should be equal")
	assert.Equal(t, last["endtime"], int64(200), "they should be equal")

	// Get segment from start time 0 to end time 150.
	r, err := model.GetSegmentTsInfo(xl, int64(0), int64(150), "UserTest", "dev001")
	assert.Equal(t, err, nil, "they should be equal")
	size := len(r)
	assert.Equal(t, size, 150, "they should be equal")

	// Get segment from start time 0 to end time 150 by fragment.
	r1, mark, err1 := model.GetFragmentTsInfo(xl, 0, int64(0), int64(150), "UserTest", "dev001", "0")
	assert.Equal(t, err1, nil, "they should be equal")
	assert.Equal(t, mark, "", "they should be equal")
	size1 := len(r1)
	assert.Equal(t, size1, 2, "they should be equal")
	assert.Equal(t, r1[0]["starttime"], int64(0))
	assert.Equal(t, r1[0]["endtime"], int64(100))
	assert.Equal(t, r1[1]["starttime"], int64(100))
	assert.Equal(t, r1[1]["endtime"], int64(200))

	r1, mark, err1 = model.GetFragmentTsInfo(xl, 1, int64(0), int64(150), "UserTest", "dev001", "0")
	assert.Equal(t, err1, nil, "they should be equal")
	assert.Equal(t, mark, "1", "they should be equal")
	assert.Equal(t, len(r1), 1, "they should be equal")

	r1, mark, err1 = model.GetFragmentTsInfo(xl, 2, int64(0), int64(150), "UserTest", "dev001", mark)
	assert.Equal(t, err1, nil, "they should be equal")
	assert.Equal(t, mark, "", "they should be equal")
	assert.Equal(t, len(r1), 1, "they should be equal")

	// Get segment from start time 0 to end time 150. only get 0 - 50 count.
	r_1, err_1 := model.GetSegmentTsInfo(xl, int64(0), int64(150), "UserTest", "dev001")
	assert.Equal(t, err_1, nil, "they should be equal")
	size_1 := len(r_1)
	assert.Equal(t, size_1, 150, "they should be equal")
	for count := 0; count < 150; count++ {
		assert.Equal(t, r_1[count]["starttime"], int64(count), "they should be equal")
		assert.Equal(t, r_1[count]["endtime"], int64(count+1), "they should be equal")
	}
	//derr := model.DeleteSegmentTS("UserTest", "dev001", 0, 200)
	//assert.Equal(t, derr, nil, "they should be equal")
	//r_4, err_4 := model.GetSegmentTsInfo(xl, int64(0), int64(150), "UserTest", "dev001")
	//assert.Equal(t, err_4, nil, "they should be equal")
	//size_4 := len(r_4)
	//assert.Equal(t, size_4, 0, "they should be equal")
}
