package models

import (
        "fmt"
        "qiniu.com/db"
        "gopkg.in/mgo.v2"
        "gopkg.in/mgo.v2/bson"
        "github.com/qiniu/xlog.v1"
        "time"
)

type namespaceModel struct {
}

var (
        Namespace *namespaceModel
)

func (m *namespaceModel) Init() error {
        return nil
}

func (m *namespaceModel) Register(xl *xlog.Logger, req NamespaceInfo) error {
        /*
                 db.namespace.update( {_id: req.Id}, {"$set": {"bucketurl": req.Bucketurl}},
                 { upsert: true })
        */
        err := db.WithCollection(
                NAMESPACE_COL,
                func(c *mgo.Collection) error {
                        _, err := c.Upsert(
                                bson.M{
                                        ITEM_ID:  req.Id,
                                },
                                bson.M{
                                        "$set": bson.M{
                                                ITEM_ID:  req.Id,
                                                NAMESPACE_ITEM_DATE: time.Now().Unix(),
                                                NAMESPACE_ITEM_BUCKETURL: req.BucketUrl,
                                        },
                                },
                        )
                        return err
                },
        )
        if err != nil {
                return err
        }
        return nil;
}

func (m *namespaceModel) Delete(xl *xlog.Logger, id string) error {
        /*
                 db.namespace.remove({"_id": id})
        */
        return db.WithCollection(
                NAMESPACE_COL,
                func(c *mgo.Collection) error {
			return c.Remove(
				bson.M{
					ITEM_ID: id,
				},
                        )
                },
        )
}

type NamespaceInfo struct {
        Id            string  `bson:"_id"        json:"_id"`
        Regtime       int     `bson:"date"       json:"date"`
        BucketUrl     string  `bson:"bucketurl"  json:"bucketurl"`
}

func (m *namespaceModel) GetNamespaceInfo(xl *xlog.Logger, index, rows int, category, like string) ([]NamespaceInfo, error) {

        /*
                 db.namespace.find({category: {"$regex": "*like*"}},
                 ).sort({"_id":1}).limit(rows),skip(rows * index)
        */
        // query by keywords
        query := bson.M{}
        if like != "" {
                query[category] = bson.M{
                        "$regex": ".*" + like + ".*",
                }
        }

        // direct to specific page
        skip := rows * index
        limit := rows
        if limit > 100 {
                limit = 100
        }

        // query
        r := []NamespaceInfo{}
        count := 0
        err := db.WithCollection(
                NAMESPACE_COL,
                func(c *mgo.Collection) error {
                        var err error
                        if err = c.Find(query).Sort(ITEM_ID).Skip(skip).Limit(limit).All(&r); err != nil {
                                return fmt.Errorf("query failed")
                        }
                        if count, err = c.Find(query).Count(); err != nil {
                                return fmt.Errorf("query count failed")
                        }
                        return nil
                },
        )
        if err != nil {
               return []NamespaceInfo{}, err
        }
        return r, nil
}

func (m *namespaceModel) UpdateNamespace(xl *xlog.Logger, info NamespaceInfo) error {
        /*
                 db.namespace.update({"_id": id}, bson.M{"$set":{"bucketurl": info.BucketUrl}}),
        */
         return db.WithCollection(
                NAMESPACE_COL,
                func(c *mgo.Collection) error {
                        return c.Update(
                                bson.M{
                                        ITEM_ID:  info.Id,
                                },
                                bson.M{
                                        "$set": bson.M{
                                                NAMESPACE_ITEM_BUCKETURL: info.BucketUrl,
                                        },
                                },
                        )
                },
        )
}

