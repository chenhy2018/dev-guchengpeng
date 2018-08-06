package models

import (
        "fmt"
        "testing"
        "qiniu.com/db"
        "github.com/stretchr/testify/assert"
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
        fmt.Printf("db test user \n")
        db.InitDb(&config)
        info := UserInfo {
                Uid : "test",
                Password : "test",
                RegTime : time.Now().Unix(),
        }
        err := AddUser(info, "admin", "linking")
        assert.Equal(t, err, nil, "they should be equal")
        info = UserInfo {
                Uid : "test1",
                Password : "test1",
                RegTime : time.Now().Unix(),
        }
        err = AddUser(info, "admin", "linking")
        assert.Equal(t, err, nil, "they should be equal")

        info = UserInfo {
                Uid : "test2",
                Password : "test2",
                RegTime : time.Now().Unix(),
        }
        err = AddUser(info, "admin", "linking")
        assert.Equal(t, err, nil, "they should be equal")

        info = UserInfo {
                Uid : "test3",
                Password : "test3",
                RegTime : time.Now().Unix(),
        }
        AddUser(info, "admin", "linking")
        assert.Equal(t, err, nil, "they should be equal")

        r, err := GetUserInfo(0, 0, "admin", "linking", "uid", "test")
        assert.Equal(t, err, nil, "they should be equal")
        size := len(r)
        assert.Equal(t, size, 4, "they should be equal")

        ValidateLogin("test", "test")
        ValidateUid("test1")
        ResetPassword("test", "test", "test1")
        ResetPassword("test", "test1", "test")
        Logout("test")
        info = UserInfo {
                Uid : "test",
                Password : "test",
                RegTime : time.Now().Unix(),
        }
        DelUser(info, "admin", "linking")
        info = UserInfo {
                Uid : "test1",
                Password : "test1",
                RegTime : time.Now().Unix(),
        }       
        DelUser(info, "admin", "linking")
        info = UserInfo {
                Uid : "test2",
                Password : "test2",
                RegTime : time.Now().Unix(),
        }       
        DelUser(info, "admin", "linking")
        info = UserInfo {
                Uid : "test3",
                Password : "test3",
                RegTime : time.Now().Unix(),
        }       
        DelUser(info, "admin", "linking")
}
