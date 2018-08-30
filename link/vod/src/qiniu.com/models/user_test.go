package models

import (
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qiniu.com/db"
	"testing"
)

func TestUser(t *testing.T) {
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
	xl.Infof("TestUser")
	db.InitDb(&config)
	info := UserInfo{
		Uid:      "test",
		Password: "test",
	}
	err := AddUser(xl, info, "admin", "linking")
	assert.Equal(t, err, nil, "they should be equal")
	info = UserInfo{
		Uid:      "test1",
		Password: "test1",
	}
	err = AddUser(xl, info, "admin", "linking")
	assert.Equal(t, err, nil, "they should be equal")

	info = UserInfo{
		Uid:      "test2",
		Password: "test2",
	}
	err = AddUser(xl, info, "admin", "linking")
	assert.Equal(t, err, nil, "they should be equal")

	info = UserInfo{
		Uid:      "test3",
		Password: "test3",
	}
	AddUser(xl, info, "admin", "linking")
	assert.Equal(t, err, nil, "they should be equal")

	r, err := GetUserInfo(xl, 0, 0, "admin", "linking", "uid", "test")
	assert.Equal(t, err, nil, "they should be equal")
	size := len(r)
	assert.Equal(t, size, 4, "they should be equal")

	ValidateLogin(xl, "test", "test")
	ValidateUid(xl, "test1")
	ResetPassword(xl, "test", "test", "test1")
	ResetPassword(xl, "test", "test1", "test")
	Logout(xl, "test")
	info = UserInfo{
		Uid:      "test",
		Password: "test",
	}
	DelUser(xl, info, "admin", "linking")
	info = UserInfo{
		Uid:      "test1",
		Password: "test1",
	}
	DelUser(xl, info, "admin", "linking")
	info = UserInfo{
		Uid:      "test2",
		Password: "test2",
	}
	DelUser(xl, info, "admin", "linking")
	info = UserInfo{
		Uid:      "test3",
		Password: "test3",
	}
	DelUser(xl, info, "admin", "linking")
}
