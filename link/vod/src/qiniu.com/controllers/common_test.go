package controllers

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func GetRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/playback/:deviceId", GetPlayBackm3u8)
	r.POST("/upload", UploadTs)
	return r
}

func PerformRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
