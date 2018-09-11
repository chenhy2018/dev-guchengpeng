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

type PlayBackTestSuite struct {
	suite.Suite
	r http.Handler
}

func (suite *PlayBackTestSuite) SetupTest() {
	suite.r = GetRouter()

}
func (suite *PlayBackTestSuite) TestPlayBackWithBadURL() {
	req, _ := http.NewRequest("GET", "/xxx/xx/xx", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(404, w.Code, "should be 404 for not implement bad url")
}

func (suite *PlayBackTestSuite) TestPlayBackWithoutUid() {
	req, _ := http.NewRequest("GET", "/playback/xxxxx/123445", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(404, w.Code, "should be 404 for bad url")
}

func (suite *PlayBackTestSuite) TestPlayBackWithoutFrom() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/playback?to=1532499345&e=1532499345&token=xxxxxx", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(VerifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request) bool { return true })
	w := PerformRequest(suite.r, req)
	//suite.Equal(true, testObj.VerifyToken(xlog.NewDummy(), int64(10), "", "", ""), "successxxx")
	suite.Equal(400, w.Code, "should be 400 for no from requset")
}

func (suite *PlayBackTestSuite) TestPlayBackWithoutTo() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/playback?from=1532499345&e=1532499345&token=xxxxxx", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no to requset")
}

func (suite *PlayBackTestSuite) TestPlayBackWithoutExpire() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/playback?from=1532499345&to=1532499345&token=xxxxxx", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no expire requset")
}

func (suite *PlayBackTestSuite) TestPlayBackWithoutToken() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/playback?from=1532499345&to=1532499345&e=1532499345", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "should be 400 for no token")
}
func (suite *PlayBackTestSuite) TestPlayBack500IfGetUseInfoFailed() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/playback?from=1532499325&to=1532499345&e=1532499345&token=13764829407:4ZNcW_AanSVccUmwq6MnA_8SWk8=", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(VerifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request) bool { return true })
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return nil, errors.New("get user info error")
	})
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "500 for get user info failed")
}
func (suite *PlayBackTestSuite) TestPlayBack() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/playback?from=1532499325&to=1532499345&e=1532499345&token=13764829407:4ZNcW_AanSVccUmwq6MnA_8SWk8=", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(VerifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request) bool { return true })
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &userInfo{}, nil })
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string) ([]map[string]interface{}, string, error) {
			return nil, "", nil
		})

	w := PerformRequest(suite.r, req)
	suite.Equal(404, w.Code, "404 for nil")
}

func (suite *PlayBackTestSuite) TestPlayBackWithGetSegmentsInfoError() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/playback?from=1532499325&to=1532499345&e=1532499345&token=13764829407:4ZNcW_AanSVccUmwq6MnA_8SWk8=", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(VerifyToken, func(xl *xlog.Logger, expire int64, realToken string, req *http.Request) bool { return true })
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string) ([]map[string]interface{}, string, error) {
			return nil, "", errors.New("get kodo data error")
		})

	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "500 for query kodo data error")
}
func TestPlayBackSuite(t *testing.T) {
	suite.Run(t, new(PlayBackTestSuite))
}
