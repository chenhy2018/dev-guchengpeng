package memcache

import (
	"reflect"
	"testing"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v1/require"
)

type mockMemcacheEx struct {
	caches map[string]interface{}
	get    int
	set    int
	del    int
}

func (mc *mockMemcacheEx) Set(xl *xlog.Logger, key string, val interface{}) error {

	return mc.SetWithExpires(xl, key, val, 0)
}

func (mc *mockMemcacheEx) SetWithExpires(xl *xlog.Logger, key string, val interface{}, expires int32) error {

	mc.set++
	mc.caches[key] = val
	return nil
}

func (mc *mockMemcacheEx) Get(xl *xlog.Logger, key string, ret interface{}) error {

	mc.get++
	val, ok := mc.caches[key]
	if !ok {
		return ErrCacheMiss
	}
	reflect.ValueOf(ret).Elem().Set(reflect.ValueOf(val).Elem())
	return nil
}

func (mc *mockMemcacheEx) Del(xl *xlog.Logger, key string) error {

	mc.del++
	delete(mc.caches, key)
	return nil
}

func TestMixed_V2(t *testing.T) {

	old := &mockMemcacheEx{caches: make(map[string]interface{})}
	new := &mockMemcacheEx{caches: make(map[string]interface{})}

	mix := NewMixedEx(new, old)

	xl := xlog.NewDummy()
	val := 10

	var ret int
	err := mix.Get(xl, "foo", &ret)
	require.Equal(t, ErrCacheMiss, err)
	require.Equal(t, 1, old.get)
	require.Equal(t, 1, new.get)

	old.Set(xl, "foo", &val)
	err = mix.Get(xl, "foo", &ret)
	require.Equal(t, 10, ret)
	require.Equal(t, 2, old.get)
	require.Equal(t, 2, new.get)

	mix.Set(xl, "foo_new", &val)
	err = old.Get(xl, "foo_new", &ret)
	require.Equal(t, ErrCacheMiss, err)
	err = new.Get(xl, "foo_new", &ret)
	require.Equal(t, 10, ret)

	mix.Del(xl, "foo")
	err = old.Get(xl, "foo", &ret)
	require.Equal(t, ErrCacheMiss, err)
	err = new.Get(xl, "foo", &ret)
	require.Equal(t, ErrCacheMiss, err)
}
