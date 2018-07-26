package models

import (
        "os"
        "fmt"
        "testing"
        "qiniu.com/db"
        "github.com/stretchr/testify/assert"
)

func TestDevice(t *testing.T) {
        url := "39.107.247.14:27017"
        dbName := "vod"
        if err := db.Connect(url, dbName); err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
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
        assert.Equal(t, r_1[0].Expire, 1000000, "they should be equal")
        for count := 0; count < 100; count++ {
                model.Delete("UserTest", fmt.Sprintf("daaa%d", count))
        }
        db.Disconnect()
}
