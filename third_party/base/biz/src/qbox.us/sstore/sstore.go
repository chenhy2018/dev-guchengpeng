package sstore

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math"

	"github.com/qiniu/log.v1"
	"qbox.us/cc/time"
	"qbox.us/objid"
)

// --------------------------------------------------------------------

type KeyFinder interface {
	Find(KeyHint uint32) []byte
}

// --------------------------------------------------------------------

type FhandleInfo struct {
	Fhandle  []byte // 330
	MimeType string // 32
	AttName  string // 32

	Fsize    int64 // 8
	Deadline int64 // 8

	KeyHint uint32 // 4
	OidLow  uint32 // 4
	OidHigh uint64 // 8

	Ver         uint16 // 2
	Compression uint8  // 1
	ChMask      uint8  // 1
	Uid         uint32 // 4

	Utype  uint32 // 4
	Public uint8  // 1

	PutTime int64 // 8 in nanosecond

	FileType uint32 // 4
}

var LenPutTime, LenFileType int

func init() {
	LenPutTime = binary.Size(FhandleInfo{}.PutTime)
	LenFileType = binary.Size(FhandleInfo{}.FileType)
}

/*
	Token []byte		// 20

	Ver int8			// 1
	Reserved int8		// 1
	Compression int8	// 1
	ChMask int8			// 1

	ChunkCnt int8		// 1
	MimeTypeLen int8	// 1
	AttNameLen int8		// 1
	Public int8			// 1

	KeyHint int32		// 4

	Fsize int64  		// 8
	Deadline int64		// 8

	Uid int32			// 4
	OidLow	int32		// 4
	OidHigh	int64		// 8

	Fhandle []byte
	MimeType string
	AttName string

	Utype int32			// 4
*/

/*
	Ver 0x13: https://pm.qbox.me/issues/20683
	AttNameLen int16,  key limit is 750 bytes
*/

const FhandleVer_10 = 0x10 // 1.0
const FhandleVer_11 = 0x11 // 1.1
const FhandleVer_12 = 0x12 // 1.2
const FhandleVer_13 = 0x13 // 1.3
const FhandleVer_14 = 0x14 // 1.4

const FhandleVer = FhandleVer_14

func EncodeFhandle(fi *FhandleInfo, key []byte) (efh string) {
	return EncodeFhandle_14(fi, key)
}

