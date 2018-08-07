package models

import (
        //"fmt"
        "strconv"
        "qiniu.com/db"
        "gopkg.in/mgo.v2"
        "gopkg.in/mgo.v2/bson"
	//"time"
)

const (
        SEGMENT_COL = "segment"
	SEGMENT_ITEM_ID   = "_id"
	SEGMENT_ITEM_UAID   = "uaid"
        SEGMENT_ITEM_FRAGMENT_START_TIME = "fragmentstarttime"
        SEGMENT_ITEM_START_TIME = "starttime"
        SEGMENT_ITEM_END_TIME = "endtime"
        SEGMENT_ITEM_FILE_NAME = "filename"
	SEGMENT_ITEM_EXPIRE = "expireAfterdays"
)

type SegmentModel struct {
}

var (
	Segment *SegmentModel
)

func (m *SegmentModel) Init() error {
        /*
             index := Index{
                   Key: []string{"uid", "uaid"},
                   Unique: false,
                   DropDups: true,
                   Background: true, // See notes.
                   Sparse: true,
             }
             err := collection.EnsureIndex(index)
        */

        index := mgo.Index{
            Key: []string{SEGMENT_ITEM_EXPIRE},
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
                 collection.update(bson.M{uid: id, uaid: id, xxx}, bson.M{"$set": bson.M{"expireAfterdays": time}},
                 { upsert: true })
        */

	err := db.WithCollection(
		SEGMENT_COL,
		func(c *mgo.Collection) error {
			_, err := c.Upsert(
				bson.M{
                                        SEGMENT_ITEM_ID: req.Uid + "." + req.UaId + "." + strconv.FormatInt(req.StartTime,10),
                                        SEGMENT_ITEM_FRAGMENT_START_TIME: req.FragmentStartTime,
                                        SEGMENT_ITEM_START_TIME: req.StartTime,
                                        SEGMENT_ITEM_END_TIME: req.EndTime,
				},
                                bson.M{
                                        "$set": bson.M{
                                                SEGMENT_ITEM_ID: req.Uid + "." + req.UaId + "." + strconv.FormatInt(req.StartTime,10),
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

func (m *SegmentModel) DeleteSegmentTS(uid,uaid string, starttime,endtime int64) error {

        /*
                 db.collection.remove({uid:"bbc", "uaid": "", xxx},
                 { justOne: false } )
        */

	return db.WithCollection(
		SEGMENT_COL,
		func(c *mgo.Collection) error {
			c.RemoveAll(
				bson.M{
					SEGMENT_ITEM_ID: bson.M{"$regex": uid + "." + uaid + ".*"},
                                        SEGMENT_ITEM_START_TIME : bson.M{ "$gte" : starttime},
                                        SEGMENT_ITEM_END_TIME : bson.M{ "$lte" :  endtime},
				},
			)
                        return nil;
		},
	)
}

func (m *SegmentModel) UpdateSegmentTSExpire(uid,uaid string, starttime,endtime, expire int64) error {
        /*
                 db.collection.update({uid:"bbc", "uaid": "", xxx},
                 { upsert: true } )
        */
	return db.WithCollection(
		SEGMENT_COL,
		func(c *mgo.Collection) error {
			return c.Update(
				bson.M{
                                        SEGMENT_ITEM_ID: bson.M{"$regex": uid + "." + uaid + ".*"},
                                        SEGMENT_ITEM_START_TIME : bson.M{ "$gte" : starttime},
                                        SEGMENT_ITEM_END_TIME : bson.M{ "$lte" :  endtime},
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
	Uid                 string `bson:"uid"                json:"uid"`
	UaId                string `bson:"uaid"               json:"uaid"`
	FragmentStartTime   int64  `bson:"fragmentstarttime"  json:"fragmentstarttime"`
        StartTime           int64  `bson:"starttime"          json:"starttime"`
        FileName            string `bson:"filename"           json:"filename"`
        EndTime             int64  `bson:"endtime"            json:"endtime"`
	Expire              int64  `bson:"expireAfterdays"    json:"expireAfterdays"`
}

func (m *SegmentModel) GetSegmentTsInfo(index, rows int, starttime,endtime int64, uid,uaid string) ([]SegmentTsInfo, error) {

        /*
                 db.collection.find(bson.M{uid:"bbc", "uaid": "aaa", "starttime": bson.M{"$gte":starttime}, "endtime": bson.M{"$lte":endtime} },
                 ).sort("starttime").limit(rows),skip(rows * index)
        */
	// query by keywords
        query := bson.M{
               SEGMENT_ITEM_ID: bson.M{"$regex": uid + "." + uaid + ".*"},
               SEGMENT_ITEM_START_TIME : bson.M{ "$gte" : starttime},
               SEGMENT_ITEM_END_TIME : bson.M{ "$lte" :  endtime},
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
func (m *SegmentModel) GetLastSegmentTsInfo(uid,uaid string) (SegmentTsInfo, error) {
        /*
                 db.collection.find(bson.M{uid:"bbc", "uaid": "aaa", "starttime": bson.M{"$gte":starttime}, "endtime": bson.M{"$lte":endtime} },
                 ).sort("starttime").limit(rows),skip(rows * index)
        */
        // query by keywords
        // query by keywords
        query := bson.M{}
        query[SEGMENT_ITEM_ID] = bson.M{
                        "$regex": uid + "." + uaid + ".*",
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
        FragmentStartTime   int64 `bson:"_id" json:"_id"`
        StartTime int64  `bson:"starttime"  json:"starttime"`
        EndTime   int64  `bson:"endtime"  json:"endtime"`
}

func (m *SegmentModel) GetFragmentTsInfo(index, rows int, starttime,endtime int64, uid,uaid string) ([]FragmentInfo, error) {

        /*
                 query = []bson.M{
                        {"$match": bson.M{
                                SEGMENT_ITEM_UID:  uid,
                                SEGMENT_ITEM_UAID: uaid,
                                SEGMENT_ITEM_FRAGMENT_START_TIME: bson.M{"$gte":starttime, "$lte": endtime},},
                        },
                        {"$group": bson.M{
                                "_id": bson.M{SEGMENT_ITEM_FRAGMENT_START_TIME : "$fragmentstarttime"},
                                SEGMENT_ITEM_START_TIME : bson.M{ "$min" :  "$starttime"},
                                SEGMENT_ITEM_END_TIME : bson.M{ "$max" :  "$endtime"},},
                        },
                        {"$sort" : bson.M{SEGMENT_ITEM_START_TIME : 1},},
                        {"$skip" : skip},
                        {"$limit": limit},
                 },

                 db.collection.aggregate(query)
        */
        // query by keywords
        skip := rows * index
        limit := rows
        if (rows == 0) {
              limit = 200
        }
        r := []FragmentInfo{}
        err := db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        err := c.Pipe(
                                []bson.M{
                                        {"$match": bson.M{
                                                 SEGMENT_ITEM_ID:  bson.M{"$regex": uid + "." + uaid + ".*"},
                                                 SEGMENT_ITEM_FRAGMENT_START_TIME: bson.M{"$gte":starttime, "$lte": endtime},},
                                        },
                                        {"$group": bson.M{
                                                 "_id": bson.M{SEGMENT_ITEM_FRAGMENT_START_TIME : "$fragmentstarttime"},
                                                 SEGMENT_ITEM_START_TIME : bson.M{ "$min" :  "$starttime"},
                                                 SEGMENT_ITEM_END_TIME : bson.M{ "$max" :  "$endtime"},},
                                        },
                                        {"$sort" : bson.M{SEGMENT_ITEM_START_TIME : 1},},
                                        {"$skip" : skip},
                                        {"$limit": limit},
                                },
                        ).All(&r)
                        if err != nil {
                                return err
                        }
                        //r = result["group"]
                        return nil
                 },
        )
        //fmt.Printf("fagment time %d start time %d end time %d ", r[0].FragmentStartTime, r[1].FragmentStartTime, r[1].StartTime );
        if err != nil {
                return []FragmentInfo{}, err
        }

        if (len(r) < 1) {
               return []FragmentInfo{}, nil
        }

        return r, nil
}
