package main

import (
	"runtime"

	"github.com/qiniu/log.v1"
	"qbox.us/auditlog2"
	"qbox.us/cc/config"
	"qbox.us/http/account.v2.1/digest_auth"
	"qbox.us/qmq/v1/mq"
	"qiniupkg.com/trace.v1"

	_ "github.com/qiniu/version"
	_ "qbox.us/autoua"
	_ "qbox.us/profile"
)

func init() {
	trace.TracerEnable(trace.SetService("MQ"))
}

type Conf struct {
	MaxProcs          int                `json:"max_procs"`
	DebugLevel        int                `json:"debug_level"`
	BindHost          string             `json:"bind_host"`
	DataPath          string             `json:"data_path"`
	ChunkBits         uint               `json:"chunkbits"`
	Expires           uint               `json:"expires"` // // 以秒为单位
	AuthConf          digest_auth.Config `json:"auth"`
	AuditlogDir       string             `json:"audit_log_dir"`
	AuditlogChunkbits int                `json:"audit_log_chunkbits"`
	SaveHours         int                `json:"save_hours_after_consume"`
	CheckInterval     int64              `json:"check_interval_sec"`
}

func main() {

	config.Init("f", "qbox", "qboxmq.conf")

	var conf Conf
	err := config.Load(&conf)
	if err != nil {
		return
	}

	log.Infof("conf: %#v", conf)
	log.SetOutputLevel(conf.DebugLevel)
	runtime.GOMAXPROCS(conf.MaxProcs)

	acc := digest_auth.New(&conf.AuthConf)

	mqCfg := &mq.Config{
		Account:  acc,
		DataPath: conf.DataPath,
		Config: auditlog2.Config{
			LogFile:   conf.AuditlogDir,
			ChunkBits: byte(conf.AuditlogChunkbits),
		},
		ChunkBits:     conf.ChunkBits,
		Expires:       conf.Expires,
		SaveHours:     conf.SaveHours,
		CheckInterval: conf.CheckInterval,
	}

	log.Println("qboxmq running at", conf.BindHost)
	err = mq.Run(conf.BindHost, mqCfg)
	if err != nil {
		log.Warn("mq.Run failed", err)
	}
}
