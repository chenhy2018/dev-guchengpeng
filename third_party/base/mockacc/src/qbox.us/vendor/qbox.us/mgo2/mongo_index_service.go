package mgo2

import (
	"strings"
	"sync"
)

import (
	"github.com/qiniu/log.v1"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// help maintain the index of the collections
type MongoIndexService struct {
	m        sync.Mutex
	session  *mgo.Session
	dbName   string
	colNames map[string]bool
}

func NewMongoIndexService(session *mgo.Session, dbName string) *MongoIndexService {
	return &MongoIndexService{
		m:        sync.Mutex{},
		session:  session,
		dbName:   dbName,
		colNames: map[string]bool{},
	}
}

// ensure index for a collection
//
// an example for col param:
//
//    bson.M{
//      "name": "developers",
//      "index": []string{
//        "serial_num",
//        "uid,status,delete",
//      },
//      "unique": []string{
//        "uid",
//        "email",
//      },
//    }

func (s *MongoIndexService) EnsureIndex(col bson.M) {
	s.m.Lock()
	defer s.m.Unlock()

	colName := col["name"].(string)
	if _, ok := s.colNames[colName]; ok {
		return
	}

	session := s.session.Copy()
	defer session.Close()

	collection := session.DB(s.dbName).C(colName)

	if _, okIn := col["index"]; okIn {
		if colIndexs, okType := col["index"].([]string); okType {
			for _, colIndex := range colIndexs {
				colIndexArr := strings.Split(colIndex, ",")
				err := collection.EnsureIndex(mgo.Index{Key: colIndexArr, Unique: false})
				if err != nil {
					log.Panic("<Mongo.C> ", "Index:", colName, " error:", err)
					return
				}
			}
		}
	}

	if _, okIn := col["unique"]; okIn {
		if colIndexs, okType := col["unique"].([]string); okType {
			for _, colIndex := range colIndexs {
				colIndexArr := strings.Split(colIndex, ",")
				err := collection.EnsureIndex(mgo.Index{Key: colIndexArr, Unique: true})
				if err != nil {
					log.Panic("<Mongo.C> ", "Unqiue:", colName+" error:", err)
					return
				}
			}
		}
	}
}
