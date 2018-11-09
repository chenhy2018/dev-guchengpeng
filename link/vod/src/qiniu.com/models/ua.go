package models

import (
	"fmt"
	"time"

	"github.com/qiniu/xlog.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"qiniu.com/db"
)

type UaModel struct {
}

var (
	Ua *UaModel
)

func (m *UaModel) Init() error {
	return nil
}

func (m *UaModel) Register(xl *xlog.Logger, req UaInfo) error {
	/*
	   db.ua.update( {id: "uid" + "namespace + "." + req.UaId"}, {"$set": {"namespace": space, "password": password}},
	   { upsert: true })
	*/
	err := db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			_, err := c.Upsert(
				bson.M{
					ITEM_ID: req.Uid + "." + req.Namespace + "." + req.UaId,
				},
				bson.M{
					"$set": bson.M{
						ITEM_ID:           req.Uid + "." + req.Namespace + "." + req.UaId,
						UA_ITEM_UID:       req.Uid,
						UA_ITEM_UAID:      req.UaId,
						UA_ITEM_PASSWORD:  req.Password,
						ITEM_CREATE_TIME:  time.Now().Unix(),
						UA_ITEM_NAMESPACE: req.Namespace,
						ITEM_UPDATA_TIME:  time.Now().Unix(),
						UA_ITEM_VOD:       req.Vod,
						UA_ITEM_LIVE:      req.Live,
						UA_ITEM_ONLINE:    req.Online,
						UA_ITEM_EXPIRE:    req.Expire,
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

func (m *UaModel) Delete(xl *xlog.Logger, cond map[string]interface{}) error {
	/*
	   db.ua.remove({id: "uid + "." + "namespace" + "." + id", uaid: id})
	*/
	return db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			return c.Remove(
				cond,
			)
		},
	)
}

type UaInfo struct {
	id        string `bson:"_id"       json:"_id"`
	Uid       string `bson:"uid"       json:"uid"`
	UaId      string `bson:"uaid"       json:"uaid"`
	Password  string `bson:"password"   json:"password"` //options
	Namespace string `bson:"namespace"  json:"namespace"`
	CreateAt  int64  `bson:"createdAt"  json:"createdAt"`
	UpdatedAt int64  `bson:"updateAt"   json:"updateAt"`
	Vod       bool   `bson:"vod"        json:"vod"`
	Live      bool   `bson:"live"       json:"live"`
	Online    bool   `bson:"online"     json:"online"`
	Expire    int    `bson:"expire"     json:"expire"`
}

func (m *UaModel) GetUaInfos(xl *xlog.Logger, limit int, mark, uid, namespace, prefix string) ([]UaInfo, string, error) {

	/*
	   db.ua.find({"uid": uid, {"_id": "$gte": newPrefix, "$lte": uid + "." + namespace + "/"},}
	   ).sort({"date":1}).limit(limit)
	*/

	newPrefix := uid + "." + namespace + "." + prefix
	if mark != "" {
		newPrefix = mark
	}
	var query = bson.M{}
	// query by keywords
	query = bson.M{
		ITEM_ID:           bson.M{"$gte": newPrefix, "$lte": uid + "/"},
		UA_ITEM_NAMESPACE: namespace,
	}

	// direct to specific page
	nextMark := ""

	if limit == 0 {
		limit = 65535
	}
	// query
	r := []UaInfo{}
	count := 0
	err := db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			var err error
			if err = c.Find(query).Sort(ITEM_ID).Limit(limit).All(&r); err != nil {
				return fmt.Errorf("query failed")
			}
			if count, err = c.Find(query).Count(); err != nil {
				return fmt.Errorf("query count failed")
			}
			return nil
		},
	)
	if err != nil {
		return []UaInfo{}, "", err
	}
	if count > limit {
		nextMark = r[limit-1].Uid + "." + r[limit-1].Namespace + "." + r[limit-1].UaId + "."
	}
	return r, nextMark, nil
}

func (m *UaModel) GetUaInfo(xl *xlog.Logger, uid, namespace, uaid string) ([]UaInfo, error) {
	/*
	   db.ua.find({namespace: namespace, uaid: id})
	*/
	// query by keywords
	query := bson.M{
		ITEM_ID:      uid + "." + namespace + "." + uaid,
		UA_ITEM_UAID: uaid,
	}

	// query
	r := []UaInfo{}
	err := db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			var err error
			if err = c.Find(query).All(&r); err != nil {
				return fmt.Errorf("query failed")
			}
			return nil
		},
	)
	if err != nil {
		return []UaInfo{}, err
	}
	return r, nil
}

func (m *UaModel) UpdateUa(xl *xlog.Logger, uid, namespace, uaid string, info UaInfo) error {
	/*
	   db.ua.update({namespace: space, uaid: uaid}, bson.M{"$set":{"namespace": space, "password": password}}),
	*/
	return db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					ITEM_ID:      uid + "." + namespace + "." + uaid,
					UA_ITEM_UAID: uaid,
				},
				bson.M{
					"$set": bson.M{
						UA_ITEM_PASSWORD:  info.Password,
						ITEM_UPDATA_TIME:  time.Now().Unix(),
						UA_ITEM_NAMESPACE: info.Namespace,
					},
				},
			)
		},
	)
}

func (m *UaModel) UpdateFunction(xl *xlog.Logger, uid, namespace, uaid, parameter string, cond map[string]interface{}) error {
	/*
	   db.ua.update({_id: id, uaid: uaid}, bson.M{"$set":{parameter: cond[parameter]}}),
	*/
	return db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					ITEM_ID:      uid + "." + namespace + "." + uaid,
					UA_ITEM_UAID: uaid,
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
