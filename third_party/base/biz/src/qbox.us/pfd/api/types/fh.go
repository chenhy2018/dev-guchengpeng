package types

import (
	"syscall"

	"qbox.us/fh/fhver"

	"github.com/qiniu/bytes"
	"github.com/qiniu/encoding/binary"
)

// --------------------------------------------------------------------

const (
	FileHandle_ChunkBits = 0x96

	FileHandleBytes = 1 + 1 + 2 + 12 + 8 + 8 + 8 + 20
)

type FileHandle struct {
	Ver    uint8    // 1  Byte
	Tag    uint8    // 1  Byte
	Upibd  uint16   // 2  Byte
	Gid    Gid      // 12 Byte
	Offset int64    // 8  Byte
	Fsize  int64    // 8  Byte
	Fid    uint64   // 8  Byte
	Hash   [20]byte // 20 Byte
}

func EncodeFh(fhi *FileHandle) []byte {

	fh := make([]byte, FileHandleBytes)
	w := bytes.NewWriter(fh)
	binary.Write(w, binary.LittleEndian, fhi)
	return fh
}

func DecodeFh(fh []byte) (fhi *FileHandle, err error) {

	if len(fh) != FileHandleBytes ||
		(fh[0] != fhver.FhPfd && fh[0] != fhver.FhPfdV2) ||
		fh[1] != FileHandle_ChunkBits {

		err = syscall.EINVAL
		return
	}
	fhi = new(FileHandle)

	r := bytes.NewReader(fh)
	err = binary.Read(r, binary.LittleEndian, fhi)
	return
}

// --------------------------------------------------------------------
