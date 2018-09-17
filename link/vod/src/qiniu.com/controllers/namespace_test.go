package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bouk/monkey"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"qiniu.com/db"
	"qiniu.com/system"
	"testing"
)

var (
	tcontext gin.Context
	user     = userInfo{
		uid: 1,
		ak:  "JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ",
		sk:  "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS",
	}
)

func initDb() {
	url := "mongodb://127.0.0.1:27017"
	dbName := "vod"
	config := db.MgoConfig{
		Host:     url,
		DB:       dbName,
		Mode:     "strong",
		Username: "",
		Password: "",
		AuthDB:   "",
		Proxies:  nil,
	}
	db.InitDb(&config)
}

func TestRegisterNamespace(t *testing.T) {

	initDb()
	// bucket maybe already exist. so not check this response.
	body := namespacebody{
		Bucket:    "ipcamera",
		Namespace: "test1",
	}

	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	param := gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	monkey.Patch(system.HaveDb, func() bool { fmt.Printf("111111"); return true })
	RegisterNamespace(c)

	// bucket already exit. return 400
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterNamespace(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// bucket is not correct. return 403
	body = namespacebody{
		Bucket:    "ipcamera1",
		Namespace: "aabb",
	}
	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	RegisterNamespace(c)
	assert.Equal(t, c.Writer.Status(), 403, "they should be equal")

	// body is not correct. return 403
	/*
	   body = namespacebody{
	   }
	*/
	body1 := "asddhjk"
	bodyBuffer, _ = json.Marshal(body1)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterNamespace(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")
}

func TestGetNamespace(t *testing.T) {
	initDb()
	req, _ := http.NewRequest("Get", "/v1/namespaces?regex=test1&limit=10&marker=&exact=true", nil)
	recoder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recoder)
	c.Request = req
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	monkey.Patch(system.HaveDb, func() bool { return true })
	GetNamespaceInfo(c)
	body, err := ioutil.ReadAll(recoder.Body)
	if err != nil {
		fmt.Printf("parse request body failed, body = %#v", body)
	}
	//{"item":[{"namespace":"test1","createdAt":1535539324,"updatedAt":1535539324,"bucket":"ipcamera","uid":"link","domain":"pdwjeyj6v.bkt.clouddn.com"}],"marker":""}
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
	//assert.Equal(t, body, bodye, "they should be equal")

	req, _ = http.NewRequest("Get", "/v1/namespaces?regex=test&limit=10&marker=&exact=true", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	c.Request = req
	GetNamespaceInfo(c)
	body, err = ioutil.ReadAll(recoder.Body)
	if err != nil {
		fmt.Printf("parse request body failed, body = %#v", body)
	}
	//{"item":[],"marker":""}
	bodye := []uint8([]byte{0x7b, 0x22, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x3a, 0x5b, 0x5d, 0x2c, 0x22, 0x6d, 0x61, 0x72, 0x6b, 0x65, 0x72, 0x22, 0x3a, 0x22, 0x22, 0x7d})
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
	assert.Equal(t, body, bodye, "they should be equal")

	req, _ = http.NewRequest("Get", "/v1/namespaces?regex=test&limit=10&marker=&exact=false", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	c.Request = req
	GetNamespaceInfo(c)
	body, err = ioutil.ReadAll(recoder.Body)
	if err != nil {
		fmt.Printf("parse request body failed, body = %#v", body)
	}
	//{"item":[],"marker":""}
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
	//assert.Equal(t, body, bodye, "they should be equal")
}

func TestUpdateNamespace(t *testing.T) {
	initDb()
	// bucket maybe already exit. so not check this response.
	body := namespacebody{
		Bucket:    "ipcamera",
		Namespace: "aab",
	}

	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)

	req, _ := http.NewRequest("Put", "/v1/namespaces/test1", bodyT)
	recoder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recoder)
	c.Request = req
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	monkey.Patch(system.HaveDb, func() bool { return true })
	// namespace is not exist
	param := gin.Param{
		Key:   "namespace",
		Value: "test",
	}
	c.Params = append(c.Params, param)

	UpdateNamespace(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// Change namespace to aab
	body = namespacebody{
		Bucket:    "ipcamera",
		Namespace: "aab",
	}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/test1", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateNamespace(c)
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")

	// Change invaild bucket, return 403
	body = namespacebody{
		Bucket:    "ipcamera1",
		Namespace: "aab",
	}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/aab", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "aab",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateNamespace(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// invaild body. return 400
	body1 := "asddhjk"
	bodyBuffer, _ = json.Marshal(body1)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/aab", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "aab",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateNamespace(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// Change namespace to test1
	body = namespacebody{
		Bucket:    "ipcamera",
		Namespace: "test1",
	}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/aab", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "aab",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateNamespace(c)
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
}

func TestDeleteNamespace(t *testing.T) {
	initDb()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	monkey.Patch(system.HaveDb, func() bool { return true })
	// remove invaild namespace aab, return 400
	req, _ := http.NewRequest("Put", "/v1/namespaces/aab", nil)
	recoder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recoder)
	param := gin.Param{
		Key:   "namespace",
		Value: "aab",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	DeleteNamespace(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// remove namespace test1, return 200

	req, _ = http.NewRequest("Put", "/v1/namespaces/test1", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	DeleteNamespace(c)
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
}

func TestAutoCreateUa(t *testing.T) {
	initDb()
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	monkey.Patch(system.HaveDb, func() bool { return true })
	xl := xlog.NewDummy()
	// bucket maybe already exist. so not check this response.
	body := namespacebody{
		Bucket:    "ipcamera",
		Namespace: "test1",
	}

	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	param := gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	RegisterNamespace(c)
	check, _, _ := IsAutoCreateUa(xl, "ipcamera")
	assert.Equal(t, check, false, "they should be equal")
	// bucket maybe already exit. so not check this response.
	body = namespacebody{
		Bucket:       "ipcamera",
		Namespace:    "test1",
		AutoCreateUa: true,
	}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/test1", bodyT)
	recoder := httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateNamespace(c)
	check, _, _ = IsAutoCreateUa(xl, "ipcamera")
	assert.Equal(t, check, true, "they should be equal")
	req, _ = http.NewRequest("Put", "/v1/namespaces/test1", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	DeleteNamespace(c)
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
}

func TestHandleUaControl(t *testing.T) {
	initDb()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	monkey.Patch(system.HaveDb, func() bool { return true })
	xl := xlog.NewDummy()
	err := HandleUaControl(xl, "ipcamera", "test1")
	assert.Equal(t, err.Error(), "can't find namespace", "they should be equal")

	// bucket maybe already exist. so not check this response.
	body := namespacebody{
		Bucket:    "ipcamera",
		Namespace: "test1",
	}

	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	param := gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	RegisterNamespace(c)
	err = HandleUaControl(xl, "ipcamera", "test1a")

	assert.Equal(t, err.Error(), "Can't find ua info", "they should be equal")
	// bucket maybe already exit. so not check this response.
	body = namespacebody{
		Bucket:       "ipcamera",
		Namespace:    "test1",
		AutoCreateUa: true,
	}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/test1", bodyT)
	recoder := httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateNamespace(c)
	err = HandleUaControl(xl, "ipcamera", "test1a")
	assert.Equal(t, err, nil, "they should be equal")

	req, _ = http.NewRequest("Delete", "/v1/namespaces/test1", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	DeleteNamespace(c)
}
