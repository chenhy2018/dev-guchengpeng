package db

import (
	"fmt"
	"gopkg.in/mgo.v2"
)

type mgoDB struct {
	session   *mgo.Session
	db        string
}

var GlobConn *mgoDB

func Connect(url, db string) error {

	if GlobConn != nil {
		return fmt.Errorf("db already connected")
	}

	session, err := mgo.Dial(url)
	if err != nil {
		return fmt.Errorf("db not connected: %s", err)
	}
	GlobConn = &mgoDB{
		session: session,
		db: db,
	}
	return nil
}

func Disconnect() {

	if GlobConn != nil && GlobConn.session != nil {
		GlobConn.session.Close()
	}
}

func Session() *mgo.Session {

	// return the original session
	return GlobConn.session
}

func cloneSession() *mgo.Session {

	// return cloned session
	return GlobConn.session.Clone()
}

func WithCollection(coll string, cb func(*mgo.Collection) error) error {

	session := cloneSession()
	defer session.Close()
	c := session.DB(GlobConn.db).C(coll)
	return cb(c)
}
