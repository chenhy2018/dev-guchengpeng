package db

import (
	"fmt"
	"gopkg.in/mgo.v2"
        "github.com/qiniu/db/mgoutil.v3"
        "qbox.us/lbsocketproxy"
)

type mgoDB struct {
	session   *mgo.Session
	db        string
}

type MgoConfig struct {
        Host           string                `json:"host"`
        DB             string                `json:"db"`
        Mode           string                `json:"mode"`
        Username       string                `json:"username"`
        Password       string                `json:"password"`
        AuthDB         string                `json:"authdb"`
        Proxies        *lbsocketproxy.Config `json:"proxies"`
}

var GlobConn *mgoDB

type ColConfig struct {
        A mgoutil.Collection      `coll:"segment"`
        B *mgo.Collection `coll:"device"`
}

func InitDb(config *MgoConfig) error {

        if GlobConn != nil {
                fmt.Printf("db already connected \n")
                return nil
        }
        var ret ColConfig
        cfg := mgoutil.Config {
                Host : config.Host,
                DB   : config.DB,
                Mode : config.Mode,
                Username : "",
                Password : "",
                AuthDB : "",
                Safe : nil,
                Proxies : config.Proxies,
        }
        session, err := mgoutil.Open(&ret, &cfg)
        if err != nil {
                return fmt.Errorf("db open failed: %s", err)
        }

        GlobConn = &mgoDB{
                session: session,
                db: config.DB,
        }
        return nil
}

func DinitDb() {
        if GlobConn != nil && GlobConn.session != nil {
                GlobConn.session.Close()
        }
        GlobConn = nil
}

func Session() *mgo.Session {

	// return the original session
	return GlobConn.session
}

func cloneSession() *mgo.Session {

	// return cloned session
        if GlobConn != nil && GlobConn.session != nil {
	        return GlobConn.session.Clone()
        }
        return nil
}

func WithCollection(coll string, cb func(*mgo.Collection) error) error {

	session := cloneSession()
        if session == nil {
                return nil
        }
	defer session.Close()
	c := session.DB(GlobConn.db).C(coll)
	return cb(c)
}
