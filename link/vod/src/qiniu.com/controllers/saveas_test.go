package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/qiniu/api.v7/storage"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/suite"
	"qiniu.com/models"
)

type SaveasTestSuite struct {
	suite.Suite
	r http.Handler
}

var (
	suser = userInfo{
		uid: "1381539624",
		ak:  "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ",
		sk:  "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS",
	}
)

func (suite *SaveasTestSuite) SetupTest() {
	suite.r = GetRouter()

}
func (suite *SaveasTestSuite) TestSaveasWithBadURL() {
	req, _ := http.NewRequest("", "/xxx/xx/xx", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(404, w.Code, "should be 404 for not implement bad url")
}

func (suite *SaveasTestSuite) TestSaveasWithoutUid() {
	body := saveasArgs{
		Format:     "mp4",
		Fname:      "1.mp4",
		Pipeline:   "ipcamera",
		Notify:     "www.baidu.com/api",
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/playback/xxxxx/123445", bodyT)
	w := PerformRequest(suite.r, req)
	suite.Equal(404, w.Code, "should be 404 for bad url")
}

func (suite *SaveasTestSuite) TestSaveasWithoutFrom() {
	body := saveasArgs{
		Format:     "mp4",
		Fname:      "1.mp4",
		Pipeline:   "ipcamera",
		Notify:     "www.baidu.com/api",
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/saveas?to=1532499345", bodyT)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &suser, nil })
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no from requset")
}

func (suite *SaveasTestSuite) TestSaveasWithoutTo() {
	body := saveasArgs{
		Format:     "mp4",
		Fname:      "1.mp4",
		Pipeline:   "ipcamera",
		Notify:     "",
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/saveas?from=1532499345", bodyT)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no to requset")
}

func (suite *SaveasTestSuite) TestSaveasWithoutToken() {
	body := saveasArgs{
		Format:     "mp4",
		Fname:      "1.mp4",
		Pipeline:   "ipcamera",
		Notify:     "www.baidu.com/api",
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/saveas?from=1532499345&to=1532499345", bodyT)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no token")
}
func (suite *SaveasTestSuite) TestSaveas() {
	body := saveasArgs{
		Format:     "mp4",
		Fname:      "test_test.mp4",
		Pipeline:   "",
		Notify:     "",
		ExpireDays: 1,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/saveas?to=1536143287&from=1536142906", bodyT)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &suser, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "1381539624",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucketAndDomain, func(xl *xlog.Logger, uid, namespace string) (string, string, error) {
		return "ipcamera", "pdwjeyj6v.bkt.clouddn.com", nil
	})

	monkey.Patch(uploadNewFile, func(filename, bucket string, data []byte, user *userInfo) error {
		return nil
	})

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*storage.OperationManager)(nil)), "Pfop", func(ss *storage.OperationManager, bucket, key, fops, pipeline, notifyURL string, force bool) (string, error) {
			return "12345", nil
		})

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, uid, userAk string) ([]map[string]interface{}, string, error) {
			info := []map[string]interface{}{
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243756373),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243764350),
					models.SEGMENT_ITEM_FILE_NAME:  "ts/ipc10a/1537243756373/1537243764350/1537243645895/7.ts",
					models.SEGMENT_ITEM_FSIZE:      123456,
				},
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243594011),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243601966),
					models.SEGMENT_ITEM_FILE_NAME:  "ts/ipc10a/1537243764284/1537243772224/1537243645895/7.ts",
					models.SEGMENT_ITEM_FSIZE:      12345678,
				}}
			return info, "", nil
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(200, w.Code, "correct")
}

func TestSaveasSuite(t *testing.T) {
	suite.Run(t, new(SaveasTestSuite))
}
