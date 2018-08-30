package models

import (
	"fmt"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qiniu.com/db"
	"testing"
	"time"
)

func TestUa(t *testing.T) {
	url := "mongodb://root:public@180.97.147.164:27017,180.97.147.179:27017/admin"
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
	xl.Infof("TestUa")
	db.InitDb(&config)
	assert.Equal(t, 0, 0, "they should be equal")
	model := UaModel{}
	// Add ua, count size 100, from 0 to 100.
	for count := 0; count < 100; count++ {
		p := UaInfo{
			Uid:       "UserTest",
			UaId:      fmt.Sprintf("daaa%03d", count),
			Namespace: "test",
			Password:  "112sss",
		}
		err := model.Register(xl, p)
		assert.Equal(t, err, nil, "they should be equal")
	}
	xl.Infof("DB Register done")
	// Get ua.
	r, _, err := model.GetUaInfos(xl, 100, "", "test", "uaid", "daaa")
	assert.Equal(t, err, nil, "they should be equal")
	size := len(r)
	assert.Equal(t, size, 100, "they should be equal")

	r_1, _, err_1 := model.GetUaInfos(xl, 0, "", "test", UA_ITEM_UAID, "daaa099")
	assert.Equal(t, err_1, nil, "they should be equal")
	size_1 := len(r_1)
	assert.Equal(t, size_1, 1, "they should be equal")
	assert.Equal(t, r_1[0].Namespace, "test", "they should be equal")
	for count := 0; count < 100; count++ {
		model.Delete(xl, "test", fmt.Sprintf("daaa%03d", count))
	}
}

func TestWrongPriUrl(t *testing.T) {
	xl := xlog.NewDummy()
	xl.Infof("TestWrongPriUrl")
	url := "mongodb://root:public@180.97.147.164:27017,180.97.147.179:27017/admin"
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
	db.DinitDb()
	xl.Infof("DB init\n")
	db.InitDb(&config)
	assert.Equal(t, 0, 0, "they should be equal")
	xl.Infof("Test sleep 60s, please use rs.stepDown(20) to switch secondard by manual\n")
	time.Sleep(time.Duration(1) * time.Second)
	model := UaModel{}
	// Add ua, count size 10, from 0 to 10.
	for count := 0; count < 100; count++ {
		p := UaInfo{
			Uid:       "UserTest",
			UaId:      fmt.Sprintf("daaa%03d", count),
			Namespace: "test",
			Password:  "112sss",
		}
		err := model.Register(xl, p)
		assert.Equal(t, err, nil, "they should be equal")
	}

	// Get ua.
	r, _, err := model.GetUaInfos(xl, 0, "", "test", "uaid", "daaa")
	assert.Equal(t, err, nil, "they should be equal")
	size := len(r)
	assert.Equal(t, size, 100, "they should be equal")
	for count := 0; count < 100; count++ {
		model.Delete(xl, "test", fmt.Sprintf("daaa%03d", count))
	}
}
