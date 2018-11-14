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
	"qiniu.com/models"
)

func TestRegisterUa(t *testing.T) {

	initDb()

	// register name space.  bucket maybe already exist. so not check this response.
	bodyN := namespacebody{
		Bucket: "ipcamera",
	}

	bodyBuffer, _ := json.Marshal(bodyN)
	bodyT := bytes.NewBuffer(bodyBuffer)
	req, _ := http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	param := gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	monkey.Patch(checkBucketInKodo, func(bucket string, user *userInfo) error { return nil })

	c.Params = append(c.Params, param)
	c.Request = req
	RegisterNamespace(c)

	// register ua

	bodyBuffer, _ = json.Marshal(bodyN)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1/uas/ipcamera1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	RegisterUa(c)

	// namespace is not correct. return 400
	req, _ = http.NewRequest("POST", "/v1/namespaces/test/uas/ipcamera1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// body is not correct. return 403
	body1 := "asddhjk"
	bodyBuffer, _ = json.Marshal(body1)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	body := uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test1/uas/ipcamera1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	RegisterUa(c)

	// 500 if get user info failed
	guard1 := monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, errors.New("get user  info error")
	})
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test/uas/ipcamera1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterUa(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	guard1.Unpatch()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	// get name space info failed
	guard2 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return nil, errors.New("xxxxx error")
		})
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test/uas/ipcamera1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterUa(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	guard2.Unpatch()

	// get name space info failed 2
	guard3 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return []models.NamespaceInfo{}, nil
		})
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test/uas/ipcamera1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")
	guard3.Unpatch()

	// get Ua info failed
	guard4 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namesapce, uaid string) ([]models.UaInfo, error) {
			return []models.UaInfo{}, errors.New("get ua  info failed")
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return []models.NamespaceInfo{models.NamespaceInfo{}}, nil
		})
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test/uas/ipcamera1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterUa(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	guard4.Unpatch()

	// ua is exist
	guard5 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namesapce, uaid string) ([]models.UaInfo, error) {
			return []models.UaInfo{models.UaInfo{}}, nil
		})
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test/uas/ipcamera1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")
	guard5.Unpatch()

	// register ua failed
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namesapce, uaid string) ([]models.UaInfo, error) {
			return []models.UaInfo{}, nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "Register", func(ss *models.UaModel, xl *xlog.Logger, ua models.UaInfo) error {
			return errors.New("register ua failed")
		})
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)
	req, _ = http.NewRequest("POST", "/v1/namespaces/test/uas/ipcamera1", bodyT)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	RegisterUa(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	monkey.UnpatchAll()

}

func TestGetUa(t *testing.T) {
	initDb()
	req, _ := http.NewRequest("Get", "/v1/namespaces/test1/uas?regex=ipcamera1&limit=1&marker=&exact=true", nil)
	recoder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recoder)
	c.Request = req
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })

	param := gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	GetUaInfo(c)
	body, err := ioutil.ReadAll(recoder.Body)
	if err != nil {
		fmt.Printf("parse request body failed, body = %#v", body)

	}
	//{"item":[{"namespace":"test1","createdAt":1535539324,"updatedAt":1535539324,"bucket":"ipcamera","uid":"link","domain":"pdwjeyj6v.bkt.clouddn.com"}],"marker":""}
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")

	req, _ = http.NewRequest("Get", "/v1/namespaces/test1/uas?regex=ipcamera2&limit=1&marker=&exact=true", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	GetUaInfo(c)
	body, err = ioutil.ReadAll(recoder.Body)
	if err != nil {
		fmt.Printf("parse request body failed, body = %#v", body)

	}
	//{"item":[],"marker":""}
	bodye := []uint8([]byte{0x7b, 0x22, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x3a, 0x5b, 0x5d, 0x2c, 0x22, 0x6d, 0x61, 0x72, 0x6b, 0x65, 0x72, 0x22, 0x3a, 0x22, 0x22, 0x7d})
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
	assert.Equal(t, body, bodye, "they should be equal")

	req, _ = http.NewRequest("Get", "/v1/namespaces?regex=ip&limit=10&marker=&exact=false", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	GetUaInfo(c)
	body, err = ioutil.ReadAll(recoder.Body)
	if err != nil {
		fmt.Printf("parse request body failed, body = %#v", body)

	}
	//{"item":[],"marker":""}
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")

	// get ua info failed
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namesapce, uaid string) ([]models.UaInfo, error) {
			return []models.UaInfo{}, errors.New("get ua info failed")
		})
	req, _ = http.NewRequest("Get", "/v1/namespaces?regex=ip&limit=10&marker=&exact=true", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	GetUaInfo(c)
	body, err = ioutil.ReadAll(recoder.Body)
	if err != nil {
		fmt.Printf("parse request body failed, body = %#v", body)
	}
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	monkey.UnpatchAll()

}

