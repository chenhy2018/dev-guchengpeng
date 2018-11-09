package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"qiniu.com/auth"
	"qiniu.com/controllers"
	"qiniu.com/db"
	"qiniu.com/system"
	log "qiniupkg.com/x/log.v7"
)

var ak = flag.String("access_key", "", "help message for flagname")
var sk = flag.String("secret_key", "", "help message for flagname")

func main() {

	r := gin.Default()
	conf, err := system.LoadConf("qbox", "linking_vod.conf")
	if err != nil {
		log.Error("Load conf fail", err)
		return
	}
	flag.Parse()
	fmt.Println(*ak, *sk)
	controllers.SetUserInfo(*ak, *sk)
	initDb(conf)
	if system.HaveQconf() == true {
		auth.Init(conf)
	}
	controllers.Init(&conf.GrpcConf)
	r.Use(controllers.HandleToken)
	r.POST("/v1/namespaces/:namespace/uas/:uaid", controllers.RegisterUa)
	r.DELETE("/v1/namespaces/:namespace/uas/:uaid", controllers.DeleteUa)
	r.PUT("/v1/namespaces/:namespace/uas/:uaid", controllers.UpdateUa)
	r.GET("/v1/namespaces/:namespace/uas", controllers.GetUaInfo)

	r.POST("/v1/namespaces/:namespace", controllers.RegisterNamespace)
	r.DELETE("/v1/namespaces/:namespace", controllers.DeleteNamespace)
	r.PUT("/v1/namespaces/:namespace", controllers.UpdateNamespace)
	r.GET("/v1/namespaces", controllers.GetNamespaceInfo)

	r.GET("/v1/namespaces/:namespace/uas/:uaid/playback", controllers.GetPlayBackm3u8)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/segments", controllers.GetSegments)
	r.GET("/v1/namespaces/:namespace/uas/:uaid/frames", controllers.GetFrames)
	r.POST("/qiniu/upload/callback", controllers.UploadTs)

	//Priavte api
	r.POST("/v1/aksk", controllers.SetPrivateAkSk)
	r.Run(conf.Bind) // listen and serve on 0.0.0.0:8080

}

func initDb(conf *system.Configuration) {
	if system.HaveDb() == false {
		return
	}
	//url := "mongodb://root:public@180.97.147.164:27017,180.97.147.179:27017/admin"
	url := conf.DbConf.Host
	config := db.MgoConfig{
		Host:     url,
		DB:       conf.DbConf.Db,
		Mode:     conf.DbConf.Mode,
		Username: conf.DbConf.User,
		Password: conf.DbConf.Password,
		AuthDB:   "admin",
		Proxies:  nil,
	}
	if err := db.InitDb(&config); err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}
