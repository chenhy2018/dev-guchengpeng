package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterUa(t *testing.T) {

	initDb()

	// register name space.  bucket maybe already exist. so not check this response.
	bodyN := namespacebody{
		Uid:       "link",
		Bucket:    "ipcamera",
		Namespace: "test1",
	}

	bodyBuffer, _ := json.Marshal(bodyN)
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

	// register ua
	// bucket maybe already exist. so not check this response.
	body := uabody{
		Uid:       "link",
		Uaid:      "ipcamera1",
		Namespace: "test1",
	}

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

	body = uabody{
		Uid:       "link",
		Uaid:      "ipcamera",
		Namespace: "test1",
	}

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
		Value: "ipcamera1",
	}
	c.Params = append(c.Params, param)
	c.Request = req
	RegisterUa(c)
}

func TestGetUa(t *testing.T) {
	initDb()
	req, _ := http.NewRequest("Get", "/v1/namespaces/test1/uas?regex=ipcamera1&limit=1&marker=&exact=true", nil)
	recoder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recoder)
	c.Request = req
	param := gin.Param{
		Key:   "namespace",
		Value: "test1",
	}
	c.Params = append(c.Params, param)
	GetUaInfo(c)
	fmt.Printf("write %v \n", recoder.Body)
	body, err := ioutil.ReadAll(recoder.Body)
	if err != nil {
		fmt.Errorf("parse request body failed, body = %#v", body)
	}
	fmt.Printf("%s \n", body)
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
	fmt.Printf("write %v \n", recoder.Body)
	body, err = ioutil.ReadAll(recoder.Body)
	if err != nil {
		fmt.Errorf("parse request body failed, body = %#v", body)
	}
	fmt.Printf("%s \n", body)
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
	fmt.Printf("write %v \n", recoder.Body)
	body, err = ioutil.ReadAll(recoder.Body)
	if err != nil {
		fmt.Errorf("parse request body failed, body = %#v", body)
	}
	fmt.Printf("%s \n", body)
	//{"item":[],"marker":""}
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")
}

func TestUpdateUa(t *testing.T) {
	initDb()
	// Change namespace to aab, return 400
	body := uabody{
		Uid:       "link",
		Uaid:      "ipcamera2",
		Namespace: "aab",
	}

	bodyBuffer, _ := json.Marshal(body)
	bodyT := bytes.NewBuffer(bodyBuffer)

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
	body = uabody{
		Uid:       "link",
		Uaid:      "ipcamera3",
		Namespace: "test1",
	}

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
	body = uabody{
		Uid:       "link",
		Uaid:      "ipcamera1",
		Namespace: "test1",
	}

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
	body = uabody{
		Uid:       "link",
		Uaid:      "aaaaa",
		Namespace: "test1",
	}

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
}

func TestDeleteUa(t *testing.T) {
	initDb()
	// remove invaild namespace aab, return 400
	req, _ := http.NewRequest("Del", "/v1/namespaces/aab/uas/ipcamera2", nil)
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
	assert.Equal(t, c.Writer.Status(), 200, "they should be equal")

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
}