func EncodeFhandle_10(fi *FhandleInfo, key []byte) (efh string) {

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	if fi.ChMask == 0 {
		fi.ChMask = uint8(fi.KeyHint ^ dwMask)
	}

	totalLen := 64 + len(fi.Fhandle) + len(fi.MimeType) + len(fi.AttName)
	chunkCnt := len(fi.Fhandle) / 20
	b := make([]byte, totalLen)

	b[20] = FhandleVer_10
	b[21] = 0x34
	b[22] = fi.Compression
	b[23] = fi.ChMask

	b[24] = byte(chunkCnt)
	b[25] = byte(len(fi.MimeType))
	b[26] = byte(len(fi.AttName))
	b[27] = 0x27

	binary.LittleEndian.PutUint32(b[28:], fi.KeyHint^dwMask)
	binary.LittleEndian.PutUint64(b[32:], uint64(fi.Fsize^dwMask2))
	binary.LittleEndian.PutUint64(b[40:], uint64(fi.Deadline))
	binary.LittleEndian.PutUint32(b[48:], fi.Uid^dwMask)
	binary.LittleEndian.PutUint32(b[52:], fi.OidLow^dwMask)
	binary.LittleEndian.PutUint64(b[56:], fi.OidHigh^uint64(dwMask2))

	off := 64 + copy(b[64:], fi.Fhandle)
	off += copy(b[off:], fi.MimeType)
	off += copy(b[off:], fi.AttName)

	for i := 64; i < off; i++ {
		b[i] ^= fi.ChMask
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	copy(b[:20], h.Sum(nil))

	log.Debug("Fhandle:\n" + hex.Dump(b))

	return base64.URLEncoding.EncodeToString(b)
}

func DecodeFhandle_10(efh string, oid string, kf KeyFinder) *FhandleInfo {

	b, err := base64.URLEncoding.DecodeString(efh)
	if err != nil || len(b) < 64 {
		log.Println("DecodeString failed:", err)
		return nil
	}

	fi := &FhandleInfo{}
	fi.Deadline = int64(binary.LittleEndian.Uint64(b[40:]))
	if fi.Deadline < time.Nanoseconds() {
		log.Println("Deadline expired")
		return nil
	}

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	fi.OidHigh = binary.LittleEndian.Uint64(b[56:]) ^ uint64(dwMask2)
	fi.OidLow = binary.LittleEndian.Uint32(b[52:]) ^ dwMask

	if fi.OidHigh != 0 && fi.OidLow != 0 {
		oidLow, oidHigh, err := objid.Decode(oid)
		if err != nil {
			log.Println("DecodeFhandle: objid.Decode failed")
			return nil
		}
		if fi.OidHigh != oidHigh || fi.OidLow != oidLow {
			log.Println("DecodeFhandle: SessionID not matched")
			return nil
		}
	}

	fi.KeyHint = binary.LittleEndian.Uint32(b[28:]) ^ dwMask
	key := kf.Find(fi.KeyHint)
	if key == nil {
		log.Println("KeyFinder: key not found")
		return nil
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	if !bytes.Equal(h.Sum(nil), b[:20]) {
		log.Println("Digest verify failed")
		return nil
	}

	fi.Ver = uint16(b[20])
	if fi.Ver != FhandleVer_10 {
		log.Println("FhandleVer not match")
		return nil
	}

	fi.Fsize = int64(binary.LittleEndian.Uint64(b[32:])) ^ dwMask2
	fi.Uid = binary.LittleEndian.Uint32(b[48:]) ^ dwMask

	fi.Compression = b[22]
	fi.ChMask = b[23]

	for i := 64; i < len(b); i++ {
		b[i] ^= fi.ChMask
	}

	off := len(b) - int(b[25]) - int(b[26])
	//	off := 64 + int(b[24])*20 + 1
	fi.Fhandle = b[64:off]

	off2 := off + int(b[25])
	fi.MimeType = string(b[off:off2])

	off = off2
	off2 = off + int(b[26])
	fi.AttName = string(b[off:off2])

	return fi
}

func EncodeFhandle_11(fi *FhandleInfo, key []byte) (efh string) {

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	if fi.ChMask == 0 {
		fi.ChMask = uint8(fi.KeyHint ^ dwMask)
	}

	totalLen := 64 + len(fi.Fhandle) + len(fi.MimeType) + len(fi.AttName)
	if fi.Utype != 0 {
		totalLen += 4
	}
	chunkCnt := len(fi.Fhandle) / 20
	b := make([]byte, totalLen)

	if fi.Utype != 0 {
		b[20] = FhandleVer_11
	} else {
		b[20] = FhandleVer_10
	}
	b[21] = 0x34
	b[22] = fi.Compression
	b[23] = fi.ChMask

	b[24] = byte(chunkCnt)
	b[25] = byte(len(fi.MimeType))
	b[26] = byte(len(fi.AttName))
	b[27] = 0x27 ^ fi.Public

	binary.LittleEndian.PutUint32(b[28:], fi.KeyHint^dwMask)
	binary.LittleEndian.PutUint64(b[32:], uint64(fi.Fsize^dwMask2))
	binary.LittleEndian.PutUint64(b[40:], uint64(fi.Deadline))
	binary.LittleEndian.PutUint32(b[48:], fi.Uid^dwMask)
	binary.LittleEndian.PutUint32(b[52:], fi.OidLow^dwMask)
	binary.LittleEndian.PutUint64(b[56:], fi.OidHigh^uint64(dwMask2))

	off := 64 + copy(b[64:], fi.Fhandle)
	off += copy(b[off:], fi.MimeType)
	off += copy(b[off:], fi.AttName)

	if fi.Utype != 0 {
		binary.LittleEndian.PutUint32(b[off:], fi.Utype^dwMask)
		off += 4
	}

	for i := 64; i < off; i++ {
		b[i] ^= fi.ChMask
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	copy(b[:20], h.Sum(nil))

	log.Debug("Fhandle:\n" + hex.Dump(b))

	return base64.URLEncoding.EncodeToString(b)
}

func EncodeFhandle_12(fi *FhandleInfo, key []byte) (efh string) {

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	if fi.ChMask == 0 {
		fi.ChMask = uint8(fi.KeyHint ^ dwMask)
	}

	totalLen := 64 + len(fi.Fhandle) + len(fi.MimeType) + len(fi.AttName) + LenPutTime
	totalLen += 4 // utype
	chunkCnt := len(fi.Fhandle) / 20
	b := make([]byte, totalLen)

	b[20] = FhandleVer_12
	b[21] = 0x34
	b[22] = fi.Compression
	b[23] = fi.ChMask

	b[24] = byte(chunkCnt)
	b[25] = byte(len(fi.MimeType))
	b[26] = byte(len(fi.AttName))
	b[27] = 0x27 ^ fi.Public

	binary.LittleEndian.PutUint32(b[28:], fi.KeyHint^dwMask)
	binary.LittleEndian.PutUint64(b[32:], uint64(fi.Fsize^dwMask2))
	binary.LittleEndian.PutUint64(b[40:], uint64(fi.Deadline))
	binary.LittleEndian.PutUint32(b[48:], fi.Uid^dwMask)
	binary.LittleEndian.PutUint32(b[52:], fi.OidLow^dwMask)
	binary.LittleEndian.PutUint64(b[56:], fi.OidHigh^uint64(dwMask2))

	off := 64 + copy(b[64:], fi.Fhandle)
	off += copy(b[off:], fi.MimeType)
	off += copy(b[off:], fi.AttName)

	binary.LittleEndian.PutUint32(b[off:], fi.Utype^dwMask)
	off += 4
	binary.LittleEndian.PutUint64(b[off:], uint64(fi.PutTime^dwMask2))
	off += LenPutTime

	for i := 64; i < off; i++ {
		b[i] ^= fi.ChMask
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	copy(b[:20], h.Sum(nil))

	log.Debug("Fhandle:\n" + hex.Dump(b))

	return base64.URLEncoding.EncodeToString(b)
}

func EncodeFhandle_13(fi *FhandleInfo, key []byte) (efh string) {

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	if fi.ChMask == 0 {
		fi.ChMask = uint8(fi.KeyHint ^ dwMask)
	}

	totalLen := 65 + len(fi.Fhandle) + len(fi.MimeType) + len(fi.AttName) + LenPutTime
	totalLen += 4 // utype
	chunkCnt := len(fi.Fhandle) / 20
	b := make([]byte, totalLen)

	b[20] = FhandleVer_13
	b[21] = 0x34
	b[22] = fi.Compression
	b[23] = fi.ChMask

	b[24] = byte(chunkCnt)

	// max len 200, defined in com/src/qbox.us/cc/mime/mime.go
	if len(fi.MimeType) > math.MaxUint8 {
		panic("MimeTypeLen overflow")
	}
	b[25] = byte(len(fi.MimeType))

	// max len is 750, defined in rs
	if len(fi.AttName) > math.MaxUint16 {
		panic("AttNameLen overflow")
	}
	binary.LittleEndian.PutUint16(b[26:], uint16(len(fi.AttName)))

	b[28] = 0x27 ^ fi.Public

	binary.LittleEndian.PutUint32(b[29:], fi.KeyHint^dwMask)
	binary.LittleEndian.PutUint64(b[33:], uint64(fi.Fsize^dwMask2))
	binary.LittleEndian.PutUint64(b[41:], uint64(fi.Deadline))
	binary.LittleEndian.PutUint32(b[49:], fi.Uid^dwMask)
	binary.LittleEndian.PutUint32(b[53:], fi.OidLow^dwMask)
	binary.LittleEndian.PutUint64(b[57:], fi.OidHigh^uint64(dwMask2))

	off := 65 + copy(b[65:], fi.Fhandle)
	off += copy(b[off:], fi.MimeType)
	off += copy(b[off:], fi.AttName)

	binary.LittleEndian.PutUint32(b[off:], fi.Utype^dwMask)
	off += 4
	binary.LittleEndian.PutUint64(b[off:], uint64(fi.PutTime^dwMask2))
	off += LenPutTime

	for i := 65; i < off; i++ {
		b[i] ^= fi.ChMask
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	copy(b[:20], h.Sum(nil))

	log.Debug("Fhandle:\n" + hex.Dump(b))

	return base64.URLEncoding.EncodeToString(b)
}

func EncodeFhandle_14(fi *FhandleInfo, key []byte) (efh string) {

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	if fi.ChMask == 0 {
		fi.ChMask = uint8(fi.KeyHint ^ dwMask)
	}

	totalLen := 65 + len(fi.Fhandle) + len(fi.MimeType) + len(fi.AttName) + LenPutTime + LenFileType
	totalLen += 4 // utype
	chunkCnt := len(fi.Fhandle) / 20
	b := make([]byte, totalLen)

	b[20] = FhandleVer_14
	b[21] = 0x34
	b[22] = fi.Compression
	b[23] = fi.ChMask

	b[24] = byte(chunkCnt)

	// max len 200, defined in com/src/qbox.us/cc/mime/mime.go
	if len(fi.MimeType) > math.MaxUint8 {
		panic("MimeTypeLen overflow")
	}
	b[25] = byte(len(fi.MimeType))

	// max len is 750, defined in rs
	if len(fi.AttName) > math.MaxUint16 {
		panic("AttNameLen overflow")
	}
	binary.LittleEndian.PutUint16(b[26:], uint16(len(fi.AttName)))

	b[28] = 0x27 ^ fi.Public

	binary.LittleEndian.PutUint32(b[29:], fi.KeyHint^dwMask)
	binary.LittleEndian.PutUint64(b[33:], uint64(fi.Fsize^dwMask2))
	binary.LittleEndian.PutUint64(b[41:], uint64(fi.Deadline))
	binary.LittleEndian.PutUint32(b[49:], fi.Uid^dwMask)
	binary.LittleEndian.PutUint32(b[53:], fi.OidLow^dwMask)
	binary.LittleEndian.PutUint64(b[57:], fi.OidHigh^uint64(dwMask2))

	off := 65 + copy(b[65:], fi.Fhandle)
	off += copy(b[off:], fi.MimeType)
	off += copy(b[off:], fi.AttName)

	binary.LittleEndian.PutUint32(b[off:], fi.Utype^dwMask)
	off += 4
	binary.LittleEndian.PutUint64(b[off:], uint64(fi.PutTime^dwMask2))
	off += LenPutTime
	binary.LittleEndian.PutUint32(b[off:], fi.FileType^dwMask)
	off += LenFileType

	for i := 65; i < off; i++ {
		b[i] ^= fi.ChMask
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	copy(b[:20], h.Sum(nil))

	log.Debug("Fhandle:\n" + hex.Dump(b))

	return base64.URLEncoding.EncodeToString(b)
}

func DecodeFhandle_10_11_12(b []byte, oid string, kf KeyFinder) *FhandleInfo {
	if len(b) < 64 {
		log.Println("invalid efh, length < 64")
		return nil
	}

	fi := &FhandleInfo{}
	fi.Deadline = int64(binary.LittleEndian.Uint64(b[40:]))
	if fi.Deadline < time.Nanoseconds() {
		log.Println("Deadline expired")
		return nil
	}

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	/*	fi.OidHigh = binary.LittleEndian.Uint64(b[56:]) ^ uint64(dwMask2)
		fi.OidLow = binary.LittleEndian.Uint32(b[52:]) ^ dwMask

		if fi.OidHigh != 0 && fi.OidLow != 0 {
			oidLow, oidHigh, err := objid.Decode(oid)
			if err != nil {
				log.Println("DecodeFhandle: objid.Decode failed")
				return nil
			}
			if fi.OidHigh != oidHigh || fi.OidLow != oidLow {
				log.Println("DecodeFhandle: SessionID not matched")
				return nil
			}
		}
	*/
	fi.KeyHint = binary.LittleEndian.Uint32(b[28:]) ^ dwMask
	key := kf.Find(fi.KeyHint)
	if key == nil {
		log.Println("KeyFinder: key not found")
		return nil
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	if !bytes.Equal(h.Sum(nil), b[:20]) {
		log.Println("Digest verify failed")
		return nil
	}

	fi.Ver = uint16(b[20])
	if fi.Ver != FhandleVer_12 && fi.Ver != FhandleVer_11 && fi.Ver != FhandleVer_10 {
		log.Println("FhandleVer not match")
		return nil
	}

	fi.Fsize = int64(binary.LittleEndian.Uint64(b[32:])) ^ dwMask2
	fi.Uid = binary.LittleEndian.Uint32(b[48:]) ^ dwMask

	fi.Compression = b[22]
	fi.ChMask = b[23]
	fi.Public = b[27] ^ 0x27

	for i := 64; i < len(b); i++ {
		b[i] ^= fi.ChMask
	}

	LenMimeType := int(b[25])
	LenAttName := int(b[26])

	//	off := 64 + int(b[24])*20 + 1
	off := len(b) - LenMimeType - LenAttName
	if fi.Ver >= FhandleVer_11 {
		off -= 4
	}
	if fi.Ver >= FhandleVer_12 {
		off -= LenPutTime
	}

	fi.Fhandle = b[64:off]

	fi.MimeType = string(b[off : off+LenMimeType])
	off += LenMimeType

	fi.AttName = string(b[off : off+LenAttName])
	off += LenAttName

	if fi.Ver >= FhandleVer_11 {
		fi.Utype = binary.LittleEndian.Uint32(b[off:]) ^ dwMask
		off += 4
	}
	if fi.Ver >= FhandleVer_12 {
		fi.PutTime = int64(binary.LittleEndian.Uint64(b[off:])) ^ dwMask2
	}

	return fi
}

func DecodeFhandle_13(b []byte, oid string, kf KeyFinder) *FhandleInfo {
	if len(b) < 65 {
		log.Println("invalid efh, length < 65")
		return nil
	}

	fi := &FhandleInfo{}
	fi.Deadline = int64(binary.LittleEndian.Uint64(b[41:]))
	if fi.Deadline < time.Nanoseconds() {
		log.Println("Deadline expired")
		return nil
	}

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	fi.KeyHint = binary.LittleEndian.Uint32(b[29:]) ^ dwMask
	key := kf.Find(fi.KeyHint)
	if key == nil {
		log.Println("KeyFinder: key not found")
		return nil
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	if !bytes.Equal(h.Sum(nil), b[:20]) {
		log.Println("Digest verify failed")
		return nil
	}

	fi.Ver = uint16(b[20])
	if fi.Ver != FhandleVer_13 {
		log.Println("FhandleVer not match")
		return nil
	}

	fi.Fsize = int64(binary.LittleEndian.Uint64(b[33:])) ^ dwMask2
	fi.Uid = binary.LittleEndian.Uint32(b[49:]) ^ dwMask

	fi.Compression = b[22]
	fi.ChMask = b[23]
	fi.Public = b[28] ^ 0x27

	for i := 65; i < len(b); i++ {
		b[i] ^= fi.ChMask
	}

	LenMimeType := int(b[25])
	LenAttName := int(binary.LittleEndian.Uint16(b[26:]))

	//	off := 65 + int(b[24])*20 + 1
	off := len(b) - LenMimeType - LenAttName
	off -= 4
	off -= LenPutTime

	fi.Fhandle = b[65:off]

	fi.MimeType = string(b[off : off+LenMimeType])
	off += LenMimeType

	fi.AttName = string(b[off : off+LenAttName])
	off += LenAttName

	fi.Utype = binary.LittleEndian.Uint32(b[off:]) ^ dwMask
	off += 4
	fi.PutTime = int64(binary.LittleEndian.Uint64(b[off:])) ^ dwMask2

	return fi
}

func DecodeFhandle_14(b []byte, oid string, kf KeyFinder) *FhandleInfo {
	if len(b) < 65 {
		log.Println("invalid efh, length < 65")
		return nil
	}

	fi := &FhandleInfo{}
	fi.Deadline = int64(binary.LittleEndian.Uint64(b[41:]))
	if fi.Deadline < time.Nanoseconds() {
		log.Println("Deadline expired")
		return nil
	}

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	fi.KeyHint = binary.LittleEndian.Uint32(b[29:]) ^ dwMask
	key := kf.Find(fi.KeyHint)
	if key == nil {
		log.Println("KeyFinder: key not found")
		return nil
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	if !bytes.Equal(h.Sum(nil), b[:20]) {
		log.Println("Digest verify failed")
		return nil
	}

	fi.Ver = uint16(b[20])
	if fi.Ver != FhandleVer_14 {
		log.Println("FhandleVer not match")
		return nil
	}

	fi.Fsize = int64(binary.LittleEndian.Uint64(b[33:])) ^ dwMask2
	fi.Uid = binary.LittleEndian.Uint32(b[49:]) ^ dwMask

	fi.Compression = b[22]
	fi.ChMask = b[23]
	fi.Public = b[28] ^ 0x27

	for i := 65; i < len(b); i++ {
		b[i] ^= fi.ChMask
	}

	LenMimeType := int(b[25])
	LenAttName := int(binary.LittleEndian.Uint16(b[26:]))

	//	off := 65 + int(b[24])*20 + 1
	off := len(b) - LenMimeType - LenAttName
	off -= 4
	off -= LenPutTime
	off -= LenFileType

	fi.Fhandle = b[65:off]

	fi.MimeType = string(b[off : off+LenMimeType])
	off += LenMimeType

	fi.AttName = string(b[off : off+LenAttName])
	off += LenAttName

	fi.Utype = binary.LittleEndian.Uint32(b[off:]) ^ dwMask
	off += 4
	fi.PutTime = int64(binary.LittleEndian.Uint64(b[off:])) ^ dwMask2
	off += 8
	fi.FileType = binary.LittleEndian.Uint32(b[off:]) ^ dwMask

	return fi
}

func DecodeFhandle(efh string, oid string, kf KeyFinder) *FhandleInfo {
	b, err := base64.URLEncoding.DecodeString(efh)
	if err != nil || len(b) < 64 {
		log.Println("DecodeString failed:", err)
		return nil
	}

	ver := uint16(b[20])
	if ver >= FhandleVer_14 {
		return DecodeFhandle_14(b, oid, kf)
	} else if ver >= FhandleVer_13 {
		return DecodeFhandle_13(b, oid, kf)
	} else {
		return DecodeFhandle_10_11_12(b, oid, kf)
	}
}

// --------------------------------------------------------------------
