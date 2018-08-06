/* Added by Lingtao Kong in 2015/07/28
To test config_NoCache.go which realize the function of removing uc's cache
*/
package configuration

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v1/require"
	"github.com/stretchr/testify/assert"
	"labix.org/v2/mgo"
	"qbox.us/mgo2"
)

var (
	oldCollConfig = &mgo2.Config{
		Host: "localhost:27017",
		DB:   "qbox_uc_configuration",
		Coll: "uc_configuration",
	}
	newCollConfig = &mgo2.Config{
		Host: "localhost:27017",
		DB:   "qbox_bucket_configuration",
		Coll: "bucket_configuration",
	}
)

func TestConfigNC(t *testing.T) {
	session, err := mgo.Dial("localhost:27016")
	require.NoError(t, err, "dial")
	coll := session.DB("test").C("test_configuration_pkg")
	coll.DropCollection()
	instance, err := OldNC(coll, nil)
	require.NoError(t, err, "new instance")
	fmt.Println("Test config no cache well!")
	xl := xlog.NewDummy()

	grp := "test_grp"
	all := map[string]string{}
	// put and get
	for i := 0; i < 100; i++ {
		key := fmt.Sprint("key", i)
		val := fmt.Sprint("val", i)
		all[key] = val
		err = instance.Put(xl, grp, key, val)
		require.NoError(t, err, "put")
		v, err := instance.Get(xl, grp, key)
		require.NoError(t, err, "get")
		require.Equal(t, val, v, "check get")
	}
	// put same again and get
	for i := 0; i < 100; i++ {
		key := fmt.Sprint("key", i)
		val := fmt.Sprint("val", i)
		all[key] = val
		err = instance.Put(xl, grp, key, val)
		require.NoError(t, err, "put")
		v, err := instance.Get(xl, grp, key)
		require.NoError(t, err, "get")
		require.Equal(t, val, v, "check get")
	}
	// put different value again and get
	for i := 0; i < 100; i++ {
		key := fmt.Sprint("key", i)
		val := fmt.Sprint("val", i+1000)
		all[key] = val
		err = instance.Put(xl, grp, key, val)
		require.NoError(t, err, "put")
		v, err := instance.Get(xl, grp, key)
		require.NoError(t, err, "get")
		require.Equal(t, val, v, "check get")
	}
	// group
	items, err := instance.Group(xl, grp)
	require.NoError(t, err, "group")
	for _, item := range items {
		require.Equal(t, all[item.Key], item.Val, "check")
		delete(all, item.Key)
	}
	require.Len(t, items, 100, "items")
	require.Len(t, all, 0, "all")
	// delete
	for i := 0; i < 100; i++ {
		key := fmt.Sprint("key", i)
		err = instance.Delete(xl, grp, key)
		require.NoError(t, err, "del")
		v, err := instance.Get(xl, grp, key)
		require.Equal(t, mgo.ErrNotFound, err, "get deleted one")
		require.Equal(t, "", v, "value should be none")
	}
	// group again
	items, err = instance.Group(xl, grp)
	require.NoError(t, err, "group")
	require.Len(t, items, 0, "no item")
}

// get entry that has fetched
func TestOldInstanceNC2(t *testing.T) {
	testGrp := "test"
	testKey := "test"

	oldSession := mgo2.Open(oldCollConfig)
	defer oldSession.Close()
	oldColl := oldSession.Coll
	oldColl.DropCollection()

	oldNC, err := OldNC(oldColl, nil)
	assert.NoError(t, err)
	err = oldColl.Insert(M{"grp": testGrp, "key": testKey, "fetched": true})
	assert.NoError(t, err)
	val, err := oldNC.Get(xlog.NewDummy(), testGrp, testKey)
	assert.Equal(t, mgo.ErrNotFound, err)
	assert.Equal(t, "", val)
}

func prepareNewCollData(t *testing.T) {
	newSession := mgo2.Open(newCollConfig)
	defer newSession.Close()
	newColl := newSession.Coll
	newColl.DropCollection()
	index := mgo.Index{Key: []string{"uid", "tbl"}, Unique: true}
	err := newColl.EnsureIndex(index)
	assert.NoError(t, err)

	data := []struct {
		uid  uint32
		tbl  string
		val  string
		drop int64
	}{
		{1, "normal", "val", 0},
		{1, "not_set", "", 0},
		{1, "dropped", "", 1},
	}
	for i, d := range data {
		msg := strconv.Itoa(i)
		m := M{"uid": d.uid, "tbl": d.tbl, "drop": d.drop}
		if d.val != "" {
			m["val"] = d.val
		}
		err = newColl.Insert(m)
		assert.NoError(t, err, msg)
	}
}

func TestNewInstanceNC(t *testing.T) {
	prepareNewCollData(t)
	newSession := mgo2.Open(newCollConfig)
	defer newSession.Close()
	newColl := newSession.Coll
	newNC, _ := NewNC(newColl, nil)

	tests := []struct {
		uid uint32
		tbl string
		err error
		val string
	}{
		{1, "normal", nil, "val"},
		{1, "not_set", nil, ""},
		{1, "dropped", mgo.ErrNotFound, ""},
		{1, "inexistent", mgo.ErrNotFound, ""},
	}
	for i, test := range tests {
		msg := strconv.Itoa(i)
		val, err := newNC.Get(xlog.NewDummy(), test.uid, test.tbl)
		assert.Equal(t, test.err, err, msg)
		assert.Equal(t, test.val, val, msg)
	}

	items, err := newNC.Group(xlog.NewDummy())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(items))
	assert.Equal(t, "1:normal", items[0].Key)
	assert.Equal(t, "val", items[0].Val)

	err = newNC.Delete(xlog.NewDummy(), 1, "normal")
	assert.NoError(t, err)
	err = newNC.Delete(xlog.NewDummy(), 2, "normal")
	assert.Equal(t, mgo.ErrNotFound, err)

	err = newNC.Put(xlog.NewDummy(), 1, "normal", "val2")
	assert.NoError(t, err)
	err = newNC.Put(xlog.NewDummy(), 2, "normal", "val2")
	assert.Equal(t, mgo.ErrNotFound, err)
}
