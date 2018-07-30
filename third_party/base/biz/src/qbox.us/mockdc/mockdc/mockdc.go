package main

import (
	"os"
	"qbox.us/cc/config"
	"github.com/qiniu/log.v1"
	"qbox.us/mockdc"
	"runtime"
)

type Conf struct {
	BindHost   string `json:"bind_host"`
	DataPath   string `json:"data_path"`
	Key        string `json:"key"`
	MaxProcs   int    `json:"max_procs"`
	DebugLevel int    `json:"debug_level"`
}

func main() {

	config.Init("f", "qbox", "mocklbd.conf")

	var conf Conf
	err := config.Load(&conf)
	if err != nil {
		return
	}

	err = os.MkdirAll(conf.DataPath, 0777)
	if err != nil {
		log.Println("Mkdir failed:", err)
		return
	}

	log.SetOutputLevel(conf.DebugLevel)

	runtime.GOMAXPROCS(conf.MaxProcs)

	cfg := &mockdc.Config{
		Key: []byte(conf.Key),
	}
	p, err := mockdc.New(cfg)
	if err != nil {
		log.Error(err)
		return
	}
	err = p.Run(conf.BindHost, conf.DataPath)
	log.Fatal(err)
}
