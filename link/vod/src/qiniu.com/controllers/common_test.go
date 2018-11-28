package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func GetRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/v1/namespaces/:namespace/uas/:uaid/live", GetLivem3u8)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/playback", GetPlayBackm3u8)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/segments", GetSegments)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/frames", GetFrames)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/fastforward", GetFastForward)
	r.POST("/upload", UploadTs)
	r.POST("/v1/namespaces/:namespace/uas/:uaid/saveas", Saveas)
	r.POST("/v1/namespaces/:namespace/uas/:uaid/store", MkStore)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/store", GetStoreList)
	r.DELETE("/v1/namespaces/:namespace/uas/:uaid/store", DeleteStoreList)
	r.PUT("/v1/namespaces/:namespace/uas/:uaid/store", UpdateStoreList)
	return r
}

func PerformRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestIsValidSpeed(t *testing.T) {
	assert.Equal(t, true, isValidSpeed(2), "2 should be valid")
	assert.Equal(t, true, isValidSpeed(4), "4 should be valid")
	assert.Equal(t, true, isValidSpeed(8), "8 should be valid")
	assert.Equal(t, true, isValidSpeed(16), "16 should be valid")
	assert.Equal(t, true, isValidSpeed(32), "32 should be valid")
	assert.Equal(t, true, isValidSpeed(1), "1 should be valid")

	assert.Equal(t, false, isValidSpeed(5), "5 should not valid")
	assert.Equal(t, false, isValidSpeed(6), "6 should not valid")
}

func TestParseRequset(t *testing.T) {
	url := "http://47.105.118.51:8088/v1/namespaces/ipcamera/uas/testdeviceid8/playback?from=1535500184&to=1535530184&speed=2&e=1536663799&token=JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ:7a3afxy3zT4STw5OKKX"
	req, _ := http.NewRequest("Get", url, nil)
	recoder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recoder)
	c.Request = req
	ret, err := ParseRequest(c, xlog.NewDummy())

	assert.Equal(t, nil, err)
	assert.Equal(t, int32(2), ret.speed)

	url = "http://47.105.118.51:8088/v1/namespaces/ipcamera/uas/testdeviceid8/playback?from=1535500184&to=1535530184&speed=5&e=1536663799&token=JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ:7a3afxy3zT4STw5OKKX"
	req, _ = http.NewRequest("Get", url, nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	c.Request = req
	ret, err = ParseRequest(c, xlog.NewDummy())

	assert.Equal(t, errors.New("Parse speed failed"), err)

	url = "http://47.105.118.51:8088/v1/namespaces/ipcamera/uas/testdeviceid8/playback?from=1535500184&to=1535530184&speed=a&e=1536663799&token=JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ:7a3afxy3zT4STw5OKKX"
	req, _ = http.NewRequest("Get", url, nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	c.Request = req
	ret, err = ParseRequest(c, xlog.NewDummy())

	assert.Equal(t, errors.New("Parse speed failed"), err)
}
