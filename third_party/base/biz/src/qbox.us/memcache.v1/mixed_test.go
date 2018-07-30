package memcache

import (
	"reflect"
	"testing"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v1/require"
)

type mockMemcache struct {
	caches map[string]interface{}
	get    int
	set    int
	del    int
}

func (mc *mockMemcache) Set(xl *xlog.Logger, key string, val interface{}) error {

	mc.set++
	mc.caches[key] = val
	return nil
}

func (mc *mockMemcache) Get(xl *xlog.Logger, key string, ret interface{}) error {

	mc.get++
	val, ok := mc.caches[key]
	if !ok {
		return ErrCacheMiss
	}
	reflect.ValueOf(ret).Elem().Set(reflect.ValueOf(val).Elem())
	return nil
}

func (mc *mockMemcache) Del(xl *xlog.Logger, key string) error {

	mc.del++
	delete(mc.caches, key)
	return nil
}

func TestMixed(t *testing.T) {

	old := &mockMemcache{caches: make(map[string]interface{})}
	new := &mockMemcache{caches: make(map[string]interface{})}

	mix := NewMixed(new, old)

	xl := xlog.NewDummy()
	val := 10

	var ret int
	err := mix.Get(xl, "foo", &ret)
	require.Equal(t, ErrCacheMiss, err)
	require.Equal(t, 1, old.get)
	require.Equal(t, 1, new.get)

	old.Set(xl, "foo", &val)
	require.Equal(t, 1, old.set)

	err = mix.Get(xl, "foo", &ret)
	require.NoError(t, err)
	require.Equal(t, 2, old.get)
	require.Equal(t, 1, old.set)
	require.Equal(t, 2, new.get)
	require.Equal(t, 1, new.set)
	require.Equal(t, 10, ret)

	err = mix.Get(xl, "foo", &ret)
	require.NoError(t, err)
	require.Equal(t, 2, old.get)
	require.Equal(t, 1, old.set)
	require.Equal(t, 3, new.get)
	require.Equal(t, 1, new.set)
	require.Equal(t, 10, ret)

	err = mix.Del(xl, "foo")
	require.NoError(t, err)
	require.Equal(t, 2, old.get)
	require.Equal(t, 1, old.set)
	require.Equal(t, 1, old.del)
	require.Equal(t, 3, new.get)
	require.Equal(t, 1, new.set)
	require.Equal(t, 1, new.del)

	err = mix.Get(xl, "foo", &ret)
	require.Equal(t, ErrCacheMiss, err)
	require.Equal(t, 3, old.get)
	require.Equal(t, 1, old.set)
	require.Equal(t, 1, old.del)
	require.Equal(t, 4, new.get)
	require.Equal(t, 1, new.set)
	require.Equal(t, 1, new.del)
}
