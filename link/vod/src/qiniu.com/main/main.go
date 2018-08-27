package main

import (
        "fmt"
        "os"

        "github.com/gin-gonic/gin"
        "qiniu.com/controllers"
        "qiniu.com/db"
        "qiniu.com/models"
        "qiniu.com/system"
)

func main() {

        r := gin.Default()
        Config, err := system.LoadConf("conf.js")
        if err != nil {
                fmt.Println("read conf file error, error = ", err)
                os.Exit(3)
        }
        initDb()

        r.POST("/v1/namespaces/:namespace/uas", controllers.RegisterUa)
        r.DELETE("/v1/namespaces/:namespace/uas/:uaid", controllers.DeleteUa)
        r.PUT("/v1/namespaces/:namespace/uas/:uaid", controllers.UpdateUa)
        r.GET("/v1/namespaces/:namespace/uas/:uaid", controllers.GetUaInfo)
        r.GET("/v1/namespaces/:namespace/uas", controllers.GetUaInfos)

        r.POST("/v1/uids/:uid/namespaces", controllers.RegisterNamespace)
        r.DELETE("/v1/uids/:uid/namespaces/:namespace", controllers.DeleteNamespace)
        r.PUT("/v1/uids/:uid/namespaces/:namespace", controllers.UpdateNamespace)
        r.GET("/v1/uids/:uid/namespaces/:namespace", controllers.GetNamespaceInfo)
        r.GET("/v1/uids/:uid/namespaces", controllers.GetNamespaceInfos)

	r.GET("/v1/namespaces/:namespace/uas/:uaid/playback", controllers.GetPlayBackm3u8)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/segments", controllers.GetSegments)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/frames", controllers.GetFrames)
	r.POST("/qiniu/upload/callback", controllers.UploadTs)
	r.Run(Config.Bind) // listen and serve on 0.0.0.0:8080

}

func initDb() {
        url := "mongodb://root:public@180.97.147.164:27017,180.97.147.179:27017/admin"
        dbName := "vod"
        config := db.MgoConfig{
                Host:     url,
                DB:       dbName,
                Mode:     "",
                Username: "root",
                Password: "public",
                AuthDB:   "admin",
                Proxies:  nil,
        }
        if err := db.InitDb(&config); err != nil {
                fmt.Println(err)
                os.Exit(3)
        }
        segment := models.SegmentModel{}
        if err := segment.Init(); err != nil {
                fmt.Println(err)
                os.Exit(3)
        }
}
