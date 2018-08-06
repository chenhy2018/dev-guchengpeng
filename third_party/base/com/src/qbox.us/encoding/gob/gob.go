package gob

import (
	"bytes"
	"encoding/gob"
	"qbox.us/cc"
	"reflect"
)

func Marshal(v interface{}) ([]byte, error) {

	b := bytes.NewBuffer(nil)
	err := gob.NewEncoder(b).Encode(v)
	return b.Bytes(), err
}

func MarshalValue(v reflect.Value) ([]byte, error) {

	b := bytes.NewBuffer(nil)
	err := gob.NewEncoder(b).EncodeValue(v)
	return b.Bytes(), err
}

func Unmarshal(data []byte, v interface{}) error {

	b := cc.NewBytesReader(data)
	return gob.NewDecoder(b).Decode(v)
}

func UnmarshalValue(data []byte, v reflect.Value) error {

	b := cc.NewBytesReader(data)
	return gob.NewDecoder(b).DecodeValue(v)
}
