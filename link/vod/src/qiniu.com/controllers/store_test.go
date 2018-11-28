package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/suite"
	"qiniu.com/models"
)

type StoreTestSuite struct {
	suite.Suite
	r http.Handler
}

var (
	storeuser = userInfo{
		uid: "123",
		ak:  "xxxxxxxxxxx002",
		sk:  "xxxxxxxxxxx003",
	}
)

func (suite *StoreTestSuite) SetupTest() {
	suite.r = GetRouter()

}
func (suite *StoreTestSuite) TestStoreWithBadURL() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)

	req, _ := http.NewRequest("POST", "/xxx/xx/xx", bodyT)
	w := PerformRequest(suite.r, req)
	suite.Equal(404, w.Code, "should be 404 for not implement bad url")
}

func (suite *StoreTestSuite) TestStoreWithoutUid() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)

	req, _ := http.NewRequest("POST", "/store/xxxxx/123445", bodyT)
	w := PerformRequest(suite.r, req)
	suite.Equal(404, w.Code, "should be 404 for bad url")
}

func (suite *StoreTestSuite) TestStoreWithoutFrom() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)

	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?to=1532499345", bodyT)
	defer monkey.UnpatchAll()
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
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &storeuser, nil })
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no from requset")
}

func (suite *StoreTestSuite) TestStoreWithoutTo() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)

	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?from=1532499345", bodyT)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no to requset")
}

func (suite *StoreTestSuite) TestStoreWithoutExpire() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)

	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?from=1532499345&to=1532499345", bodyT)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no expire requset")
}

func (suite *StoreTestSuite) TestStoreWithoutToken() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?from=1532499345&to=1532499345", bodyT)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no token")
}
func (suite *StoreTestSuite) TestStore500IfGetUseInfoFailed() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?from=1532499325&to=1532499345", bodyT)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return nil, errors.New("get user info error")
	})
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "500 for get user info failed")
}
func (suite *StoreTestSuite) TestStore() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?from=1532499325&to=1532499345", bodyT)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &storeuser, nil })
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			info := []models.NamespaceInfo{}
			item := models.NamespaceInfo{
				Space:        "ipcamera",
				Regtime:      111111,
				UpdateTime:   3333333,
				Bucket:       "ipcamera",
				Uid:          "link",
				AutoCreateUa: true,
			}
			info = append(info, item)
			return info, nil
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
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, uid, userAk string) ([]map[string]interface{}, string, error) {
			return nil, "", nil
		})

	w := PerformRequest(suite.r, req)
	suite.Equal(404, w.Code, "404 for nil")
}

func (suite *StoreTestSuite) TestStoreWithGetSegmentsInfoError() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?from=1532499325&to=1532499345", bodyT)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &storeuser, nil })
	monkey.Patch(GetBucketAndDomain, func(xl *xlog.Logger, uid, namespace string) (string, string, error) {
		return "ipcamera", "", nil
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
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, uid, userAk string) ([]map[string]interface{}, string, error) {
			return nil, "", errors.New("get kodo data error")
		})

	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "500 for query kodo data error")
}

func (suite *StoreTestSuite) TestStoreWithBadParam() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?to=1532499345&from=15324993x45", bodyT)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &storeuser, nil })
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no from requset")
}
func (suite *StoreTestSuite) TestStoreWithBadToken() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?to=1532499345&from=1532499045", bodyT)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &storeuser, nil })
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for bad token")
}

func (suite *StoreTestSuite) TestStoreWithBadNameSpace() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?to=1532499345&from=1532499045", bodyT)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &storeuser, nil })
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
	monkey.Patch(GetBucketAndDomain, func(xl *xlog.Logger, uid, namespace string) (string, string, error) {
		return "", "", errors.New("bucket can't find")
	})

	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "bucket can't find")
}

func (suite *StoreTestSuite) TestMkStoreWithCorrectDomain() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?to=1536143287&from=1536142906", bodyT)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{ak: "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ", sk: "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS", uid: "1381539624"}, nil
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
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, uid, userAk string) ([]map[string]interface{}, string, error) {
			info := []map[string]interface{}{
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243756373),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243764350),
					models.SEGMENT_ITEM_FILE_NAME:  "ts/ipc10a/1537243756373/1537243764350/1537243645895/7.ts",
					models.SEGMENT_ITEM_FSIZE:      int64(123456),
				},
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243594011),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243601966),
					models.SEGMENT_ITEM_FILE_NAME:  "ts/ipc10a/1537243764284/1537243772224/1537243645895/7.ts",
					models.SEGMENT_ITEM_FSIZE:      int64(12345688),
				}}
			return info, "", nil
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(200, w.Code, "correct")
}

