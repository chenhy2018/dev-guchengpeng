package models

import (
	"fmt"
	"testing"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qiniu.com/db"
	//"time"
)

func TestUa(t *testing.T) {
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
	r, _, err := model.GetUaInfos(xl, 100, "", "UserTest", "test", "daaa")
	assert.Equal(t, err, nil, "they should be equal")
	size := len(r)
	assert.Equal(t, size, 100, "they should be equal")

	r, next, err := model.GetUaInfos(xl, 2, "", "UserTest", "test", "daaa09")
	assert.Equal(t, err, nil, "they should be equal")
	size = len(r)
	assert.Equal(t, size, 2, "they should be equal")
	assert.Equal(t, r[0].UaId, "daaa090", "they should be equal")
	assert.Equal(t, next, "UserTest.daaa091.", "they should be equal")

	r, _, err = model.GetUaInfos(xl, 100, next, "UserTest", "test", "daaa09")
	assert.Equal(t, err, nil, "they should be equal")
	size = len(r)
	assert.Equal(t, size, 8, "they should be equal")
	assert.Equal(t, r[0].UaId, "daaa092", "they should be equal")

	r_1, _, err_1 := model.GetUaInfos(xl, 0, "", "UserTest", "test", "daaa099")
	assert.Equal(t, err_1, nil, "they should be equal")
	size_1 := len(r_1)
	assert.Equal(t, size_1, 1, "they should be equal")
	assert.Equal(t, r_1[0].Namespace, "test", "they should be equal")
	assert.Equal(t, r_1[0].Vod, false, "they should be equal")
	assert.Equal(t, r_1[0].Live, false, "they should be equal")
	assert.Equal(t, r_1[0].Online, false, "they should be equal")
	assert.Equal(t, r_1[0].Expire, 0, "they should be equal")

	cond := map[string]interface{}{
		UA_ITEM_VOD: true,
	}
	model.UpdateFunction(xl, "UserTest", "daaa099", UA_ITEM_VOD, cond)

	cond = map[string]interface{}{
		UA_ITEM_LIVE: true,
	}
	model.UpdateFunction(xl, "UserTest", "daaa099", UA_ITEM_LIVE, cond)

	cond = map[string]interface{}{
		UA_ITEM_ONLINE: true,
	}
	model.UpdateFunction(xl, "UserTest", "daaa099", UA_ITEM_ONLINE, cond)

	cond = map[string]interface{}{
		UA_ITEM_EXPIRE: 7,
	}
	model.UpdateFunction(xl, "UserTest", "daaa099", UA_ITEM_EXPIRE, cond)

	r_1, _, err_1 = model.GetUaInfos(xl, 0, "", "UserTest", "test", "daaa099")

	assert.Equal(t, r_1[0].Vod, true, "they should be equal")
	assert.Equal(t, r_1[0].Live, true, "they should be equal")
	assert.Equal(t, r_1[0].Online, true, "they should be equal")
	assert.Equal(t, r_1[0].Expire, 7, "they should be equal")

	for count := 0; count < 100; count++ {
		cond := map[string]interface{}{
			UA_ITEM_UID:  "UserTest",
			UA_ITEM_UAID: fmt.Sprintf("daaa%03d", count),
		}
		model.Delete(xl, cond)
	}
}

func TestWrongPriUrl(t *testing.T) {
	xl := xlog.NewDummy()
	xl.Infof("TestWrongPriUrl")
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
	db.DinitDb()
	xl.Infof("DB init\n")
	db.InitDb(&config)
	assert.Equal(t, 0, 0, "they should be equal")
	// xl.Infof("Test sleep 60s, please use rs.stepDown(20) to switch secondard by manual\n")
	// time.Sleep(time.Duration(1) * time.Second)
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
	r, _, err := model.GetUaInfos(xl, 0, "", "UserTest", "test", "daaa")
	assert.Equal(t, err, nil, "they should be equal")
	size := len(r)
	assert.Equal(t, size, 100, "they should be equal")
	for count := 0; count < 100; count++ {
		cond := map[string]interface{}{
			UA_ITEM_UID:  "UserTest",
			UA_ITEM_UAID: fmt.Sprintf("daaa%03d", count),
		}
		model.Delete(xl, cond)
	}
}