func TestUpdateUa(t *testing.T) {
	initDb()
	// Change namespace to aab, return 400
	body := uabody{}

	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	req, _ := http.NewRequest("Put", "/v1/namespaces/test1/uas/ipcamera2", bodyT)
	recoder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recoder)
	c.Request = req
	// namespace is not exist
	param := gin.Param{
		Key:   "namespace",
		Value: "test",
	}
	c.Params = append(c.Params, param)

	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)

	UpdateUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// Change uaid invaild ipcamera2 to ipcamera3, return 400.
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/test1/uas/ipcamera2", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)

	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera2",
	}
	c.Params = append(c.Params, param)

	c.Request = req
	UpdateUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// invaild namespace, return 400.
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/aab/uas/ipcamera2", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "aab",
	}
	c.Params = append(c.Params, param)

	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)

	c.Request = req
	UpdateUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// invaild body. return 400
	body1 := "asddhjk"
	bodyBuffer, _ = json.Marshal(body1)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/aab/uas/ipcamera2", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "aab",
	}
	c.Params = append(c.Params, param)

	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)

	c.Request = req
	UpdateUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// Change ua to aaaaa
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/test1/uas/ipcamera1", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)

	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateUa(c)
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")

	// get userinfo failed
	guard1 := monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, errors.New("get user  info error")
	})
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/test1/uas/ipcamera1", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)

	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateUa(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	guard1.Unpatch()

	// get namesapce info failed
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	// get name space info failed
	guard2 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return nil, errors.New("xxxxx error")
		})
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/test1/uas/ipcamera1", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)

	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateUa(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	guard2.Unpatch()

	// get ua info failed
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			return []models.NamespaceInfo{models.NamespaceInfo{}}, nil
		})
	guard4 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namesapce, uaid string) ([]models.UaInfo, error) {
			return []models.UaInfo{}, errors.New("get ua  info failed")
		})
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/test1/uas/ipcamera1", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)

	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")
	guard4.Unpatch()

	// update ua failed
	guard7 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namesapce, uaid string) ([]models.UaInfo, error) {
			if uaid == "ipcamera1" {
				return []models.UaInfo{models.UaInfo{}}, nil
			}
			return []models.UaInfo{}, nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "UpdateUa", func(ss *models.UaModel, xl *xlog.Logger, uid, namesapce, uaid string, ua models.UaInfo) error {
			return errors.New("update ua failed")
		})
	body = uabody{}

	bodyBuffer, _ = json.Marshal(body)
	bodyT = bytes.NewBuffer(bodyBuffer)

	req, _ = http.NewRequest("Put", "/v1/namespaces/test1/uas/ipcamera1", bodyT)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)

	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	UpdateUa(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	guard7.Unpatch()
	monkey.UnpatchAll()

}

func TestDeleteUa(t *testing.T) {
	initDb()
	// remove uid ipcamer1, return 400
	fmt.Printf("remove invaild uid ipcamer1, return 400. \n")
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	req, _ := http.NewRequest("Del", "/v1/uas/ipcamera1", nil)
	recoder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recoder)
	param := gin.Param{
		Key:   "namespace",
		Value: "aab",
	}
	c.Params = append(c.Params, param)
	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	DeleteUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// remove ua aaaaa, return 200

	req, _ = http.NewRequest("Del", "/v1/namespaces/test1/uas/aaaaa", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	param = gin.Param{
		Key:   "uaid",
		Value: "aaaaa",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	DeleteUa(c)
	assert.Equal(t, c.Writer.Status(), 400, "they should be equal")

	// remove ua ipcamera1, return 200

	req, _ = http.NewRequest("Del", "/v1/namespaces/test1/uas/ipcamera1", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	param = gin.Param{
		Key:   "uaid",
		Value: "ipcamera",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	DeleteUa(c)
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")

	// remove namespace test1, return 200
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) { return &user, nil })
	req, _ = http.NewRequest("Del", "/v1/namespaces/test1", nil)
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

	// get userinfo failed
	guard1 := monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namesapce, uaid string) ([]models.UaInfo, error) {
			return []models.UaInfo{}, errors.New("get ua  info failed")
		})
	req, _ = http.NewRequest("Del", "/v1/namespaces/test1", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	DeleteUa(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	guard1.Unpatch()

	// delete ua failed
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namesapce, uaid string) ([]models.UaInfo, error) {
			return []models.UaInfo{models.UaInfo{}}, nil

		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "Delete", func(ss *models.UaModel, xl *xlog.Logger, cond map[string]interface{}) error {
			return errors.New("delete ua failed")
		})
	req, _ = http.NewRequest("Del", "/v1/namespaces/test1", nil)
	recoder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recoder)
	param = gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	DeleteUa(c)
	assert.Equal(t, c.Writer.Status(), 500, "they should be equal")
	monkey.UnpatchAll()
}
