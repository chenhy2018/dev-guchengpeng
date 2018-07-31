package direct

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"testing"
	"unsafe"
)

// ---------------------------------------------------------

const (
	globalBytes = 32
)

type global struct {
	Did uint64 // 当前 did

	Iblk     uint32 // 当前 block 号
	IblkLeft uint32 // 剩余 block 个数

	Off4k     uint32 // 当前数据区偏移 (in 4k)
	Off4kLeft uint32 // 当前剩余数据区大小 (in 4k)

	DidRangeCnt uint32 // 当前 did range 个数
	Crc32       uint32
}

func (g *global) calCrc32() uint32 {

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, g)
	data := buf.Bytes()
	data = data[:len(data)-4]
	return crc32.ChecksumIEEE(data)
}

// ---------------------------------------------------------

func Test(t *testing.T) {

	g := &global{Off4k: 234}
	crc := g.calCrc32()
	crc2 := Crc32(unsafe.Pointer(g), globalBytes-4)
	fmt.Println(crc, crc2)
	if crc != crc2 {
		t.Fatal("ChecksumIEEE failed:", crc, crc2)
	}
}

// ---------------------------------------------------------
