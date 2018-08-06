package mgo2

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

func TestMongoIndexService_EnsureIndex(t *testing.T) {
	session, err := mgo.Dial("localhost")
	if !assert.NoError(t, err) {
		return
	}
	defer session.Close()

	collection := session.DB("biz_test").C("test1")
	_ = collection.DropCollection()
	service := NewMongoIndexService(
		session, "biz_test")
	service.EnsureIndex(bson.M{
		"name": "test1",
		"unique": []string{
			"uid",
			"email",
		},
	})
	indexes, err := collection.Indexes()
	assert.NoError(t, err)
	assert.Equal(t, len(indexes), 3)
}
