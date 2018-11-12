package models

import (
	"encoding/base64"
	"fmt"
	"github.com/qiniu/xlog.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"qiniu.com/db"
	"time"
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
					ITEM_ID: req.Uid + "." + req.Space,
				},
				bson.M{
					"$set": bson.M{
						ITEM_ID:                       req.Uid + "." + req.Space,
						NAMESPACE_ITEM_ID:             req.Space,
						ITEM_CREATE_TIME:              time.Now().Unix(),
						NAMESPACE_ITEM_BUCKET:         req.Bucket,
						ITEM_UPDATA_TIME:              time.Now().Unix(),
						NAMESPACE_ITEM_UID:            req.Uid,
						NAMESPACE_ITEM_AUTO_CREATE_UA: req.AutoCreateUa,
						NAMESPACE_ITEM_EXPIRE:         req.Expire,
					},
				},
			)
			return err
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *NamespaceModel) Delete(xl *xlog.Logger, uid, id string) error {
	/*
	   db.namespace.remove({"_id": uid + "." + id})
	*/
	return db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			return c.Remove(
				bson.M{
					ITEM_ID: uid + "." + id,
				},
			)
		},
	)
}

type NamespaceInfo struct {
	id           string `bson:"_id"  json:"_id"`
	Space        string `bson:"namespace"  json:"-"`
	Regtime      int64  `bson:"createdAt"  json:"createdAt"`
	UpdateTime   int64  `bson:"updatedAt"  json:"updatedAt"`
	Bucket       string `bson:"bucket"     json:"bucket"`
	Uid          string `bson:"uid"        json:"-"`
	AutoCreateUa bool   `bson:"auto"       json:"auto"`
	Expire       int    `bson:"expire"     json:"expire"`
}

func (m *NamespaceModel) GetNamespaceInfo(xl *xlog.Logger, uid, namespace string) ([]NamespaceInfo, error) {
	/*
	   db.namespace.find({"uid":uid, "namespace": namespace})
	*/
	r := []NamespaceInfo{}
	err := db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			return c.Find(
				bson.M{
					ITEM_ID: uid + "." + namespace,
				},
			).All(&r)
		},
	)
	return r, err
}

func (m *NamespaceModel) GetNamespaceByBucket(xl *xlog.Logger, bucket string) ([]NamespaceInfo, error) {
	/*
	   db.namespace.find({"bucket": bucket})
	*/
	r := []NamespaceInfo{}
	err := db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			return c.Find(
				bson.M{
					NAMESPACE_ITEM_BUCKET: bucket,
				},
			).All(&r)
		},
	)
	return r, err
}

func (m *NamespaceModel) GetNamespaceInfos(xl *xlog.Logger, limit int, mark, uid, prefix string) ([]NamespaceInfo, string, error) {

	/*
		db.namespace.find({"_id": "$gte": newPrefix, "$lte": uid + "/"},
		   ).sort({"namespace":1}).limit(limit),skip(mark)
	*/

	newPrefix := uid + "." + prefix
	if mark != "" {
		newMark, err := base64.URLEncoding.DecodeString(mark)
		if err == nil {
			newPrefix = uid + "." + string(newMark)
		}
	}
	// query by keywords
	query := bson.M{
		ITEM_ID: bson.M{"$gte": newPrefix, "$lte": uid + "/"},
	}
	nextMark := ""

	if limit == 0 {
		limit = 1000
	}

	// query
	r := []NamespaceInfo{}
	err := db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			var err error
			if err = c.Find(query).Sort(ITEM_ID).Limit(limit + 1).All(&r); err != nil {
				return fmt.Errorf("query failed")
			}
			return nil
		},
	)
	if err != nil {
		return []NamespaceInfo{}, "", err
	}

	var encoded string
	count := len(r)
	if len(r) > limit {
		nextMark = r[limit].Space
		encoded = base64.URLEncoding.EncodeToString([]byte(nextMark))
		count = len(r) - 1
	}
	return r[0:count], encoded, nil

}

func (m *NamespaceModel) UpdateBucket(xl *xlog.Logger, uid, space, bucket, domain string) error {
	/*
	   db.namespace.update({"uid": uid, "namespace": space}, bson.M{"$set":{"bucket": bucket, "domain" : domain }}),
	*/
	return db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					NAMESPACE_ITEM_ID: space,
					ITEM_ID:           uid + "." + space,
				},
				bson.M{
					"$set": bson.M{
						NAMESPACE_ITEM_BUCKET: bucket,
						ITEM_UPDATA_TIME:      time.Now().Unix(),
					},
				},
			)
		},
	)
}

func (m *NamespaceModel) UpdateAutoCreateUa(xl *xlog.Logger, uid, space string, auto bool) error {
	/*
	   db.namespace.update({"uid": uid, "namespace": space}, bson.M{"$set":{"autocreateua": auto}}),
	*/
	return db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					NAMESPACE_ITEM_ID: space,
					ITEM_ID:           uid + "." + space,
				},
				bson.M{
					"$set": bson.M{
						NAMESPACE_ITEM_AUTO_CREATE_UA: auto,
						ITEM_UPDATA_TIME:              time.Now().Unix(),
					},
				},
			)
		},
	)
}

func (m *NamespaceModel) UpdateExpire(xl *xlog.Logger, uid, space string, expire int) error {
	/*
	   db.namespace.update({"uid": uid, "namespace": space}, bson.M{"$set":{"expire": expire}}),
	*/
	return db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					NAMESPACE_ITEM_ID: space,
					ITEM_ID:           uid + "." + space,
				},
				bson.M{
					"$set": bson.M{
						NAMESPACE_ITEM_EXPIRE: expire,
					},
				},
			)
		},
	)
}

func (m *NamespaceModel) UpdateNamespace(xl *xlog.Logger, uid, space, newSpace string) error {
	/*
	   db.namespace.update({"uid": uid, "namespace": space}, bson.M{"$set":{"namespace": newSpace}}),
	*/
	r := []NamespaceInfo{}
	err := db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			return c.Find(
				bson.M{
					ITEM_ID: uid + "." + space,
				},
			).All(&r)
		},
	)
	if err != nil {
		return err
	}
	err = db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			return c.Remove(
				bson.M{
					ITEM_ID: uid + "." + space,
				},
			)
		},
	)
	if err != nil {
		return err
	}
	if len(r) == 0 {
		return fmt.Errorf("Can't find old namespace")
	}
	err = db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			_, err := c.Upsert(
				bson.M{
					ITEM_ID: uid + "." + newSpace,
				},
				bson.M{
					"$set": bson.M{
						ITEM_ID:                       uid + "." + newSpace,
						ITEM_CREATE_TIME:              r[0].Regtime,
						NAMESPACE_ITEM_ID:             newSpace,
						NAMESPACE_ITEM_BUCKET:         r[0].Bucket,
						ITEM_UPDATA_TIME:              time.Now().Unix(),
						NAMESPACE_ITEM_UID:            r[0].Uid,
						NAMESPACE_ITEM_AUTO_CREATE_UA: r[0].AutoCreateUa,
						NAMESPACE_ITEM_EXPIRE:         r[0].Expire,
					},
				},
			)
			return err
		},
	)
	if err != nil {
		return err
	}
	return nil

}
