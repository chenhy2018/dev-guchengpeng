package models

import (
        "testing"
        "qiniu.com/db"
        "github.com/stretchr/testify/assert"
        "github.com/qiniu/xlog.v1"
        "time"
)

func TestUser(t *testing.T) {
        url := "mongodb://root:public@180.97.147.164:27017,180.97.147.179:27017/admin"
        dbName := "vod"
        config := db.MgoConfig {
                Host : url,
                DB   : dbName,
                Mode : "strong",
                Username : "root",
                Password : "public",
                AuthDB : "admin",
                Proxies : nil,
        }
        xl := xlog.NewDummy()
        xl.Infof("TestUser")
        db.InitDb(&config)
        info := UserInfo {
                Uid : "test",
                Password : "test",
                RegTime : time.Now().Unix(),
        }
        err := AddUser(xl, info, "admin", "linking")
        assert.Equal(t, err, nil, "they should be equal")
        info = UserInfo {
                Uid : "test1",
                Password : "test1",
                RegTime : time.Now().Unix(),
        }
        err = AddUser(xl, info, "admin", "linking")
        assert.Equal(t, err, nil, "they should be equal")

        info = UserInfo {
                Uid : "test2",
                Password : "test2",
                RegTime : time.Now().Unix(),
        }
        err = AddUser(xl, info, "admin", "linking")
        assert.Equal(t, err, nil, "they should be equal")

        info = UserInfo {
                Uid : "test3",
                Password : "test3",
                RegTime : time.Now().Unix(),
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
        info = UserInfo {
                Uid : "test",
                Password : "test",
                RegTime : time.Now().Unix(),
        }
        DelUser(xl, info, "admin", "linking")
        info = UserInfo {
                Uid : "test1",
                Password : "test1",
                RegTime : time.Now().Unix(),
        }       
        DelUser(xl, info, "admin", "linking")
        info = UserInfo {
                Uid : "test2",
                Password : "test2",
                RegTime : time.Now().Unix(),
        }       
        DelUser(xl, info, "admin", "linking")
        info = UserInfo {
                Uid : "test3",
                Password : "test3",
                RegTime : time.Now().Unix(),
        }       
        DelUser(xl, info, "admin", "linking")
}
