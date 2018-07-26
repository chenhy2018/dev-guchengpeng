package models

import (
	"fmt"
        "qiniu.com/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"time"
)


const (
        SEGMENT_COL = "segment"
	SEGMENT_ITEM_UUID   = "uuid"
	SEGMENT_ITEM_DEVICEID = "deviceid"
        SEGMENT_ITEM_FRAGMENT_START_TIME = "fragmentstarttime"
        SEGMENT_ITEM_START_TIME = "starttime"
        SEGMENT_ITEM_END_TIME = "endtime"
        SEGMENT_ITEM_FILE_NAME = "filename"
	SEGMENT_ITEM_EXPIRE = "expire"
)

type SegmentModel struct {
}

var (
	Segment *SegmentModel
)

func init() {

	Segment = &SegmentModel{}
}

type SegmentReq struct {
        Uuid string
        Deviceid string
        FragmentStartTime int64
        StartTime int64
        EndTime int64
        FileName string
        ExpireDay int64
}

func (m *SegmentModel) AddSegmentTS(req SegmentReq) error {

	err := db.WithCollection(
		SEGMENT_COL,
		func(c *mgo.Collection) error {
			_, err := c.Upsert(
				bson.M{
                                        SEGMENT_ITEM_UUID: req.Uuid,
					SEGMENT_ITEM_DEVICEID: req.Deviceid,
                                        SEGMENT_ITEM_FRAGMENT_START_TIME: req.FragmentStartTime,
                                        SEGMENT_ITEM_START_TIME: req.StartTime,
                                        SEGMENT_ITEM_END_TIME: req.EndTime,
				},
                                bson.M{
                                        "$set": bson.M{
                                                SEGMENT_ITEM_UUID: req.Uuid,
                                                SEGMENT_ITEM_DEVICEID: req.Deviceid,
                                                SEGMENT_ITEM_FRAGMENT_START_TIME: req.FragmentStartTime,
                                                SEGMENT_ITEM_START_TIME: req.StartTime,
                                                SEGMENT_ITEM_END_TIME: req.EndTime,
                                                SEGMENT_ITEM_FILE_NAME: req.FileName,
                                                SEGMENT_ITEM_EXPIRE: req.ExpireDay,
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

func (m *SegmentModel) DeleteSegmentTS(uuid,deviceid string, starttime,endtime int64) error {

	return db.WithCollection(
		SEGMENT_COL,
		func(c *mgo.Collection) error {
			c.RemoveAll(
				bson.M{
					SEGMENT_ITEM_UUID: uuid,
                                        SEGMENT_ITEM_DEVICEID: deviceid,
                                        SEGMENT_ITEM_START_TIME: bson.M{"$gte":starttime},
                                        SEGMENT_ITEM_END_TIME: bson.M{"$lte":endtime},
				},
			)
                        return nil;
		},
	)
}

func (m *SegmentModel) UpdateSegmentTSExpire(uuid,deviceid string, starttime,endtime, expire int64) error {

	return db.WithCollection(
		SEGMENT_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
					SEGMENT_ITEM_UUID: uuid,
                                        SEGMENT_ITEM_DEVICEID: deviceid,
                                        SEGMENT_ITEM_START_TIME: starttime,
                                        SEGMENT_ITEM_END_TIME: endtime,
				},
				bson.M{
					"$set": bson.M{
						SEGMENT_ITEM_EXPIRE: expire,
					},
				},
			)
		},
	)
}

type SegmentTsInfo struct {
	UUID      string  `bson:"uuid"       json:"uuid"`
	DevicdID  string  `bson:"deviceid"   json:"deviceid"`
	FragmentStartTime   int `bson:"fragmentstartTime" json:"fragmentstartTime"`
        StartTime int64  `bson:"starttime"  json:"starttime"`
        FileName string `bson:"filename"  json:"filename"`
        EndTime   int64  `bson:"endtime"  json:"endtime"`
	Expire    int64  `bson:"expire" json:"expire"`
}

func (m *SegmentModel) GetSegmentTsInfo(index, rows int, starttime,endtime int64, uuid,deviceid string) ([]SegmentTsInfo, error) {

	// query by keywords
        query := bson.M{
                 SEGMENT_ITEM_UUID:uuid,
                 SEGMENT_ITEM_DEVICEID:deviceid,
                 SEGMENT_ITEM_START_TIME: bson.M{"$gte":starttime},
                 SEGMENT_ITEM_END_TIME : bson.M{"$lte":endtime},
        }
        skip := rows * index
        limit := rows
        r := []SegmentTsInfo{}
        count := 0
        err := db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        var err error
                        if limit > 0 {
                                err = c.Find(query).Sort(SEGMENT_ITEM_START_TIME).Skip(skip).Limit(limit).All(&r);
                        } else {
                                err = c.Find(query).Sort(SEGMENT_ITEM_START_TIME).Skip(skip).All(&r);
                        }
                        if count, err = c.Find(query).Count(); err != nil {
                                return fmt.Errorf("query count failed")
                        }
                        return nil
                 },
        )
        if err != nil {
                return []SegmentTsInfo{}, err
        }
	return r, nil
}

func (m *SegmentModel) GetFragmentTsInfo(index, rows int, starttime,endtime int64, uuid,deviceid string) ([]SegmentTsInfo, error) {
        // query by keywords
        query := bson.M{
                 SEGMENT_ITEM_UUID:uuid,
                 SEGMENT_ITEM_DEVICEID:deviceid,
                 SEGMENT_ITEM_FRAGMENT_START_TIME: bson.M{"$gte":starttime, "$lte": endtime},
        }
        skip := rows * index
        limit := rows
        r := []SegmentTsInfo{}
        count := 0
        err := db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        var err error
                        if limit > 0 {
                                err = c.Find(query).Sort(SEGMENT_ITEM_START_TIME).Skip(skip).Limit(limit).All(&r);
                        } else {
                                err = c.Find(query).Sort(SEGMENT_ITEM_START_TIME).Skip(skip).All(&r);
                        }
                        if count, err = c.Find(query).Count(); err != nil {
                                return fmt.Errorf("query count failed")
                        }
                        return nil
                 },
        )
        if err != nil {
                return []SegmentTsInfo{}, err
        }
        return r, nil

}
