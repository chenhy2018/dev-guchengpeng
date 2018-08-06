package mgo3

import (
	"strings"
	"time"

	"qbox.us/lbsocketproxy"

	"github.com/qiniu/log.v1"
	"labix.org/v2/mgo"
)

// ------------------------------------------------------------------------

func Dail(host string, mode string, syncTimeoutInS int64) *mgo.Session {
	return DialWithProxy(host, mode, syncTimeoutInS, nil)
}

// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
func getMongoHosts(raw string) (hosts []string, user, password, authDB string) {
	if strings.HasPrefix(raw, "mongodb://") {
		raw = raw[len("mongodb://"):]
	}
	if idx := strings.Index(raw, "@"); idx != -1 {
		sp := strings.SplitN(raw[:idx], ":", 2)
		user, password = sp[0], sp[1]
		raw = raw[idx+1:]
	}
	if idx := strings.Index(raw, "/"); idx != -1 {
		authDB = raw[idx+1:]
		if idx := strings.Index(authDB, "?"); idx != -1 {
			authDB = authDB[:idx]
		}
		raw = raw[:idx]
	}
	if idx := strings.Index(raw, "?"); idx != -1 {
		raw = raw[:idx]
	}
	return strings.Split(raw, ","), user, password, authDB
}

func DialWithProxy(host, mode string, syncTimeoutInS int64, proxyConf *lbsocketproxy.Config) *mgo.Session {
	return DialWithProxyAuth(host, mode, syncTimeoutInS, proxyConf, "", "", "")
}

func checkConflict(a, b string) (final string) {
	if a != "" && b != "" {
		log.Panic("conflict", a, b)
	}
	if a != "" {
		final = a
	}
	final = b
	return
}

func DialWithProxyAuth(host, mode string, syncTimeoutInS int64, proxyConf *lbsocketproxy.Config, username, password, authDB string) *mgo.Session {
	addrs, userURL, passwordURL, authDBURL := getMongoHosts(host)
	username = checkConflict(username, userURL)
	password = checkConflict(password, passwordURL)
	authDB = checkConflict(authDB, authDBURL)
	timeout := time.Second * 10
	info := mgo.DialInfo{
		Addrs:    addrs,
		Timeout:  timeout,
		Username: username,
		Password: password,
		Database: authDB,
	}
	if proxyConf != nil {
		proxy, err := lbsocketproxy.NewLbSocketProxy(proxyConf)
		if err != nil {
			log.Panic("lbsocketproxy.NewLbSocketProxy failed", err, proxyConf)
		}
		info.Dial = proxy.Dial
	}
	session, err := mgo.DialWithInfo(&info)
	if err != nil {
		log.Panic("Connect to mongo failed", err)
	}
	session.SetSyncTimeout(1 * time.Minute)
	session.SetSocketTimeout(1 * time.Minute)

	if mode != "" {
		SetMode(session, mode, true)
	}
	if syncTimeoutInS != 0 {
		session.SetSyncTimeout(time.Duration(int64(time.Second) * syncTimeoutInS))
	}
	return session
}

// ------------------------------------------------------------------------

type Safe struct {
	W        int    `json:"w"`
	WMode    string `json:"wmode"`
	WTimeout int    `json:"wtimeoutms"`
	FSync    bool   `json:"fsync"`
	J        bool   `json:"j"`
}

type Config struct {
	Host           string                `json:"host"`
	DB             string                `json:"db"`
	Coll           string                `json:"coll"`
	Mode           string                `json:"mode"`
	Username       string                `json:"username"`
	Password       string                `json:"password"`
	AuthDB         string                `json:"authdb"`
	SyncTimeoutInS int64                 `json:"timeout"` // 以秒为单位
	Safe           *Safe                 `json:"safe"`
	Proxies        *lbsocketproxy.Config `json:"proxies"`
}

type Session struct {
	*mgo.Session
	DB   *mgo.Database
	Coll Collection
}

func Open(cfg *Config) *Session {

	session := DialWithProxyAuth(cfg.Host, cfg.Mode, cfg.SyncTimeoutInS, cfg.Proxies, cfg.Username, cfg.Password, cfg.AuthDB)
	SetSafe(session, cfg.Safe)
	db := session.DB(cfg.DB)
	c := db.C(cfg.Coll)

	return &Session{session, db, Collection{c}}
}

// ------------------------------------------------------------------------

// W 和 WMode 只在 replset 模式下生效，非replset不能配置，否则会出错
// WMode只在2.0版本以上才生效
func SetSafe(session *mgo.Session, safe *Safe) {
	if safe == nil {
		return
	}
	session.SetSafe(&mgo.Safe{
		W:        safe.W,
		WMode:    safe.WMode,
		WTimeout: safe.WTimeout,
		FSync:    safe.FSync,
		J:        safe.J,
	})
}
