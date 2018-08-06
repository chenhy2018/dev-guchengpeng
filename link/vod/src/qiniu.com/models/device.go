package models

import (
	"fmt"
        "qiniu.com/db"
        "gopkg.in/mgo.v2"
        "gopkg.in/mgo.v2/bson"
	"time"
)


const (
        DEVICE_COL = "device"
	DEVICE_ITEM_UUID   = "uuid"
	DEVICE_ITEM_DEVICEID = "deviceid"
	DEVICE_ITEM_DATE   = "date"
        DEVICE_ITEM_BUCKET_URL = "bucketurl"
	DEVICE_ITEM_EXPIRE = "remaindays"
)

type deviceModel struct {
}

var (
	Device *deviceModel
)

func (m *deviceModel) Init() error {

//     index := Index{
//         Key: []string{"uuid"},
//         Unique: true,
//         DropDups: true,
//         Background: true, // See notes.
//         Sparse: true,
//     }
//     err := collection.EnsureIndex(index)

       index := mgo.Index{
           Key: []string{DEVICE_ITEM_UUID},
           Unique: true,
           DropDups: true,
           Background: true, // See notes.
           Sparse: true,
       }

        return db.WithCollection(
                DEVICE_COL,
                func(c *mgo.Collection) error {
                        return c.EnsureIndex(index)
                },
        )
}

type RegisterReq struct {
        Uuid string
        Deviceid string
        BucketUrl string
        RemainDays int64
}

func (m *deviceModel) Register(req RegisterReq) error {

	err := db.WithCollection(
		DEVICE_COL,
		func(c *mgo.Collection) error {
			_, err := c.Upsert(
				bson.M{
                                        DEVICE_ITEM_UUID: req.Uuid,
					DEVICE_ITEM_DEVICEID: req.Deviceid,
				},
				bson.M{
					"$set": bson.M{
                                                DEVICE_ITEM_UUID: req.Uuid,
						DEVICE_ITEM_DEVICEID: req.Deviceid,
						DEVICE_ITEM_DATE: time.Now().Unix(),
                                                DEVICE_ITEM_BUCKET_URL: req.BucketUrl,
                                                DEVICE_ITEM_EXPIRE : req.RemainDays,
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

func (m *deviceModel) Delete(uuid,deviceid string) error {

	return db.WithCollection(
		DEVICE_COL,
		func(c *mgo.Collection) error {
			return c.Remove(
				bson.M{
					DEVICE_ITEM_UUID: uuid,
                                        DEVICE_ITEM_DEVICEID: deviceid,
				},
			)
		},
	)
}

func (m *deviceModel) UpdateRemaindays(uuid,deviceid string, remaindays int64) error {

	return db.WithCollection(
		DEVICE_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					DEVICE_ITEM_UUID: uuid,
                                        DEVICE_ITEM_DEVICEID: deviceid,
				},
				bson.M{
					"$set": bson.M{
						DEVICE_ITEM_EXPIRE: remaindays,
					},
				},
			)
		},
	)
}

type deviceInfo struct {
	UUID      string  `bson:"uuid"       json:"uuid"`
	DevicdID  string  `bson:"deviceid"   json:"deviceid"`
	Regtime   int     `bson:"date"       json:"date"`
        BucketUrl string  `bson:"bucketurl"  json:"bucketurl"`
	Expire    int64   `bson:"remaindays" json:"remaindays"`
}

func (m *deviceModel) GetDeviceInfo(index, rows int, category, like string) ([]deviceInfo, error) {

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
	r := []deviceInfo{}
	count := 0
	err := db.WithCollection(
		DEVICE_COL,
		func(c *mgo.Collection) error {
			var err error
			if err = c.Find(query).Sort(DEVICE_ITEM_EXPIRE).Skip(skip).Limit(limit).All(&r); err != nil {
				return fmt.Errorf("query failed")
			}
			if count, err = c.Find(query).Count(); err != nil {
				return fmt.Errorf("query count failed")
			}
			return nil
		},
	)
	if err != nil {
		return []deviceInfo{}, err
	}
	return r, nil
}
