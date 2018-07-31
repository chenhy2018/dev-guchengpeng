package models

import (
        "fmt"
        "testing"
        "qiniu.com/db"
        "github.com/stretchr/testify/assert"
        "time"
)

func TestDevice(t *testing.T) {
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
        fmt.Printf("db INItDB")
        db.InitDb(&config)
        assert.Equal(t, 0, 0, "they should be equal")
        model := deviceModel{}
        // Add device, count size 100, from 0 to 100. 
        for count := 0; count < 100; count++ {
       		p := RegisterReq{
        		Uuid : "UserTest",
        		Deviceid : fmt.Sprintf("daaa%d", count),
        		BucketUrl : "www.qiniu.io/test/",
        		RemainDays : int64(count),
        	}
        	err := model.Register(p)
        	assert.Equal(t, err, nil, "they should be equal")
        }

        // Get device.
        r, err := model.GetDeviceInfo(0,0,"uuid", "UserTest")
        assert.Equal(t, err, nil, "they should be equal")
        size := len(r)
        assert.Equal(t, size, 100, "they should be equal")
        
        for count := 0; count < 100; count++ {
                assert.Equal(t, r[count].Expire, int64(count), "they should be equal") 
        }
        model.UpdateRemaindays("UserTest", "daaa99", 1000000);
        r_1, err_1 := model.GetDeviceInfo(0,0,DEVICE_ITEM_DEVICEID, "daaa99")
        assert.Equal(t, err_1, nil, "they should be equal")
        size_1 := len(r_1)
        assert.Equal(t, size_1, 1, "they should be equal")
        assert.Equal(t, r_1[0].Expire, int64(1000000), "they should be equal")
        for count := 0; count < 100; count++ {
                model.Delete("UserTest", fmt.Sprintf("daaa%d", count))
        }
}

func TestWrongPriUrl(t *testing.T) {
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
        db.DinitDb()
        fmt.Printf("db INITDB\n")
        db.InitDb(&config)
        assert.Equal(t, 0, 0, "they should be equal")
        fmt.Printf("Test sleep 60s, please use rs.stepDown(20) to switch secondard by manual\n")
        time.Sleep(time.Duration(60)*time.Second)
        model := deviceModel{}
        // Add device, count size 10, from 0 to 10.
        for count := 0; count < 100; count++ {
                p := RegisterReq{
                        Uuid : "UserTest",
                        Deviceid : fmt.Sprintf("daaa%d", count),
                        BucketUrl : "www.qiniu.io/test/",
                        RemainDays : int64(count),
                }
                err := model.Register(p)
                assert.Equal(t, err, nil, "they should be equal")
        }

        // Get device.
        r, err := model.GetDeviceInfo(0,0,"uuid", "UserTest")
        assert.Equal(t, err, nil, "they should be equal")
        size := len(r)
        assert.Equal(t, size, 100, "they should be equal")
}
