package reqlogger

import (
	"encoding/base64"
	"encoding/binary"
	"math/rand"
	"time"
)

var rnd = uint32(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(999999999))

func NewReqId() string {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], rnd)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}

func DecodeReqId(reqId string) (uint, int64) {
	b, err := base64.URLEncoding.DecodeString(reqId)
	if err != nil || len(b) < 12 {
		return 0, 0
	}
	rnd := binary.LittleEndian.Uint32(b[:4])
	unixNano := binary.LittleEndian.Uint64(b[4:])
	return uint(rnd), int64(unixNano)
}
