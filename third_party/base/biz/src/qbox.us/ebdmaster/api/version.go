package api

import (
	"encoding/binary"
	"io"

	"qbox.us/errors"
)

var ErrVersion = errors.New("error version")
var initVersion = 201711201104

// ebdmaster的通信协议如果发生变化，需要在此处增加新的版本信息。版本号必须是递增的方式，默认格式为 year|month|day|hour|minute。
var versions = []int64{
	int64(initVersion), // ebd增加版本验证。所有从ebdmaster接收的信息需要验证当前的版本号。
	201801091129,       // 后继的版本号顺序添加。
	201801241133,       // ebdmaster、ecb增加crc64校验。
}

func init() {
	preVersion := int64(0)
	for _, v := range versions {
		if v < preVersion {
			panic(ErrVersion)
		}
		preVersion = v
	}
}

func CurrentVersion() int64 {
	return versions[len(versions)-1]
}

func CheckVersion(r io.Reader) (err error, version int64) {
	if err = binary.Read(r, binary.LittleEndian, &version); err != nil {
		return
	}
	if version != CurrentVersion() {
		err = ErrVersion
	}
	return
}
