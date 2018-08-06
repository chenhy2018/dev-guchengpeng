package types

import "log"

const (
	// 24 bytes for sector info
	// crc32util.DecodeSize(16M - 24b)
	SUMaxLen = 16776168
)

func init() {
	log.Printf("N: %v, M: %v", N, M)
}
