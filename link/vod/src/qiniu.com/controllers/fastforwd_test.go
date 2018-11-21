package controllers

import (
	"errors"
	"fmt"
	"github.com/bouk/monkey"
	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/suite"
	"net/http"
	"qiniu.com/models"
	"reflect"
	"testing"
)

type FastForwardTestSuite struct {
	suite.Suite
	r http.Handler
}

func (suite *FastForwardTestSuite) SetupTest() {
	suite.r = GetRouter()

}
func (suite *FastForwardTestSuite) TestGetFastForwardWithoutToken() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/fastforward?to=1536143287&from=1536142906&e=1536142906&speed=2", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(401, w.Code, "500 for query kodo data error")
}
func (suite *FastForwardTestSuite) TestGetFastForwardExpire() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/fastforward?to=1536143287&from=1536142906&e=1536142906&token=xxxxxxxxxxx002:xxxxx&speed=2", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) {
		return &userInfo{ak: "fadsfasfsd", sk: "fadsfasfsadf", uid: "123"}, nil, 200
	})
	w := PerformRequest(suite.r, req)
	suite.Equal(401, w.Code, "500 for query kodo data error")
}
func (suite *FastForwardTestSuite) TestGetFastForwardWithbadToken() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/fastforward?to=1536143287&from=1536142906&e=1536142906&token=xxxxxxxxxxx002:xxxxx&speed=2", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfoByAk, func(xl *xlog.Logger, req *http.Request) (*userInfo, error, int) {
		return &userInfo{ak: "fadsfasfsd", sk: "fadsfasfsadf", uid: "123"}, nil, 200
	})
	monkey.Patch(verifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request, user *userInfo) bool {
		return false
	})
	w := PerformRequest(suite.r, req)
	suite.Equal(401, w.Code, "500 for query kodo data error")
}

func (suite *FastForwardTestSuite) TestGetFastForwardStreamError() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/fastforward?to=1536143287&from=1536142906&e=1536142906&token=xxxxxxxx:xxxxxxspeed=2", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{ak: "fadsfasfsd", sk: "fadsfasfsadf", uid: "123"}, nil
	})
	monkey.Patch(verifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request, user *userInfo) bool {
		return true
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
		return "ipcamera", "www.baidu.com", nil
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
	monkey.Patch(getFastForwardStream, func(xl *xlog.Logger, params *requestParams, c *gin.Context, user *userInfo, bucket, domain, filename string) error {
		return errors.New("get Ts Stream Error")
	})
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "500 for query kodo data error")
}

func TestFastForwardSuite(t *testing.T) {
	suite.Run(t, new(FastForwardTestSuite))
}
