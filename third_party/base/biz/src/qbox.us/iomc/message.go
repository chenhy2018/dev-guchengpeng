package iomc

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
)

func Encode(m *Message) (data []byte, err error) {
	return json.Marshal(m)
}

func Decode(data []byte) (m *Message, err error) {
	msg := Message{}
	err = json.Unmarshal(data, &msg)
	if err != nil {
		return
	}
	return &msg, err
}

func (m *Message) BuildIOMemcacheKey() string {
	rawKey := "io_v2:" + strconv.FormatUint(uint64(m.Itbl), 36) + ":" + m.Key
	return base64.URLEncoding.EncodeToString([]byte(rawKey))
}
