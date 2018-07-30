package types

import (
	"encoding/base64"
	"encoding/binary"
	"time"
)

var (
	InvalidGid Gid = [12]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

type Gid [12]byte

func NewGid(dgid uint32) (gid Gid) {
	binary.LittleEndian.PutUint32(gid[:4], dgid)
	now := time.Now().UnixNano()
	binary.LittleEndian.PutUint64(gid[4:12], uint64(now))
	return
}

func (self Gid) UnixTime() int64 {
	return int64(binary.LittleEndian.Uint64(self[4:12]))
}

func (self Gid) MotherDgid() uint32 {
	return binary.LittleEndian.Uint32(self[:4])
}

func (self Gid) String() string {
	return EncodeGid(self)
}

var encodeLen = base64.URLEncoding.EncodedLen(12)

func (self Gid) MarshalJSON() ([]byte, error) {
	b := make([]byte, encodeLen+2)
	b[0], b[encodeLen+1] = '"', '"'
	base64.URLEncoding.Encode(b[1:], self[:])
	return b, nil
}

func (self *Gid) UnmarshalJSON(b []byte) error {
	_, err := base64.URLEncoding.Decode(self[:], b[1:encodeLen+1])
	return err
}

func EncodeGid(gid Gid) string {
	return base64.URLEncoding.EncodeToString(gid[:])
}

func DecodeGid(egid string) (gid Gid, err error) {
	b, err := base64.URLEncoding.DecodeString(egid)
	copy(gid[:], b)
	return
}
