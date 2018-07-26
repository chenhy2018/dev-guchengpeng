package models

import (
	//"fmt"
        "qiniu.com/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"time"
)


const (
        SEGMENT_COL = "segment"
	SEGMENT_ITEM_UUID   = "uuid"
	SEGMENT_ITEM_DEVICEID = "deviceid"
        SEGMENT_ITEM_SEGMENT_START_TIME = "segment_start_time"
        SEGMENT_ITEM_START_TIME = "start_time"
        SEGMENT_ITEM_END_TIME = "end_time"
        SEGMENT_ITEM_FILE_NAME = "file_name"
	SEGMENT_ITEM_EXPIRE = "expire"
)

type segmentModel struct {
        
}

var (
	Segment *segmentModel
)

func init() {

	Segment = &segmentModel{}
}

type SegmentReq struct {
        Uuid string
        Deviceid string
        SegmentStartTime int64
        StartTime int64
        EndTime int64
        FileName string
        ExpireDay int64
}

func (m *segmentModel) AddSegmentTS(req SegmentReq) error {

	err := db.WithCollection(
		SEGMENT_COL,
		func(c *mgo.Collection) error {
			_, err := c.Upsert(
				bson.M{
                                        SEGMENT_ITEM_UUID: req.Uuid,
					SEGMENT_ITEM_DEVICEID: req.Deviceid,
                                        SEGMENT_ITEM_SEGMENT_START_TIME: req.SegmentStartTime,
                                        SEGMENT_ITEM_START_TIME: req.StartTime,
                                        SEGMENT_ITEM_END_TIME: req.EndTime,
				},
                                bson.M{
                                        "$set": bson.M{
                                                SEGMENT_ITEM_UUID: req.Uuid,
                                                SEGMENT_ITEM_DEVICEID: req.Deviceid,
                                                SEGMENT_ITEM_SEGMENT_START_TIME: req.SegmentStartTime,
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

func (m *segmentModel) DeleteSegmentTS(uuid,deviceid,starttime,endtime string) error {

	return db.WithCollection(
		SEGMENT_COL,
		func(c *mgo.Collection) error {
			return c.Remove(
				bson.M{
					SEGMENT_ITEM_UUID: uuid,
                                        SEGMENT_ITEM_DEVICEID: deviceid,
                                        SEGMENT_ITEM_START_TIME: starttime,
                                        SEGMENT_ITEM_END_TIME: endtime,
				},
			)
		},
	)
}

func (m *segmentModel) UpdateSegmentTSExpire(uuid,deviceid string, starttime,endtime, expire int64) error {

	return db.WithCollection(
		DEVICE_COL,
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

type SegmentTsDispInfo struct {
	UUID      string  `bson:"uuid"       json:"uuid"`
	DevicdID  string  `bson:"deviceid"   json:"deviceid"`
	SegmentStartTime   int     `bson:"segmentstartTime"       json:"segmentstartTime"`
        StartTime int64  `bson:"starttime"  json:"starttime"`
        FileName string `bson:"filename"  json:"filename"`
        EndTime   int64  `bson:"endtime"  json:"endtime"`
	Expire    int64  `bson:"expire" json:"expire"`
}

func (m *segmentModel) Display(index, rows, starttime,endtime int64, uuid,deviceid string) ([]SegmentTsDispInfo, error) {

	// query by keywords

	// direct to specific page
	//skip := rows * index
	limit := rows
	if limit > 100 {
		limit = 100
	}

	// query
	r := []SegmentTsDispInfo{}
	return r, nil
}
