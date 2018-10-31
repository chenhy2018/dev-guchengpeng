package models

import (
	"fmt"
	"strconv"
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

	//     index := Index{
	//         Key: []string{"uid"},
	//     }
	//     db.collection.EnsureIndex(index)

	index := mgo.Index{
		Key: []string{UA_ITEM_UID},
	}

	return db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			return c.EnsureIndex(index)
		},
	)
}

func (m *UaModel) Register(xl *xlog.Logger, req UaInfo) error {
	/*
	   db.ua.update( {uaid: id, uid: uid, xxx}, {"$set": {"namespace": space, "password": password}},
	   { upsert: true })
	*/
	err := db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			_, err := c.Upsert(
				bson.M{
					UA_ITEM_UID:  req.Uid,
					UA_ITEM_UAID: req.UaId,
				},
				bson.M{
					"$set": bson.M{
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
	   db.ua.remove({"uid": uid, uaid: id})
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
	Uid       string `bson:"uid"        json:"uid"`
	UaId      string `bson:"uaid"       json:"uaid"`
	Password  string `bson:"password"   json:"password"` //options
	Namespace string `bson:"namespace"  json:"namespace"`
	CreateAt  int64  `bson:"createdAt"  json:"createdAt"`
	UpdatedAt int64  `bson:"updatedAt"  json:"updatedAt"`
	Vod       bool   `bson:"vod"        json:"vod"`
	Live      bool   `bson:"live"       json:"live"`
	Online    bool   `bson:"online"     json:"online"`
	Expire    int    `bson:"expire"     json:"expire"`
}

func (m *UaModel) GetUaInfos(xl *xlog.Logger, limit int, mark, uid, namespace, category, like string) ([]UaInfo, string, error) {

	/*
	   db.ua.find({"uid": uid, {category: {"$regex": "*like*"}},}
	   ).sort({"date":1}).limit(limit),skip(mark)
	*/
	var query = bson.M{}
	// query by keywords
	if namespace != "" {
		query = bson.M{
			UA_ITEM_UID:       uid,
			UA_ITEM_NAMESPACE: namespace,
			category:          bson.M{"$regex": ".*" + like + ".*"},
		}
	} else {
		query = bson.M{
			UA_ITEM_UID: uid,
			category:    bson.M{"$regex": ".*" + like + ".*"},
		}
	}
	// direct to specific page
	nextMark := ""
	// direct to specific page
	skip, err := strconv.ParseInt(mark, 10, 32)
	if err != nil {
		skip = 0
	}

	if limit == 0 {
		limit = 65535
	}

	// query
	r := []UaInfo{}
	count := 0
	err = db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			var err error
			if err = c.Find(query).Sort(UA_ITEM_UAID).Skip(int(skip)).Limit(limit).All(&r); err != nil {
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
		nextMark = fmt.Sprintf("%d", limit)
	}
	return r, nextMark, nil
}

func (m *UaModel) GetUaInfo(xl *xlog.Logger, uid, uaid string) ([]UaInfo, error) {
	/*
		db.ua.find({uid: uid, uaid: id})
	*/
	// query by keywords
	query := bson.M{
		UA_ITEM_UID:  uid,
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

func (m *UaModel) UpdateUa(xl *xlog.Logger, uid, uaid string, info UaInfo) error {
	/*
	   db.ua.update({uid: uid, uaid: uaid}, bson.M{"$set":{"namespace": space, "password": password}}),
	*/
	return db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					UA_ITEM_UID:  uid,
					UA_ITEM_UAID: uaid,
				},
				bson.M{
					"$set": bson.M{
						UA_ITEM_UID:       info.Uid,
						UA_ITEM_UAID:      info.UaId,
						UA_ITEM_PASSWORD:  info.Password,
						ITEM_UPDATA_TIME:  time.Now().Unix(),
						UA_ITEM_NAMESPACE: info.Namespace,
						UA_ITEM_VOD:       info.Vod,
						UA_ITEM_LIVE:      info.Live,
						UA_ITEM_ONLINE:    info.Online,
						UA_ITEM_EXPIRE:    info.Expire,
					},
				},
			)
		},
	)
}

func (m *UaModel) UpdateFunction(xl *xlog.Logger, uid, uaid, parameter string, cond map[string]interface{}) error {
	/*
	   db.ua.update({uid: uid, uaid: uaid}, bson.M{"$set":{parameter: cond[parameter]}}),
	*/
	return db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					UA_ITEM_UID:  uid,
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
