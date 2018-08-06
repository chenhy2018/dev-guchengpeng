package memcache

import "github.com/qiniu/xlog.v1"

type mixedMemcache struct {
	new Memcache
	old Memcache
}

func NewMixed(new, old Memcache) Memcache {

	return &mixedMemcache{new, old}
}

func (mc *mixedMemcache) Set(xl *xlog.Logger, key string, val interface{}) error {

	return mc.new.Set(xl, key, val)
}

func (mc *mixedMemcache) Get(xl *xlog.Logger, key string, ret interface{}) error {

	err := mc.new.Get(xl, key, ret)
	if err == nil {
		return nil
	}
	err = mc.old.Get(xl, key, ret)
	if err == nil {
		mc.new.Set(xl, key, ret)
	}
	return err
}

func (mc *mixedMemcache) Del(xl *xlog.Logger, key string) error {

	newErr := mc.new.Del(xl, key)
	oldErr := mc.old.Del(xl, key)
	if newErr != nil {
		return newErr
	}
	return oldErr
}
