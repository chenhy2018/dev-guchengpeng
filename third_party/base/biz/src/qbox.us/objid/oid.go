package objid

import (
	"encoding/base64"
	"encoding/binary"
	"syscall"
)

// -------------------------------------------------------------------------

func Encode(low uint32, hi uint64) string {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:4], low)
	binary.LittleEndian.PutUint64(b[4:], hi)
	return base64.URLEncoding.EncodeToString(b[:])
}

func Decode(objid string) (low uint32, hi uint64, err error) {
	b, err := base64.URLEncoding.DecodeString(objid)
	if err != nil {
		return
	}
	if len(b) != 12 {
		err = syscall.EINVAL
		return
	}
	low = binary.LittleEndian.Uint32(b)
	hi = binary.LittleEndian.Uint64(b[4:])
	return
}

// -------------------------------------------------------------------------
