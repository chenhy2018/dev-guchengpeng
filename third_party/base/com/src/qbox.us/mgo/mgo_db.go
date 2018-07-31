package mgo

import (
	"time"

	"github.com/qiniu/log.v1"
	"launchpad.net/mgo"
)

// ------------------------------------------------------------------------

func Dail(host string, mode string, syncTimeoutInS int64) *mgo.Session {

	session, err := mgo.Dial(host)
	if err != nil {
		log.Panic("Connect MongoDB failed:", err, host)
	}

	if mode != "" {
		SetMode(session, mode, true)
	}
	if syncTimeoutInS != 0 {
		session.SetSyncTimeout(time.Duration(syncTimeoutInS * 1e9))
	}

	return session
}

// ------------------------------------------------------------------------

type Config struct {
	Host           string `json:"host"`
	DB             string `json:"db"`
	Coll           string `json:"coll"`
	Mode           string `json:"mode"`
	SyncTimeoutInS int64  `json:"timeout"` // 以秒为单位
}

type Session struct {
	*mgo.Session
	DB   *mgo.Database
	Coll *mgo.Collection
}

func Open(cfg *Config) *Session {

	session := Dail(cfg.Host, cfg.Mode, cfg.SyncTimeoutInS)
	db := session.DB(cfg.DB)
	c := db.C(cfg.Coll)

	return &Session{session, db, c}
}

// ------------------------------------------------------------------------
