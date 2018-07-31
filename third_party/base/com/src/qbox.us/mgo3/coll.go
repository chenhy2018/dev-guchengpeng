package mgo3

import (
	"strings"

	"github.com/qiniu/log.v1"
	"labix.org/v2/mgo"
)

// ------------------------------------------------------------------------

type Collection struct {
	*mgo.Collection
}

// ensure indexes for a collection
//
// eg. c.EnsureIndexes([]string{"uid", "email"}, []string{"serial_num", "uid,status,delete"})
//
func (c Collection) EnsureIndexes(uniques []string, indexes []string) {

	for _, colIndex := range uniques {
		colIndexArr := strings.Split(colIndex, ",")
		err := c.EnsureIndex(mgo.Index{Key: colIndexArr, Unique: true})
		if err != nil {
			log.Panic("<Mongo.C>:", c.Name, "Index:", colIndexArr, " error:", err)
			return
		}
	}

	for _, colIndex := range indexes {
		colIndexArr := strings.Split(colIndex, ",")
		err := c.EnsureIndex(mgo.Index{Key: colIndexArr})
		if err != nil {
			log.Panic("<Mongo.C>:", c.Name, "Index:", colIndexArr, " error:", err)
			return
		}
	}
}

func (c Collection) Copy() Collection {

	db := c.Database
	return Collection{db.Session.Copy().DB(db.Name).C(c.Name)}
}

func (c Collection) Close() (err error) {

	c.Database.Session.Close()
	return nil
}

// ------------------------------------------------------------------------
