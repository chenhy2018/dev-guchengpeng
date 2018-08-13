package models

import (
        "strconv"
        "qiniu.com/db"
        "gopkg.in/mgo.v2"
        "gopkg.in/mgo.v2/bson"
        "github.com/qiniu/xlog.v1"
        "time"
)

type SegmentModel struct {
}

var (
        Segment *SegmentModel
)

func (m *SegmentModel) Init() error {
        /*
             index := Index{
                   Key: []string{"expireAt"},
                   ExpireAfter : time.Second,
             }
             err := collection.EnsureIndex(index)
        */

        index := mgo.Index{
            Key: []string{SEGMENT_ITEM_EXPIRE},
            ExpireAfter : time.Second,
        }
        return db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        return c.EnsureIndex(index)
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
        Expire              time.Time  `bson:"expireAt"           json:"expireAt"`
}
func (m *SegmentModel) AddSegmentTS(xl *xlog.Logger, req SegmentTsInfo) error {
        /*
                 collection.update({"_id": req.Uid + "." + req.UaId + "." + strconv.FormatInt(req.StartTime,10) },
                                    "fragmentstarttime" : req.FragmentStartTime,
                                     "starttime" : req.StartTime,
                                     "endtime" : req.EndTime,}
                                   {"$set": bson.M{"expireAfterdays": time}},
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

func (m *SegmentModel) DeleteSegmentTS(xl *xlog.Logger, uid,uaid string, starttime,endtime int64) error {

        /*
                 db.collection.remove({"_id": {"$gte": uid + "." + uaid + ".", "$lte": uid + "." + uaid + "/"},
                                        "starttime": { "$gte" : starttime},
                                         "endtime" : { "$gte" : endtime},},
                 { justOne: false } )
        */

        return db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        c.RemoveAll(
                                bson.M{
                                        SEGMENT_ITEM_ID: bson.M{"$gte": uid + "." + uaid + ".", "$lte": uid + "." + uaid + "/"},
                                        SEGMENT_ITEM_START_TIME : bson.M{"$gte" : starttime},
                                        SEGMENT_ITEM_END_TIME : bson.M{"$lte" :  endtime},
                                },
                        )
                        return nil;
                },
	)
}

func (m *SegmentModel) UpdateSegmentTSExpire(xl *xlog.Logger, uid,uaid string, starttime,endtime, expire int64) error {
        /*
                 db.collection.update({"_id": {"$gte": uid + "." + uaid + ".", "$lte": uid + "." + uaid + "/"},
                                        "starttime": { "$gte" : starttime},
                                         "endtime" :{ "$lte" : endtime},},
                                      { $set : {"expireAfterdays" : expire} }
                                      { upsert: true } )
        */
	return db.WithCollection(
                SEGMENT_COL,
		func(c *mgo.Collection) error {
                        return c.Update(
                                bson.M{
                                        SEGMENT_ITEM_ID: bson.M{"$gte": uid + "." + uaid + ".", "$lte": uid + "." + uaid + "/"},
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

func (m *SegmentModel) GetSegmentTsInfo(xl *xlog.Logger, index, rows int, starttime,endtime int64, uid,uaid string) ([]map[string]interface{}, error) {

        /*
                 db.collection.find(bson.M{"_id": {"$gte": uid + "." + uaid + ".", "$lte": uid + "." + uaid + "/"},
                                           "starttime": {"$gte":starttime},
                                           "endtime": {"$lte":endtime} },
                 ).sort("starttime").limit(rows),skip(rows * index)
        */
	// query by keywords
        query := bson.M{
               SEGMENT_ITEM_ID: bson.M{"$gte": uid + "." + uaid + ".", "$lte": uid + "." + uaid + "/"},
               SEGMENT_ITEM_START_TIME : bson.M{ "$gte" : starttime},
               SEGMENT_ITEM_END_TIME : bson.M{ "$lte" :  endtime},
        }
        skip := rows * index
        limit := rows
        var r []map[string]interface{}

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
	return r, err
}
func (m *SegmentModel) GetLastSegmentTsInfo(xl *xlog.Logger, uid,uaid string) (map[string]interface{}, error) {
        /*
                 db.collection.find( {"_id": {"$gte": uid + "." + uaid + ".", "$lte": uid + "." + uaid + "/"},
                                      "starttime": {"$gte":starttime},
                                      "endtime": {"$lte":endtime} },
                 ).sort("starttime").limit(rows),skip(rows * index)
        */
        // query by keywords
        query := bson.M{
                SEGMENT_ITEM_ID: bson.M{"$gte": uid + "." + uaid + ".", "$lte": uid + "." + uaid + "/"},
        }
        var r map[string]interface{}
        err := db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        err := c.Find(query).Sort("-starttime").Limit(1).One(&r);
                        return err
                 },
        )
        return r, err
}

func (m *SegmentModel) GetFragmentTsInfo(xl *xlog.Logger, index, rows int, starttime,endtime int64, uid,uaid string) ([]map[string]interface{}, error) {

        /*
                 query ={
                        {"$match": {
                                "_id": {"$gte": uid + "." + uaid + ".", "$lte": uid + "." + uaid + "/"},
                                SEGMENT_ITEM_FRAGMENT_START_TIME: {"$gte":starttime, "$lte": endtime},},
                        },
                        {"$group": {
                                "_id": {SEGMENT_ITEM_FRAGMENT_START_TIME : "$fragmentstarttime"},
                                SEGMENT_ITEM_START_TIME : { "$min" :  "$starttime"},
                                SEGMENT_ITEM_END_TIME : { "$max" :  "$endtime"},},
                        },
                        {"$sort" : {SEGMENT_ITEM_START_TIME : 1},},
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
        var r []map[string]interface{}

        err := db.WithCollection(
                SEGMENT_COL,
                func(c *mgo.Collection) error {
                        err := c.Pipe(
                                []bson.M{
                                        {"$match": bson.M{
                                                 SEGMENT_ITEM_ID: bson.M{"$gte": uid + "." + uaid + ".", "$lte": uid + "." + uaid + "/"},
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
                        return nil
                 },
        )
        return r, err
}
