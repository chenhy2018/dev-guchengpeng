package portalio

import (
	"testing"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/stretchr/testify/assert"
)

func TestPositiveQuery(t *testing.T) {
	query := Query{
		Id:        QueryId("53a8eddd0b12e047ff000001"),
		Uid:       uint32(1369076594),
		Email:     "duxiaofeng@qiniu.com",
		StartTime: time.Date(2016, 4, 8, 16, 42, 0, 0, time.Local),
		EndTime:   time.Date(2016, 4, 9, 16, 42, 0, 0, time.Local),
		Skip:      10,
		Limit:     10,
		Sort: []string{
			"-create_at",
		},
	}

	actualBsonQuery := query.GetBsonQuery("create_at")

	expectedValid := true
	expectedSkip := 10
	expectedLimit := 10
	expectedTimeField := "create_at"
	expectedBsonQuery := bson.M{
		"_id":   bson.ObjectIdHex("53a8eddd0b12e047ff000001"),
		"uid":   uint32(1369076594),
		"email": "duxiaofeng@qiniu.com",
		expectedTimeField: bson.M{
			"$gte": time.Date(2016, 4, 8, 16, 42, 0, 0, time.Local),
			"$lte": time.Date(2016, 4, 9, 16, 42, 0, 0, time.Local),
		},
	}

	assert.Equal(t, expectedValid, query.Valid(), "valid testing")
	assert.Equal(t, expectedSkip, query.GetSkip(), "valid skip")
	assert.Equal(t, expectedLimit, query.GetLimit(), "valid limit")
	assert.Equal(t, expectedBsonQuery["_id"].(bson.ObjectId).Hex(), actualBsonQuery["_id"].(bson.ObjectId).Hex(), "valid id")
	assert.Equal(t, expectedBsonQuery["uid"].(uint32), actualBsonQuery["uid"].(uint32), "valid uid")
	assert.Equal(t, expectedBsonQuery["email"].(string), actualBsonQuery["email"].(string), "valid email")
	assert.Equal(t, expectedBsonQuery[expectedTimeField].(bson.M)["$gte"].(time.Time).Unix(), actualBsonQuery[expectedTimeField].(bson.M)["$gte"].(time.Time).Unix(), "valid start time")
	assert.Equal(t, expectedBsonQuery[expectedTimeField].(bson.M)["$lte"].(time.Time).Unix(), actualBsonQuery[expectedTimeField].(bson.M)["$lte"].(time.Time).Unix(), "valid end time")
}
