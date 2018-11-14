package controllers

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/suite"
	"qiniu.com/models"
)

type FramesTestSuite struct {
	suite.Suite
	r http.Handler
}

func (suite *FramesTestSuite) SetupSuite() {
	suite.r = GetRouter()
}

func (suite *FramesTestSuite) TearDownSuite() {
	monkey.UnpatchAll()
}
func (suite *FramesTestSuite) TestFrames1() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/frames?from=1532499fdsfs325&to=1532499345", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "400 for request params error")

}
func (suite *FramesTestSuite) TestFrames2() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/frames?from=1532499355&to=1532499345", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "400 for request from great than to")

}
func (suite *FramesTestSuite) TestFrames3() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/frames?from=1530499355&to=1532499345", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "400 for request from great than to")

}

func (suite *FramesTestSuite) TestFrames4() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/frames?from=1537872668&to=1537876443", nil)
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, errors.New("get user info failed")
	})
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "500 for get user info  failed")

}
func (suite *FramesTestSuite) TestFrames5() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/frames?from=1537872668&to=1537876443", nil)
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userinfo, nil
	})
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "400 for can't find namesapce")
}

func (suite *FramesTestSuite) TestFrames6() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/frames?from=1537872668&to=1537876443", nil)
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userinfo, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFrameInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, starttime int64, endtime int64, bucket string, uaid string, uid, userAk string) ([]map[string]interface{}, error) {
			info := []map[string]interface{}{
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1536142906000),
					models.SEGMENT_ITEM_END_TIME:   int64(1536143141000),
					models.SEGMENT_ITEM_FILE_NAME:  "ts/ipc00a/1537856214961/1537856214961/7.ts",
				},
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1536143141000),
					models.SEGMENT_ITEM_END_TIME:   int64(1536143280000),
					models.SEGMENT_ITEM_FILE_NAME:  "ts/ipc00a/1537856214961/1537856214961/7.ts",
				}}
			return info, nil
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(200, w.Code, "")
	monkey.UnpatchAll()
}
func (suite *FramesTestSuite) TestFrames7() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/frames?from=1537872668&to=1537876443", nil)
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userinfo, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFrameInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, starttime int64, endtime int64, bucket string, uaid string, uid, userAk string) ([]map[string]interface{}, error) {
			return nil, errors.New("can't  find frames")
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "")
	monkey.UnpatchAll()
}
func (suite *FramesTestSuite) TestFrames9() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/frames?from=1537872668&to=1537876443", nil)
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userinfo, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFrameInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, starttime int64, endtime int64, bucket string, uaid string, uid, userAk string) ([]map[string]interface{}, error) {
			return nil, nil
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(200, w.Code, "")
	monkey.UnpatchAll()
}
func TestFramesTestSuite(t *testing.T) {
	suite.Run(t, new(FramesTestSuite))
}
