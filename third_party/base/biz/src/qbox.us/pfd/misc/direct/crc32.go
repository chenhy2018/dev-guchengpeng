package direct

import (
	"hash/crc32"
	"unsafe"
)

// ---------------------------------------------------------

//
// 计算一个结构的crc32值
//
func Crc32(pointer unsafe.Pointer, n int) uint32 {
	b := ((*[1 << 30]byte)(pointer))[:n]
	return crc32.ChecksumIEEE(b)
}

// ---------------------------------------------------------
