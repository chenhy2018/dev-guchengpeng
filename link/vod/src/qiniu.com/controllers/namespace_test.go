package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qiniu.com/db"
	"qiniu.com/models"
	"qiniu.com/system"
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
		Bucket: "ipcamera",
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
	monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	monkey.Patch(system.HaveDb, func() bool { fmt.Printf("111111"); return true })
	RegisterNamespace(c)

	// bucket already exit. return 400
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	body = namespacebody{
		Bucket: "ipcamera",
	}
	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Params = append(c.Params, param)
	c.Request = req
	RegisterNamespace(c)
	assert.Equal(t, c.Writer.Status(), 403, "they should be equal")

	// get namespace  info error
	body = namespacebody{
		Bucket: "doman",
	}
	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	guard := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return nil, errors.New("xxxxx error")
		})
	RegisterNamespace(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	guard.Unpatch()
	// namesapce already exist
	body = namespacebody{
		Bucket: "doman",
	}
	guard2 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return []models.NamespaceInfo{models.NamespaceInfo{}}, nil
		})
	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterNamespace(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")
	guard2.Unpatch()
	// get user info failed. return 500
	guard4 := monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, errors.New("get user  info error")
	})
	body = namespacebody{
		Bucket: "ipcamera10",
	}
	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterNamespace(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	guard4.Unpatch()
	// body is not correct. return 403
	/*
	   body = namespacebody{
	   }
	*/
	// namesapce already exist
	guard3 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "Register", func(ss *models.NamespaceModel, xl *xlog.Logger, namespace models.NamespaceInfo) error {
			return errors.New("register namesapce failed")
		})
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	body = namespacebody{
		Bucket: "doman",
	}
	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterNamespace(c)
	assert.Equal(t, c.Writer.Status(), 500, "internal error for register namesapce failed")
	guard3.Unpatch()

	body1 := "asddhjk"
	bodyBuffer, _ = json.Marshal(body1)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterNamespace(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")
	monkey.UnpatchAll()
}

func TestGetNamespace(t *testing.T) {
	initDb()
	monkey.UnpatchAll()
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

	// 500 internal error if get userinfo failed
	guard1 := monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, errors.New("get user  info error")
	})
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
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	guard1.Unpatch()

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return nil, errors.New("xxxxx error")
		})
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, errors.New("get user  info error")
	})
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
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	monkey.UnpatchAll()
}

func TestUpdateNamespace(t *testing.T) {
	initDb()
	monkey.UnpatchAll()
	// bucket maybe already exit. so not check this response.
	body := namespacebody{
		Bucket: "ipcamera",
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
		Bucket: "ipcamera",
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
		Bucket: "ipcamera1",
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

	// get user info failed if get user info failed
	guard1 := monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, errors.New("get user  info error")
	})
	body = namespacebody{
		Bucket: "ipcamera",
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
	assert.Equal(t, c.Writer.Status(), 500, "500 internal error if get user info error")
	guard1.Unpatch()

	// 500 if update namesapce error
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return []models.NamespaceInfo{models.NamespaceInfo{}}, nil
		})
	guard2 := monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &user, nil
	})
	body = namespacebody{
		Bucket: "ipcamera",
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
	assert.Equal(t, c.Writer.Status(), 400, "400 not support update namesapce")
	guard2.Unpatch()

	// get user info failed if get update autocraeteUa error
	guard4 := monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &user, nil
	})

	guard6 := monkey.Patch(updateAutoCreateUa, func(xl *xlog.Logger, uid, space string, auto, newauto bool) error {
		return errors.New("update auto create ua failed")
	})
	monkey.Patch(updateBucket, func(xl *xlog.Logger, uid, space, bucket, newBucket string, info *userInfo) error {
		return nil
	})
	body = namespacebody{
		Bucket: "ipcamera",
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
	assert.Equal(t, c.Writer.Status(), 500, "500 internal error if get namesapce info error")
	guard4.Unpatch()
	guard6.Unpatch()

	monkey.UnpatchAll()
}

func TestDeleteNamespace(t *testing.T) {
	initDb()
	monkey.UnpatchAll()
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

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return []models.NamespaceInfo{models.NamespaceInfo{}}, nil
		})
	// 500 internal error if get user info failed
	guard1 := monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return nil, errors.New("get user info failed")
	})
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
	assert.Equal(t, c.Writer.Status(), 500, "internal error if get user info failed")
	guard1.Unpatch()

	// 500 internal error if get namespace info error
	guard2 := monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &user, nil
	})
	guard3 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return nil, errors.New("get namespace error")
		})
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
	assert.Equal(t, c.Writer.Status(), 500, "internal error if get user info failed")
	guard2.Unpatch()
	guard3.Unpatch()

	// 500 internal error if get namespace info error
	guard4 := monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &user, nil
	})
	guard5 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return []models.NamespaceInfo{models.NamespaceInfo{}}, nil
		})
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
	assert.Equal(t, c.Writer.Status(), 500, "internal error if get user info failed")
	guard4.Unpatch()
	guard5.Unpatch()

	monkey.UnpatchAll()
}

func TestAutoCreateUa(t *testing.T) {
	initDb()
	monkey.UnpatchAll()
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	monkey.Patch(system.HaveDb, func() bool { return true })
	xl := xlog.NewDummy()
	// bucket maybe already exist. so not check this response.
	body := namespacebody{
		Bucket: "ipcamera",
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

func TestGetNameSpaceInfo(t *testing.T) {
	initDb()
	monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	monkey.Patch(system.HaveDb, func() bool { return true })
	xl := xlog.NewDummy()
	err, _ := GetNameSpaceInfo(xl, "ipcamera", "test1")
	assert.Equal(t, err.Error(), "can't find namespace", "they should be equal")

	// bucket maybe already exist. so not check this response.
	body := namespacebody{
		Bucket: "ipcamera",
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
	err, _ = GetNameSpaceInfo(xl, "ipcamera", "test1a")

	assert.Equal(t, err.Error(), "Can't find ua info", "they should be equal")
	// bucket maybe already exit. so not check this response.
	body = namespacebody{
		Bucket:       "ipcamera",
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
	err, _ = GetNameSpaceInfo(xl, "ipcamera", "test1a")
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
