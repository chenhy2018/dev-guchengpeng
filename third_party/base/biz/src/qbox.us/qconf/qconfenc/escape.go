package qconfenc

import (
	"encoding/base64"
	"reflect"
	"syscall"
)

// ------------------------------------------------------------------------

func Escape(v string) string {

	venc := base64.URLEncoding.EncodeToString([]byte(v))
	switch len(v) % 3 {
	case 1:
		return venc[:len(venc)-2]
	case 2:
		return venc[:len(venc)-1]
	}
	return venc
}

func Unescape(venc string) (v string, err error) {

	switch len(venc) % 4 {
	case 2:
		venc += "=="
	case 3:
		venc += "="
	}
	b, err := base64.URLEncoding.DecodeString(venc)
	if err == nil {
		v = string(b)
	}
	return
}

// ------------------------------------------------------------------------

func UnescapeMapKeys(v1 interface{}) error {

	v2 := reflect.ValueOf(v1)
	if v2.Kind() != reflect.Ptr {
		return syscall.EINVAL
	}
	v := v2.Elem()
	if v.Kind() != reflect.Map {
		return syscall.EINVAL
	}
	ret := reflect.MakeMap(v.Type())
	keyencs := v.MapKeys()
	for _, keyenc := range keyencs {
		if keyenc.Kind() != reflect.String {
			return syscall.EINVAL
		}
		key, err := Unescape(keyenc.String())
		if err != nil {
			return err
		}
		ret.SetMapIndex(reflect.ValueOf(key), v.MapIndex(keyenc))
	}
	v.Set(ret)
	return nil
}

// ------------------------------------------------------------------------
