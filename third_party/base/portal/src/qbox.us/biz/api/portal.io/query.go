package portalio

import (
	"labix.org/v2/mgo/bson"
	"qbox.us/biz/utils.v2/validator"
	"time"
)

var (
	DefaultSkip  = 0
	DefaultLimit = 10
)

type QueryId string

func (q QueryId) toBsonObjectId() bson.ObjectId {
	if bson.IsObjectIdHex(string(q)) {
		return bson.ObjectIdHex(string(q))
	}

	return ""
}

type Query struct {
	Id        QueryId   `param:"id"`
	Uid       uint32    `param:"uid"`
	Email     string    `param:"email"`
	StartTime time.Time `param:"stime"` // unix seconds or YYYY-MM-DD or YYYY-MM-DD hh:mm:ss or time.RFC3339
	EndTime   time.Time `param:"etime"` // unix seconds or YYYY-MM-DD or YYYY-MM-DD hh:mm:ss or time.RFC3339
	Skip      int       `param:"skip"`
	Limit     int       `param:"limit"`
	Sort      []string  `param:"sort"`
}

func (c Query) Valid() bool {
	if c.Email != "" && !validator.IsEmail(c.Email) {
		return false
	}

	return true
}

func (c Query) GetSkip() int {
	if c.Skip <= 0 {
		return DefaultSkip
	}

	return c.Skip
}

func (c Query) GetLimit() int {
	if c.Limit <= 0 {
		return DefaultLimit
	}

	return c.Limit
}

func (c Query) GetBsonQuery(timeFieldName string) (query bson.M) {
	query = bson.M{}

	if c.Id != "" {
		query["_id"] = c.Id.toBsonObjectId()
	}

	if c.Uid != 0 {
		query["uid"] = c.Uid
	}

	if c.Email != "" {
		query["email"] = c.Email
	}

	if !c.StartTime.IsZero() {
		query[timeFieldName] = bson.M{
			"$gte": c.StartTime,
		}
	}

	if !c.EndTime.IsZero() {
		b := query[timeFieldName]
		timeQuery := bson.M{}

		if b != nil {
			timeQuery = b.(bson.M)
		}

		timeQuery["$lte"] = c.EndTime
		query[timeFieldName] = timeQuery
	}

	return
}
