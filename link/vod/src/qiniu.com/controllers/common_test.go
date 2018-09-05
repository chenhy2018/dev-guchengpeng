package controllers

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func GetRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/v1/namespaces/:namespace/uas/:uaid/playback", GetPlayBackm3u8)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/segments", GetSegments)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/frames", GetFrames)
	r.POST("/upload", UploadTs)
	return r
}

func PerformRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
