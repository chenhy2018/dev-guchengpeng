package mgocon

import (
	"qbox.us/mgo2"
)

type MongoDB interface {
	String() string
}

var (
	MongoDBs = map[MongoDB]func() *mgo2.Database{}
)
