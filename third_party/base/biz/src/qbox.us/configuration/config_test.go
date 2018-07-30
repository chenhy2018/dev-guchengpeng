package configuration

import (
	"fmt"
	"testing"

	"labix.org/v2/mgo"

	"github.com/stretchr/testify.v1/require"
)

func TestConfig(t *testing.T) {
	session, err := mgo.Dial("localhost:27016")
	require.NoError(t, err, "dial")
	coll := session.DB("test").C("test_configuration_pkg")
	coll.DropCollection()
	instance, err := New(coll, nil)
	require.NoError(t, err, "new instance")

	grp := "test_grp"
	all := map[string]string{}
	// put and get
	for i := 0; i < 100; i++ {
		key := fmt.Sprint("key", i)
		val := fmt.Sprint("val", i)
		all[key] = val
		err = instance.Put(grp, key, val)
		require.NoError(t, err, "put")
		v, err := instance.Get(grp, key)
		require.NoError(t, err, "get")
		require.Equal(t, val, v, "check get")
	}
	// put same again and get
	for i := 0; i < 100; i++ {
		key := fmt.Sprint("key", i)
		val := fmt.Sprint("val", i)
		all[key] = val
		err = instance.Put(grp, key, val)
		require.NoError(t, err, "put")
		v, err := instance.Get(grp, key)
		require.NoError(t, err, "get")
		require.Equal(t, val, v, "check get")
	}
	// put different value again and get
	for i := 0; i < 100; i++ {
		key := fmt.Sprint("key", i)
		val := fmt.Sprint("val", i+1000)
		all[key] = val
		err = instance.Put(grp, key, val)
		require.NoError(t, err, "put")
		v, err := instance.Get(grp, key)
		require.NoError(t, err, "get")
		require.Equal(t, val, v, "check get")
	}
	// group
	items, err := instance.Group(grp)
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
		err = instance.Delete(grp, key)
		require.NoError(t, err, "del")
		v, err := instance.Get(grp, key)
		require.Equal(t, mgo.ErrNotFound, err, "get deleted one")
		require.Equal(t, "", v, "value should be none")
	}
	// group again
	items, err = instance.Group(grp)
	require.NoError(t, err, "group")
	require.Len(t, items, 0, "no item")
}
