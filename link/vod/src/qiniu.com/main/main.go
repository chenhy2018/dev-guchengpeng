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
	initDbConnect()
	defer db.Disconnect()

	r.GET("/playback/:uid/:deviceId", controllers.PlayBackGetm3u8)
	r.POST("/upload", controllers.UploadTs)

	r.Run() // listen and serve on 0.0.0.0:8080

}

func initDbConnect() {
	//
	// connect mongoDB
	//

	// get mongoDB information
	//dbName, _ := conf.GlbVars.MongodbName.Get()
	//url, _ := conf.GlbVars.MongodbUrl.Get()
	//usr, _ := conf.GlbVars.MongodbUserTab.Get()
	//log, _ := conf.GlbVars.MongodbLogTab.Get()
	dbName := "vod"
	url := "180.97.147.164:27017"
	//url := "mongodb://180.97.147.164:27017:180.97.147.179:27017"
	// connect to DB
	if err := db.Connect(url, dbName, "root", "public"); err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}
