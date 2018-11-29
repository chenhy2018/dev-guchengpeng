package controllers

import (
	"errors"
	"github.com/bouk/monkey"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/suite"
	"net/http"
	"qiniu.com/linking/vod.v1/models"
	"reflect"
	"testing"
)

type LiveTestSuite struct {
	suite.Suite
	r http.Handler
}

func (suite *LiveTestSuite) SetupTest() {
	suite.r = GetRouter()

}
func (suite *LiveTestSuite) TestLiveWithBadURL() {
	req, _ := http.NewRequest("GET", "/xxx/xx/xx", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(404, w.Code, "should be 404 for not implement bad url")
}

func (suite *LiveTestSuite) TestLiveWithoutUid() {
	req, _ := http.NewRequest("GET", "/live/xxxxx/123445", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(404, w.Code, "should be 404 for bad url")
}

func (suite *LiveTestSuite) TestLiveWithoutFrom() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/live", nil)
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
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) { return &userinfo, nil, 200 })
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no from requset")
}

func (suite *LiveTestSuite) TestLiveWithoutExpire() {
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) { return &userinfo, nil, 200 })
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/live?from=1532499345", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(401, w.Code, "should be 401 for no expire requset")
}

func (suite *LiveTestSuite) TestLiveWithoutToken() {
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) { return &userinfo, nil, 200 })
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/live?from=1532499345", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(401, w.Code, "should be 400 for no token")
}
func (suite *LiveTestSuite) TestLive500IfGetUseInfoFailed() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/live?from=1532499325", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) {
		return nil, errors.New("get user info error"), 500
	})
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "500 for get user info failed")
}
func (suite *LiveTestSuite) TestLive() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/live?from=1532499325", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(verifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request, user *userInfo) bool {
		return true
	})
	monkey.Patch(redisGet, func(key string) string {
		return "12345"
	})
	monkey.Patch(redisSet, func(xl *xlog.Logger, key, value string) error {
		return nil
	})
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) { return &userinfo, nil, 200 })
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
	suite.Equal(500, w.Code, "500 for nil")
}

func (suite *LiveTestSuite) TestLiveWithGetSegmentsInfoError() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/live?from=1532499325&to=1532499345", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(verifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request, user *userInfo) bool {
		return true
	})
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) { return &userinfo, nil, 200 })
	monkey.Patch(GetBucketAndDomain, func(xl *xlog.Logger, uid, namespace string) (string, string, error) {
		return "ipcamera", "www.baidu.com", nil
	})
	monkey.Patch(redisGet, func(key string) string {
		return "12345"
	})
	monkey.Patch(redisSet, func(xl *xlog.Logger, key, value string) error {
		return nil
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

func (suite *LiveTestSuite) TestLiveWithBadParam() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/live?to=1532499345&from=15324993x45", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(verifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request, user *userInfo) bool {
		return true
	})
	monkey.Patch(redisGet, func(key string) string {
		return "12345"
	})
	monkey.Patch(redisSet, func(xl *xlog.Logger, key, value string) error {
		return nil
	})
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) { return &userinfo, nil, 200 })
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no from requset")
}
func (suite *LiveTestSuite) TestLiveWithBadToken() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/live?to=1532499345&from=1532499045", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(verifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request, user *userInfo) bool {
		return true
	})
	monkey.Patch(redisGet, func(key string) string {
		return "12345"
	})
	monkey.Patch(redisSet, func(xl *xlog.Logger, key, value string) error {
		return nil
	})
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) { return &userinfo, nil, 200 })
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for bad token")
}

func (suite *LiveTestSuite) TestLiveWithBadNameSpace() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/live?to=1532499345&from=1532499045", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(verifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request, user *userInfo) bool {
		return true
	})
	monkey.Patch(redisGet, func(key string) string {
		return "12345"
	})
	monkey.Patch(redisSet, func(xl *xlog.Logger, key, value string) error {
		return nil
	})
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) { return &userinfo, nil, 200 })
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

func (suite *LiveTestSuite) TestLiveWithCorrectDomain() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/live?to=1536143287&from=1536142906", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(verifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request, user *userInfo) bool {
		return true
	})
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) {
		return &userInfo{ak: "fadsfasfsd", sk: "fadsfasfsadf", uid: "123"}, nil, 200
	})
	monkey.Patch(redisGet, func(key string) string {
		return "12345"
	})
	monkey.Patch(redisSet, func(xl *xlog.Logger, key, value string) error {
		return nil
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
	monkey.Patch(GetBucketAndDomain, func(xl *xlog.Logger, uid, namespace string) (string, string, error) {
		return "ipcamera", "", nil
	})

	monkey.Patch(uploadNewFile, func(filename, bucket string, data []byte, user *userInfo) error {
		return nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, uid, userAk string) ([]map[string]interface{}, string, error) {
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
			return info, "", nil
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(200, w.Code, "correct")
}

func TestLiveSuite(t *testing.T) {
	suite.Run(t, new(LiveTestSuite))
}
