package main

import (
	"fmt"
	"os"
	"path/filepath"

	"controllers"
	"system"

	"github.com/gin-gonic/gin"
	"qiniu.com/db"
)

func main() {

	r := gin.Default()
	Config, err := system.LoadConf("conf.js")
	if err != nil {
		fmt.Println("read conf file error, error = ", err)
		os.Exit(3)
	}
	initDb()
	setTemplate(r)

	r.Static("/static", "src/static/")
	r.GET("/playback/:uid/:deviceid.m3u8", controllers.AddTokenAndRedirect)
	r.GET("/segments/:uid/:deviceid", controllers.AddTokenAndRedirect)

	r.GET("/qiniu/upload/token", controllers.GetUnloadToken)

	r.GET("/index", controllers.Index)
	r.GET("/login", controllers.Login)
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
}

func setTemplate(engine *gin.Engine) {
	engine.LoadHTMLGlob(filepath.Join("src/views/*"))
}
