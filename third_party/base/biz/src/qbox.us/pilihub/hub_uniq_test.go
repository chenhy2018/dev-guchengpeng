package pilihub

import (
	"testing"

	"github.com/qiniu/db/mgoutil.v3"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"time"
)

func dropDB() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic("dial to mongo failed.")
	}
	defer session.Close()
	session.DB("hubuniq").DropDatabase()
}

func newHubUniq(ver Version) HubUniq {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic("dial to mongo failed")
	}
	coll := mgoutil.Collection{session.DB("hubuniq").C("hub")}
	return New(coll, ver)
}

func TestCreate(t *testing.T) {
	dropDB()
	v1 := newHubUniq(V1)
	v2 := newHubUniq(V2)
	start := time.Now().Add(-time.Millisecond)

	hub1, err := v1.Create("hub1", 1)
	assert.NoError(t, err)
	assert.Equal(t, "hub1", hub1.HubName)
	assert.Equal(t, 1, hub1.Uid)
	assert.Equal(t, V1, hub1.Version)
	assert.True(t, start.Before(hub1.CreatedAt))

	_, err = v2.Create("hub1", 1)
	assert.Equal(t, ErrHubExist, err)

	_, err = v2.Create("hub1", 2)
	assert.Equal(t, ErrHubExist, err)

}

func TestDelete(t *testing.T) {
	dropDB()
	v1 := newHubUniq(V1)
	v2 := newHubUniq(V2)

	hub1, err := v1.Create("hub1", 1)
	assert.NoError(t, err)

	//Version不对无法删除
	err = v2.Delete(hub1.HubName, hub1.Uid)
	assert.Equal(t, ErrHubNotFound, err)

	//uid不对无法删除
	hub1.Uid = 2
	err = v1.Delete(hub1.HubName, hub1.Uid)
	assert.Equal(t, ErrHubNotFound, err)

	hub1.Uid = 1
	err = v1.Delete(hub1.HubName, hub1.Uid)
	assert.NoError(t, err)

	//又可以创建同名hub
	hub2, err := v2.Create("hub1", 1)
	assert.NoError(t, err)
	err = v2.Delete(hub2.HubName, hub2.Uid)
	assert.NoError(t, err)

}

func TestGet(t *testing.T) {
	dropDB()
	v1 := newHubUniq(V1)
	v2 := newHubUniq(V2)

	hub1, err := v1.Create("hub1", 1)
	assert.NoError(t, err)

	getHub1, err := v1.Get("hub1")
	assert.NoError(t, err)
	assert.Equal(t, hub1, getHub1)

	getHub2, err := v2.Get("hub1")
	assert.NoError(t, err)
	assert.Equal(t, hub1, getHub2)

	getHub1, err = v1.Get("orz")
	assert.Equal(t, ErrHubNotFound, err)

	getHub2, err = v2.Get("orz123")
	assert.Equal(t, ErrHubNotFound, err)
}
