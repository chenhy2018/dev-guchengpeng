package unsafe

import (
	"reflect"
	"unsafe"
)

func StringPtr(s string) unsafe.Pointer {
	return unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data)
}

func BytesPtr(b []byte) unsafe.Pointer {
	return unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&b)).Data)
}
