package memcache

import (
	"github.com/qiniu/xlog.v1"
)

type mixedMemcacheEx struct {
	new MemcacheEx
	old MemcacheEx
}

func NewMixedEx(new, old MemcacheEx) MemcacheEx {

	return &mixedMemcacheEx{new, old}
}

func (mc *mixedMemcacheEx) Set(xl *xlog.Logger, key string, val interface{}) error {

	return mc.SetWithExpires(xl, key, val, 0)
}

func (mc *mixedMemcacheEx) SetWithExpires(xl *xlog.Logger, key string, val interface{}, expires int32) error {
	return mc.new.SetWithExpires(xl, key, val, expires)
}

func (mc *mixedMemcacheEx) Get(xl *xlog.Logger, key string, ret interface{}) error {
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

func (mc *mixedMemcacheEx) Del(xl *xlog.Logger, key string) error {

	newErr := mc.new.Del(xl, key)
	oldErr := mc.old.Del(xl, key)
	if newErr != nil {
		return newErr
	}
	return oldErr
}