func (suite *StoreTestSuite) TestGetStoreWithCorrectDomain() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?to=1536143287&from=1536142906", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{ak: "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ", sk: "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS", uid: "1381539624"}, nil
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
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetStoreInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, limit int, starttime, endtime int64, bucketurl, namespace, uaid string, marker string, uid, userAk string) ([]map[string]interface{}, string, error) {
			info := []map[string]interface{}{
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243756373),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243764350),
					models.SEGMENT_ITEM_FILE_NAME:  "ipcamera/testdeviceid8/store/1536142906000/1536143287000/12469144.m3u8",
					models.SEGMENT_ITEM_FSIZE:      int64(12469144),
				},
			}
			return info, "", nil
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(200, w.Code, "correct")
}

/*
func (suite *StoreTestSuite) TestDelStoreWithCorrectDomain() {
	req, _ := http.NewRequest("DELETE", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?to=1536143287&from=1536142906", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{ak: "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ", sk: "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS", uid: "1381539624"}, nil
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

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetStoreInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, limit int, starttime, endtime int64, bucketurl, uaid string, marker string, uid, userAk string) ([]map[string]interface{}, string, error) {
			info := []map[string]interface{}{
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243756373),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243764350),
					models.SEGMENT_ITEM_FILE_NAME:  "ipcamera/testdeviceid8/store/1536142906000/1536143287000/12469144.m3u8",
					models.SEGMENT_ITEM_FSIZE:      int64(12469144),
				},
			}
			return info, "", nil
		})

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, uid, userAk string) ([]map[string]interface{}, string, error) {
			info := []map[string]interface{}{
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243756373),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243764350),
					models.SEGMENT_ITEM_FILE_NAME:  "ts/ipc10a/1537243756373/1537243764350/1537243645895/7.ts",
					models.SEGMENT_ITEM_FSIZE:      int64(123456),
				},
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243594011),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243601966),
					models.SEGMENT_ITEM_FILE_NAME:  "ts/ipc10a/1537243764284/1537243772224/1537243645895/7.ts",
					models.SEGMENT_ITEM_FSIZE:      int64(12345688),
				}}
			return info, "", nil
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(200, w.Code, "correct")
}
*/

func (suite *StoreTestSuite) TestUpdateStoreWithCorrectDomain() {
	body := storeArgs{
		ExpireDays: 7,
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("PUT", "/v1/namespaces/ipcamera/uas/testdeviceid8/store?to=1536143287&from=1536142906", bodyT)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{ak: "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ", sk: "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS", uid: "1381539624"}, nil
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

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetStoreInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, limit int, starttime, endtime int64, bucketurl, namespace, uaid string, marker string, uid, userAk string) ([]map[string]interface{}, string, error) {
			info := []map[string]interface{}{
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243756373),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243764350),
					models.SEGMENT_ITEM_FILE_NAME:  "ipcamera/testdeviceid8/store/1536142906000/1536143287000/12469144.m3u8",
					models.SEGMENT_ITEM_FSIZE:      int64(12469144),
				},
			}
			return info, "", nil
		})

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, uid, userAk string) ([]map[string]interface{}, string, error) {
			info := []map[string]interface{}{
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243756373),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243764350),
					models.SEGMENT_ITEM_FILE_NAME:  "ts/ipc10a/1537243756373/1537243764350/1537243645895/7.ts",
					models.SEGMENT_ITEM_FSIZE:      int64(123456),
				},
				map[string]interface{}{
					models.SEGMENT_ITEM_START_TIME: int64(1537243594011),
					models.SEGMENT_ITEM_END_TIME:   int64(1537243601966),
					models.SEGMENT_ITEM_FILE_NAME:  "ts/ipc10a/1537243764284/1537243772224/1537243645895/7.ts",
					models.SEGMENT_ITEM_FSIZE:      int64(12345688),
				}}
			return info, "", nil
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(200, w.Code, "correct")
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
