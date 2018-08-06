package sharding

import (
	"time"

	"github.com/qiniu/log.v1"
	"labix.org/v2/mgo"
	"qbox.us/mgo2"
)

// ------------------------------------------------------------------------

type Config struct {
	Hosts          []string `json:"hosts"`
	DB             string   `json:"db"`
	Coll           string   `json:"coll"`
	Mode           string   `json:"mode"`
	Direct         int32    `json:"direct"`  // 如果非0，表示 Hosts 是完整列表
	SyncTimeoutInS int32    `json:"timeout"` // 以秒为单位
}

type Session struct {
	*mgo.Session
	DB   *mgo.Database
	Coll *mgo.Collection
}

func Open(cfg *Config) *Session {

	info := &mgo.DialInfo{
		Addrs:   cfg.Hosts,
		Direct:  cfg.Direct != 0,
		Timeout: time.Duration(int64(cfg.SyncTimeoutInS) * 1e9),
	}

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Panic("Connect MongoDB failed:", err, cfg.Hosts)
	}

	if cfg.Mode != "" {
		mgo2.SetMode(session, cfg.Mode, true)
	}

	db := session.DB(cfg.DB)
	c := db.C(cfg.Coll)

	return &Session{session, db, c}
}

// ------------------------------------------------------------------------
