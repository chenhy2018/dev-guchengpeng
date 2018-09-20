package controllers

import (
	"bytes"
	"encoding/json"
	"github.com/bouk/monkey"
	"github.com/gin-gonic/gin"
	//"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	//"io/ioutil"
	"net/http"
	"net/http/httptest"
	"qiniu.com/system"
	"testing"
)

func TestAkSk(t *testing.T) {
	// test aksk api
	body := akskInfo{
		Ak: "ak1111",
		Sk: "sk2222",
	}
	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("Post", "/v1/aksk", bodyT)
	recoder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recoder)
	c.Request = req
	// If use qconf. we shouldn't set this api. report 403 error.
	monkey.Patch(system.HaveQconf, func() bool { return true })
	SetPrivateAkSk(c)
	assert.Equal(t, c.Writer.Status(), 403, "they should be equal")

	body = akskInfo{
		Ak: "ak1111",
		Sk: "sk2222",
	}
	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("Post", "/v1/aksk", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	c.Request = req
	defer monkey.UnpatchAll()
	monkey.Patch(system.HaveQconf, func() bool { return false })
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
}

func TestHandleToken(t *testing.T) {
	// test handle token api
	req, _ := http.NewRequest("Post", "/v1/aksk", nil)
	recoder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recoder)
	c.Request = req
	HandleToken(c)
	assert.Equal(t, c.Writer.Status(), 401, "they should be equal")
	req, _ = http.NewRequest("Post", "/v1/aksk", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	c.Request = req
	c.Request.Header.Set("Authorization", "QBox 11111111:12233")
	HandleToken(c)
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
	// Only set ak is also ok
	req, _ = http.NewRequest("Post", "/v1/aksk", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	c.Request = req
	c.Request.Header.Set("Authorization", "QBox 11111111")
	HandleToken(c)
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
	// Only set QBox will report 401
	req, _ = http.NewRequest("Post", "/v1/aksk", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	c.Request = req
	c.Request.Header.Set("Authorization", "QBox ")
	HandleToken(c)
	assert.Equal(t, c.Writer.Status(), 401, "they should be equal")
}
