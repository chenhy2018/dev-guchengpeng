package models

import (
	"fmt"
	"testing"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qiniu.com/db"
)

func TestNamespace(t *testing.T) {
	url := "mongodb://127.0.0.1:27017"
	dbName := "vod"
	config := db.MgoConfig{
		Host:     url,
		DB:       dbName,
		Mode:     "strong",
		Username: "root",
		Password: "public",
		AuthDB:   "admin",
		Proxies:  nil,
	}
	xl := xlog.NewDummy()
	xl.Infof("TestNamespace")
	db.InitDb(&config)
	assert.Equal(t, 0, 0, "they should be equal")
	model := NamespaceModel{}
	// Add ua, count size 100, from 0 to 100.
	for count := 0; count < 100; count++ {
		p := NamespaceInfo{
			Space:  fmt.Sprintf("test%03d", count),
			Bucket: "www.qiniu.com/112sss",
			Uid:    fmt.Sprintf("Uid"),
		}
		err := model.Register(xl, p)
		assert.Equal(t, err, nil, "they should be equal")
	}

	xl.Infof("DB Namespace Register done")
	// Get Namespace.
	r, err := model.GetNamespaceInfo(xl, "Uid", "test099")
	assert.Equal(t, err, nil, "they should be equal")
	assert.Equal(t, r[0].Space, "test099", "they should be equal")

	r_1, _, err_1 := model.GetNamespaceInfos(xl, 0, "", "Uid", "test099")
	assert.Equal(t, err_1, nil, "they should be equal")
	size_1 := len(r_1)
	assert.Equal(t, size_1, 1, "they should be equal")
	assert.Equal(t, r_1[0].Space, "test099", "they should be equal")

	r_1, next, err_1 := model.GetNamespaceInfos(xl, 2, "", "Uid", "test09")
	assert.Equal(t, err_1, nil, "they should be equal")
	size_1 = len(r_1)
	assert.Equal(t, size_1, 2, "they should be equal")
	assert.Equal(t, r_1[0].Space, "test090", "they should be equal")

	r_1, next, err_1 = model.GetNamespaceInfos(xl, 100, next, "Uid", "test09")
	assert.Equal(t, err_1, nil, "they should be equal")
	size_1 = len(r_1)
	assert.Equal(t, size_1, 8, "they should be equal")
	assert.Equal(t, r_1[0].Space, "test092", "they should be equal")

	for count := 0; count < 100; count++ {
		model.Delete(xl, "Uid", fmt.Sprintf("daaa%03d", count))
	}
}

func TestUpdateNamespace(t *testing.T) {
	xl := xlog.NewDummy()
	xl.Infof("TestUpdateNamespace")

	model := NamespaceModel{}
	p := NamespaceInfo{
		Space:  fmt.Sprintf("test"),
		Bucket: "www.qiniu.com/112sss",
		Uid:    fmt.Sprintf("Uid"),
	}
	err := model.Register(xl, p)
	assert.Equal(t, err, nil, "they should be equal")
	model.UpdateNamespace(xl, "Uid", "test", "test1")
	r_1, _, err_1 := model.GetNamespaceInfos(xl, 0, "", "Uid", "test1")
	assert.Equal(t, len(r_1), 1, "they should be equal")
	model.UpdateBucket(xl, "Uid", "test1", "www.qiniu.com")
	r_1, _, err_1 = model.GetNamespaceInfos(xl, 0, "", "Uid", "test1")
	assert.Equal(t, r_1[0].Bucket, "www.qiniu.com", "they should be equal")
	model.UpdateAutoCreateUa(xl, "Uid", "test1", true)
	r_1, _, err_1 = model.GetNamespaceInfos(xl, 0, "", "Uid", "test1")
	assert.Equal(t, r_1[0].AutoCreateUa, true, "they should be equal")
	model.UpdateExpire(xl, "Uid", "test1", 7)
	r_1, _, err_1 = model.GetNamespaceInfos(xl, 0, "", "Uid", "test1")
	assert.Equal(t, r_1[0].Expire, 7, "they should be equal")
	assert.Equal(t, err_1, nil, "they should be equal")
	model.Delete(xl, "Uid", "test1")
}
