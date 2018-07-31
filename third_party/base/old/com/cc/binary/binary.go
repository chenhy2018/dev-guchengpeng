package binary

import (
	"bytes"
	"encoding/binary"
	"os"
	"qbox.us/cc"
)

func ToBinary(data interface{}) []byte {
	b := bytes.NewBuffer(nil)
	err := binary.Write(b, binary.LittleEndian, data)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func FromBinary(b []byte, data interface{}) os.Error {
	r := cc.NewBytesReader(b)
	return binary.Read(r, binary.LittleEndian, data)
}
