package models

import (
        "fmt"
        "qiniu.com/db"
        "gopkg.in/mgo.v2"
        "gopkg.in/mgo.v2/bson"
        "github.com/qiniu/xlog.v1"
        "time"
        "strconv"
)

type NamespaceModel struct {
}

var (
        Namespace *NamespaceModel
)

func (m *NamespaceModel) Init() error {
        return nil
}

func (m *NamespaceModel) Register(xl *xlog.Logger, req NamespaceInfo) error {
        /*
                 db.namespace.update( {"uid":req.Uid,  "namespace": req.Space}, {"$set": {"bucketurl": req.Bucketurl}},
                 { upsert: true })
        */
        err := db.WithCollection(
                NAMESPACE_COL,
                func(c *mgo.Collection) error {
                        _, err := c.Upsert(
                                bson.M{
                                        NAMESPACE_ITEM_ID:  req.Space,
                                        NAMESPACE_ITEM_UID : req.Uid,
                                },
                                bson.M{
                                        "$set": bson.M{
                                                NAMESPACE_ITEM_ID:  req.Space,
                                                ITEM_CREATE_TIME: time.Now().Unix(),
                                                NAMESPACE_ITEM_BUCKET: req.Bucket,
                                                ITEM_UPDATA_TIME : time.Now().Unix(),
                                                NAMESPACE_ITEM_UID : req.Uid,
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

func (m *NamespaceModel) Delete(xl *xlog.Logger, uid,id string) error {
        /*
                 db.namespace.remove({"uid": uid,  "namespace": id})
        */
        return db.WithCollection(
                NAMESPACE_COL,
                func(c *mgo.Collection) error {
			return c.Remove(
				bson.M{
					NAMESPACE_ITEM_ID: id,
                                        NAMESPACE_ITEM_UID : uid,
				},
                        )
                },
        )
}

type NamespaceInfo struct {
        Space         string  `bson:"namespace"  json:"namespace"`
        Regtime       int64   `bson:"createdAt"  json:"createdAt"`
        UpdateTime    int64   `bson:"updatedAt"  json:"updatedAt"`  
        Bucket        string  `bson:"bucket"     json:"bucket"`
        Uid           string  `bson:"uid"        json:"uid"`
}

func (m *NamespaceModel) GetNamespaceInfo(xl *xlog.Logger,uid, namespace string) (NamespaceInfo, error) {
        /*
                 db.namespace.find({"uid":uid, "namespace": namespace})
        */
        r := NamespaceInfo{}
        err := db.WithCollection(
                NAMESPACE_COL,
                func(c *mgo.Collection) error {
                        return c.Find(
                                bson.M{
                                        NAMESPACE_ITEM_ID:  namespace,
                                        NAMESPACE_ITEM_UID : uid,
                                },
                        ).One(&r)
                },
        )
        return r, err 
}

func (m *NamespaceModel) GetNamespaceInfos(xl *xlog.Logger, limit int, mark, uid, category, like string) ([]NamespaceInfo, string, error) {

        /*
                 db.namespace.find({"uid" : uid, { category: {"$regex": "*like*"}}},
                 ).sort({"namespace":1}).limit(limit),skip(mark)
        */
        // query by keywords
        query := bson.M{
                NAMESPACE_ITEM_UID : uid,
                category : bson.M{ "$regex": ".*" + like + ".*", },
        }
        nextMark := ""
        // direct to specific page
        skip , err := strconv.ParseInt(mark, 10, 32)
        if err != nil {
                skip = 0
        }

        if limit == 0 {
                limit = 65535
        }

        // query
        r := []NamespaceInfo{}
        count := 0
        err = db.WithCollection(
                NAMESPACE_COL,
                func(c *mgo.Collection) error {
                        var err error
                        if err = c.Find(query).Sort(NAMESPACE_ITEM_ID).Skip(int(skip)).Limit(limit).All(&r); err != nil {
                                return fmt.Errorf("query failed")
                        }
                        if count, err = c.Find(query).Count(); err != nil {
                                return fmt.Errorf("query count failed")
                        }
                        return nil
                },
        )
        if err != nil {
               return []NamespaceInfo{}, "", err
        }
        if (count == limit) {
               nextMark = fmt.Sprintf("%d", count)
        }
        return r, nextMark, nil
}

func (m *NamespaceModel) UpdateNamespace(xl *xlog.Logger, uid, space string, info NamespaceInfo) error {
        /*
                 db.namespace.update({"uid": uid, "namespace": space}, bson.M{"$set":{"bucketurl": info.BucketUrl}}),
        */
         return db.WithCollection(
                NAMESPACE_COL,
                func(c *mgo.Collection) error {
                        return c.Update(
                                bson.M{
                                        NAMESPACE_ITEM_ID:  space,
                                        NAMESPACE_ITEM_UID : uid,
                                },
                                bson.M{
                                        "$set": bson.M{
                                                NAMESPACE_ITEM_ID:  info.Space,
                                                NAMESPACE_ITEM_BUCKET: info.Bucket,
                                                ITEM_UPDATA_TIME : time.Now().Unix(),
                                                NAMESPACE_ITEM_UID : info.Uid,
                                        },
                                },
                        )
                },
        )
}
