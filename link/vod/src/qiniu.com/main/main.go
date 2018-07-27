package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"qiniu.com/controllers"
	"qiniu.com/db"
)

func main() {

	r := gin.Default()
        initDb()
	r.GET("/playback/:uid/:deviceId", controllers.PlayBackGetm3u8)
	r.POST("/upload", controllers.UploadTs)

	r.Run() // listen and serve on 0.0.0.0:8080

}

func initDb() {
        url := "mongodb://root:public@180.97.147.164:27017,180.97.147.179:27017/admin"
        dbName := "vod"
        config := db.MgoConfig {
                Host : url,
                DB   : dbName,
                Mode : "",
                Username : "root",
                Password : "public",
                AuthDB : "admin",
                Proxies : nil,
        }
        if err := db.InitDb(&config); err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

}
