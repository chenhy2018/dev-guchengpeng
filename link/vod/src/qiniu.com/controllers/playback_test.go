package controllers

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
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
	req, _ := http.NewRequest("GET", "/playback/12345??to=1532499345&e=1532499345&token=xxxxxx", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "should be 500 for no from requset")
}

func (suite *PlayBackTestSuite) TestPlayBackWithoutTo() {
	req, _ := http.NewRequest("GET", "/playback/12345??from=1532499345&e=1532499345&token=xxxxxx", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "should be 500 for no to requset")
}

func (suite *PlayBackTestSuite) TestPlayBackWithoutExpire() {
	req, _ := http.NewRequest("GET", "/playback/12345??from=1532499345&to=1532499345&token=xxxxxx", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "should be 500 for no from requset")
}

func (suite *PlayBackTestSuite) TestPlayBackWithoutToken() {
	req, _ := http.NewRequest("GET", "/playback/12345??from=1532499345&to=1532499345&e=1532499345", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "should be 500 for no token")
}

func (suite *PlayBackTestSuite) TestPlayBack() {
	req, _ := http.NewRequest("GET", "/playback/12345.m3u8?from=1532499325&to=1532499345&e=1532499345&token=13764829407:4ZNcW_AanSVccUmwq6MnA_8SWk8=", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(401, w.Code, "401 for bad token")
}

func TestPlayBackSuite(t *testing.T) {
	suite.Run(t, new(PlayBackTestSuite))
}
