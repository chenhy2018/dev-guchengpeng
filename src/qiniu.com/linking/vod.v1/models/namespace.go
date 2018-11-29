package models

import (
	"encoding/base64"
	"fmt"
	"github.com/qiniu/xlog.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"qiniu.com/linking/vod.v1/db"
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
						NAMESPACE_ITEM_CATEGORY:       req.Category,
						NAMESPACE_ITEM_DOMAIN:         req.Domain,
						NAMESPACE_ITEM_REMARK:         req.Remark,
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

func (m *NamespaceModel) Update(xl *xlog.Logger, req NamespaceInfo) error {
	/*
	   db.namespace.update( {"uid":req.Uid,  "namespace": req.Space}, {"$set": {"bucketurl": req.Bucketurl}},
	   { upsert: true })
	*/
	err := db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			err := c.Update(
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
						NAMESPACE_ITEM_CATEGORY:       req.Category,
						NAMESPACE_ITEM_DOMAIN:         req.Domain,
						NAMESPACE_ITEM_REMARK:         req.Remark,
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
	Id           string `bson:"_id"  json:"-"`
	Space        string `bson:"namespace"  json:"namespace"`
	Regtime      int64  `bson:"createdAt"  json:"createdAt"`
	UpdateTime   int64  `bson:"updatedAt"  json:"updatedAt"`
	Bucket       string `bson:"bucket"     json:"bucket"`
	Uid          string `bson:"uid"        json:"-"`
	Domain       string `bson:"domain"     json:"domain"`
	AutoCreateUa bool   `bson:"auto"       json:"auto"`
	Expire       int    `bson:"expire"     json:"expire"`
	Category     string `bson:"category"   json:"category"`
	Remark       string `bson:"remark"     json:"remark"`
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

func (m *NamespaceModel) GetNamespaceByBucket(xl *xlog.Logger, uid, bucket string) ([]NamespaceInfo, error) {
	/*
	   db.namespace.find({"bucket": bucket})
	*/
	r := []NamespaceInfo{}
	err := db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			return c.Find(
				bson.M{
					NAMESPACE_ITEM_UID:    uid,
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

func (m *NamespaceModel) UpdateFunction(xl *xlog.Logger, uid, namespace, parameter string, cond map[string]interface{}) error {
	/*
	   db.ua.update({_id: id, uaid: uaid}, bson.M{"$set":{parameter: cond[parameter]}}),
	*/
	return db.WithCollection(
		NAMESPACE_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					ITEM_ID:           uid + "." + namespace,
					NAMESPACE_ITEM_ID: namespace,
				},
				bson.M{
					"$set": bson.M{
						ITEM_UPDATA_TIME: time.Now().Unix(),
						parameter:        cond[parameter],
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
						NAMESPACE_ITEM_ID:             newSpace,
						NAMESPACE_ITEM_BUCKET:         r[0].Bucket,
						ITEM_UPDATA_TIME:              time.Now().Unix(),
						NAMESPACE_ITEM_UID:            r[0].Uid,
						NAMESPACE_ITEM_AUTO_CREATE_UA: r[0].AutoCreateUa,
						NAMESPACE_ITEM_EXPIRE:         r[0].Expire,
						NAMESPACE_ITEM_CATEGORY:       r[0].Category,
						NAMESPACE_ITEM_DOMAIN:         r[0].Domain,
						NAMESPACE_ITEM_REMARK:         r[0].Remark,
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
