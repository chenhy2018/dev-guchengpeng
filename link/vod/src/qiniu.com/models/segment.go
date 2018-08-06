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

func (m *SegmentModel) Init() error {
        /*
             index := Index{
                   Key: []string{"uuid", "deviceid"},
                   Unique: true,
                   DropDups: true,
                   Background: true, // See notes.
                   Sparse: true,
             }
             err := collection.EnsureIndex(index)
        */

        index := mgo.Index{
            Key: []string{SEGMENT_ITEM_UUID, SEGMENT_ITEM_DEVICEID},
            Unique: true,
            DropDups: true,
            Background: true, // See notes.
            Sparse: true,
        }
        return db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        return c.EnsureIndex(index)
                },
        )
}

func (m *SegmentModel) AddSegmentTS(req SegmentTsInfo) error {
        /*
                 collection.update(bson.M{uuid: id, deviceid: id, xxx}, bson.M{"$set": bson.M{"expire": time}},
                 { upsert: true })
        */

	err := db.WithCollection(
		SEGMENT_COL,
		func(c *mgo.Collection) error {
			_, err := c.Upsert(
				bson.M{
                                        SEGMENT_ITEM_UUID: req.Uuid,
					SEGMENT_ITEM_DEVICEID: req.DeviceId,
                                        SEGMENT_ITEM_FRAGMENT_START_TIME: req.FragmentStartTime,
                                        SEGMENT_ITEM_START_TIME: req.StartTime,
                                        SEGMENT_ITEM_END_TIME: req.EndTime,
				},
                                bson.M{
                                        "$set": bson.M{
                                                SEGMENT_ITEM_UUID: req.Uuid,
                                                SEGMENT_ITEM_DEVICEID: req.DeviceId,
                                                SEGMENT_ITEM_FRAGMENT_START_TIME: req.FragmentStartTime,
                                                SEGMENT_ITEM_START_TIME: req.StartTime,
                                                SEGMENT_ITEM_END_TIME: req.EndTime,
                                                SEGMENT_ITEM_FILE_NAME: req.FileName,
                                                SEGMENT_ITEM_EXPIRE: req.Expire,
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

        /*
                 db.collection.remove({uid:"bbc", "deviceid": "", xxx},
                 { justOne: false } )
        */

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
        /*
                 db.collection.update({uid:"bbc", "deviceid": "", xxx},
                 { upsert: true } )
        */
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
	Uuid      string  `bson:"uuid"       json:"uuid"`
	DeviceId  string  `bson:"deviceid"   json:"deviceid"`
	FragmentStartTime   int64 `bson:"fragmentstarttime" json:"fragmentstarttime"`
        StartTime int64  `bson:"starttime"  json:"starttime"`
        FileName string `bson:"filename"  json:"filename"`
        EndTime   int64  `bson:"endtime"  json:"endtime"`
	Expire    int64  `bson:"expire" json:"expire"`
}

func (m *SegmentModel) GetSegmentTsInfo(index, rows int, starttime,endtime int64, uuid,deviceid string) ([]SegmentTsInfo, error) {

        /*
                 db.collection.find(bson.M{uid:"bbc", "deviceid": "aaa", "starttime": bson.M{"$gte":starttime}, "endtime": bson.M{"$lte":endtime} },
                 ).sort("starttime").limit(rows),skip(rows * index)
        */
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
        err := db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        var err error
                        if limit > 0 {
                                err = c.Find(query).Sort(SEGMENT_ITEM_START_TIME).Skip(skip).Limit(limit).All(&r);
                        } else {
                                err = c.Find(query).Sort(SEGMENT_ITEM_START_TIME).Skip(skip).All(&r);
                        }
                        return err
                 },
        )
        if err != nil {
                return []SegmentTsInfo{}, err
        }
	return r, nil
}
func (m *SegmentModel) GetLastSegmentTsInfo(uuid,deviceid string) (SegmentTsInfo, error) {
        /*
                 db.collection.find(bson.M{uid:"bbc", "deviceid": "aaa", "starttime": bson.M{"$gte":starttime}, "endtime": bson.M{"$lte":endtime} },
                 ).sort("starttime").limit(rows),skip(rows * index)
        */
        // query by keywords
        query := bson.M{
                 SEGMENT_ITEM_UUID:uuid,
                 SEGMENT_ITEM_DEVICEID:deviceid,
        }
        r := SegmentTsInfo{}
        err := db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        err := c.Find(query).Sort("-starttime").Limit(1).One(&r);
                        return err
                 },
        )
        if err != nil {
                return SegmentTsInfo{}, err
        }
        return r, nil
}

type FragmentInfo struct {
        FragmentStartTime   int64 `bson:"fragmentstartTime" json:"fragmentstartTime"`
        StartTime int64  `bson:"starttime"  json:"starttime"`
        EndTime   int64  `bson:"endtime"  json:"endtime"`
}


func searchLastSegmentbyFragmentId(arr []SegmentTsInfo, low, high int, k int64) int {
	if low < 0 || high < 0 {
		return -1
	}
	for low < high {
		mid := low + (high-low)>>1
                //fmt.Printf("%d %d %d FragmentStartTime %d %d %d\n", mid, low, high, arr[mid].StartTime, arr[mid + 1].StartTime, k)
		if k < arr[mid].FragmentStartTime {
                        high = mid - 1
		} else if k > arr[mid].FragmentStartTime {
                        low = mid + 1
		} else if (k == arr[mid].FragmentStartTime && k < arr[mid + 1].FragmentStartTime) {
		        return mid + 1
		} else {
                        low = mid + 1
                }
	}
	return -1
}

func (m *SegmentModel) GetFragmentTsInfo(index, rows int, starttime,endtime int64, uuid,deviceid string) ([]FragmentInfo, error) {

        /*
                 db.collection.find(bson.M{uid:"bbc", "deviceid": "aaa", "startfragmenttime": bson.M{"$gte":starttime, "$lte": endtime} },
                 ).sort("starttime").limit(rows),skip(rows * index)
        */
        // query by keywords
        query := bson.M{
                 SEGMENT_ITEM_UUID:uuid,
                 SEGMENT_ITEM_DEVICEID:deviceid,
                 SEGMENT_ITEM_FRAGMENT_START_TIME: bson.M{"$gte":starttime, "$lte": endtime},
        }
        skip := rows * index
        limit := rows
        r := []SegmentTsInfo{}
        err := db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        var err error
                        if limit > 0 {
                                err = c.Find(query).Sort(SEGMENT_ITEM_FRAGMENT_START_TIME).Skip(skip).Limit(limit).All(&r);
                        } else {
                                err = c.Find(query).Sort(SEGMENT_ITEM_FRAGMENT_START_TIME).Skip(skip).All(&r);
                        }
                        return err
                 },
        )
        if err != nil {
                return []FragmentInfo{}, err
        }

        info := []FragmentInfo{}
        fragmentStartTime := r[0].FragmentStartTime
        low := 0
        high := len(r) - 1
        for low <= high && low >= 0 {
                low = searchLastSegmentbyFragmentId(r, low, high, fragmentStartTime);
                var one = FragmentInfo{}
                if (low == -1) {
                       one = FragmentInfo{ FragmentStartTime:fragmentStartTime,   StartTime: fragmentStartTime, EndTime: r[high].EndTime}
                } else {
                       one = FragmentInfo{ FragmentStartTime:fragmentStartTime,   StartTime: fragmentStartTime, EndTime: r[low - 1].EndTime}
                       fragmentStartTime = r[low].StartTime
                }
                info = append(info, one);
                
        }
        return info, nil
}
