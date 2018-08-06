package models

import (
	"fmt"
        "qiniu.com/db"
        "gopkg.in/mgo.v2"
        "gopkg.in/mgo.v2/bson"
	"time"
)


const (
        UA_COL = "device"
	UA_ITEM_UUID   = "uuid"
	UA_ITEM_UAID = "uaid"
	UA_ITEM_DATE   = "date"
        UA_ITEM_BUCKET_URL = "bucketurl"
	UA_ITEM_EXPIRE = "remaindays"
)

type uaModel struct {
}

var (
	Ua *uaModel
)

func (m *uaModel) Init() error {

//     index := Index{
//         Key: []string{"uuid"},
//         Unique: true,
//         DropDups: true,
//         Background: true, // See notes.
//         Sparse: true,
//     }
//     err := collection.EnsureIndex(index)

       index := mgo.Index{
           Key: []string{UA_ITEM_UUID},
           Unique: true,
           DropDups: true,
           Background: true, // See notes.
           Sparse: true,
       }

        return db.WithCollection(
                UA_COL,
                func(c *mgo.Collection) error {
                        return c.EnsureIndex(index)
                },
        )
}

func (m *uaModel) Register(req UaInfo) error {

	err := db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			_, err := c.Upsert(
				bson.M{
                                        UA_ITEM_UUID: req.UuId,
					UA_ITEM_UAID: req.UaId,
				},
				bson.M{
					"$set": bson.M{
                                                UA_ITEM_UUID: req.UuId,
						UA_ITEM_UAID: req.UaId,
						UA_ITEM_DATE: time.Now().Unix(),
                                                UA_ITEM_BUCKET_URL: req.BucketUrl,
                                                UA_ITEM_EXPIRE : req.RemainDays,
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

func (m *uaModel) Delete(uuid,uaid string) error {

	return db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			return c.Remove(
				bson.M{
					UA_ITEM_UUID: uuid,
                                        UA_ITEM_UAID: uaid,
				},
			)
		},
	)
}

func (m *uaModel) UpdateRemaindays(uuid,uaid string, remaindays int64) error {

	return db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					UA_ITEM_UUID: uuid,
                                        UA_ITEM_UAID: uaid,
				},
				bson.M{
					"$set": bson.M{
						UA_ITEM_EXPIRE: remaindays,
					},
				},
			)
		},
	)
}

type UaInfo struct {
	UuId          string  `bson:"uuid"       json:"uuid"`
	UaId          string  `bson:"uaid"       json:"uaid"`
	Regtime       int     `bson:"date"       json:"date"`
        BucketUrl     string  `bson:"bucketurl"  json:"bucketurl"`
        RemainDays    int64   `bson:"remaindays" json:"remaindays"`
}

func (m *uaModel) GetUaInfo(index, rows int, category, like string) ([]UaInfo, error) {

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
	r := []UaInfo{}
	count := 0
	err := db.WithCollection(
		UA_COL,
		func(c *mgo.Collection) error {
			var err error
			if err = c.Find(query).Sort(UA_ITEM_DATE).Skip(skip).Limit(limit).All(&r); err != nil {
				return fmt.Errorf("query failed")
			}
			if count, err = c.Find(query).Count(); err != nil {
				return fmt.Errorf("query count failed")
			}
			return nil
		},
	)
	if err != nil {
		return []UaInfo{}, err
	}
	return r, nil
}
