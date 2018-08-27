package models

import (
        "fmt"
        "testing"
        "qiniu.com/db"
        "github.com/stretchr/testify/assert"
        "github.com/qiniu/xlog.v1"
)

func TestNamespace(t *testing.T) {
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
        xl.Infof("TestNamespace")
        db.InitDb(&config)
        assert.Equal(t, 0, 0, "they should be equal")
        model := NamespaceModel{}
        // Add ua, count size 100, from 0 to 100. 
        for count := 0; count < 100; count++ {
                p := NamespaceInfo{
                        Space : fmt.Sprintf("test%03d",count),
                        Bucket : "www.qiniu.com/112sss",
                        Uid: fmt.Sprintf("Uid%03d",count),
                }
                err := model.Register(xl, p)
                assert.Equal(t, err, nil, "they should be equal")
        }
        xl.Infof("DB Namespace Register done")
        // Get Namespace.
        r, err := model.GetNamespaceInfo(xl, "Uid099", "test099")
        assert.Equal(t, err, nil, "they should be equal")
        assert.Equal(t, r.Space, "test099", "they should be equal")

        r_1, _,err_1 := model.GetNamespaceInfos(xl, 0, "", "Uid099", "namespace", "test099")
        assert.Equal(t, err_1, nil, "they should be equal")
        size_1 := len(r_1)
        assert.Equal(t, size_1, 1, "they should be equal")
        assert.Equal(t, r_1[0].Space, "test099", "they should be equal")
        for count := 0; count < 100; count++ {
                model.Delete(xl, "Uid099", fmt.Sprintf("daaa%03d", count))
        }
}
