package main

import (
        "github.com/gin-gonic/gin"
        "qiniu.com/db"
        "fmt"
        "os"
        
)

func main() {

	r := gin.Default()

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
        if err := db.Connect(url, dbName); err != nil {
                fmt.Println(err)
                os.Exit(3)
        }
        defer db.Disconnect()


	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})

	})

	r.Run() // listen and serve on 0.0.0.0:8080

}
