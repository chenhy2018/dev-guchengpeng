package direct

import (
	"encoding/base64"
	"io"
	"unsafe"
)

// ---------------------------------------------------------

func Bytes(p unsafe.Pointer, n int) []byte {

	return ((*[1 << 30]byte)(p))[:n]
}

func EncodedBytes(p unsafe.Pointer, n int) string {

	return base64.URLEncoding.EncodeToString(Bytes(p, n))
}

// ---------------------------------------------------------

//
// 复制一块内存
//
func CopyTo(dest []byte, src unsafe.Pointer) {

	psrc := ((*[1 << 30]byte)(src))[:len(dest)]
	copy(dest, psrc)
}

func CopyFrom(dest unsafe.Pointer, src []byte) {

	pdest := ((*[1 << 30]byte)(dest))[:len(src)]
	copy(pdest, src)
}

// ---------------------------------------------------------

//
// 读取一个结构
//
func ReadAt(f io.ReaderAt, off int64, pointer unsafe.Pointer, n int) (err error) {

	b := ((*[1 << 30]byte)(pointer))[:n]
	_, err = f.ReadAt(b, off)
	return
}

// ---------------------------------------------------------

//
// 保存一个结构
//
func WriteAt(f io.WriterAt, off int64, pointer unsafe.Pointer, n int) (err error) {

	b := ((*[1 << 30]byte)(pointer))[:n]
	_, err = f.WriteAt(b, off)
	return
}

// ---------------------------------------------------------
