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

	r.GET("/playback/:uid/:deviceId", controllers.GetPlayBackm3u8)
	r.GET("/segments/:uid/:deviceId", controllers.GetSegments)
	r.POST("/qiniu/upload/callback", controllers.UploadTs)
	//	r.POST("/qiniu/notify/*action", controllers.FopNotify)
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
