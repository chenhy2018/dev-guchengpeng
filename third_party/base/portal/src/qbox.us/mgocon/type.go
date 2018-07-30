package mgocon

import (
	"time"

	"labix.org/v2/mgo/bson"
)

type UnixNano bson.MongoTimestamp

func (t UnixNano) Time() time.Time {
	return time.Unix(0, int64(t))
}
