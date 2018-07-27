package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"qiniu.com/controllers"
	"qiniu.com/db"
	"qiniu.com/system"
)

func main() {

	r := gin.Default()
	Config, err := system.LoadConf("conf.js")
	if err != nil {
		fmt.Println("read conf file error, error = ", err)
		os.Exit(3)
	}
	initDbConnect(Config)
	defer db.Disconnect()
	r.GET("/playback/:uid/:deviceId", controllers.PlayBackGetm3u8)
	r.POST("/upload", controllers.UploadTs)
	r.Run(Config.Bind) // listen and serve on 0.0.0.0:8080

}

func initDbConnect(conf *system.Configuration) {
	//
	// connect mongoDB
	//

	// get mongoDB information
	//dbName, _ := conf.GlbVars.MongodbName.Get()
	//url, _ := conf.GlbVars.MongodbUrl.Get()
	//usr, _ := conf.GlbVars.MongodbUserTab.Get()
	//log, _ := conf.GlbVars.MongodbLogTab.Get()
	//url := "mongodb://180.97.147.164:27017:180.97.147.179:27017"
	// connect to DB
	if err := db.Connect(conf.DbConf.Host, conf.DbConf.Db, conf.DbConf.User, conf.DbConf.Password); err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}
