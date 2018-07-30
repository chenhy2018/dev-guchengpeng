package sha1bdebd

import (
	"syscall"

	"qbox.us/fh/fhver"
	"qbox.us/fh/sha1bd"
	"qbox.us/pfd/api/types"

	"github.com/qiniu/bytes"
	"github.com/qiniu/encoding/binary"
)

type FileHandle struct {
	Ver         uint8
	Tag         uint8
	Gid         types.Gid
	Fsize       int64
	Fid         uint64
	SbdInstance sha1bd.Instance
}

const HeaderSize = 1 + 1 + 12 + 8 + 8
const FileHandle_ChunkBits = 0x96

func EncodeFh(fhi *FileHandle) []byte {
	fhi.Ver = fhver.FhSha1bdEbd
	fhi.Tag = FileHandle_ChunkBits

	fh := make([]byte, HeaderSize+len(fhi.SbdInstance))
	w := bytes.NewWriter(fh)
	binary.Write(w, binary.LittleEndian, fhi)
	return fh
}

func DecodeFh(fh []byte) (fhi *FileHandle, err error) {

	if len(fh) <= HeaderSize ||
		fh[0] != fhver.FhSha1bdEbd ||
		fh[1] != FileHandle_ChunkBits {

		err = syscall.EINVAL
		return
	}
	fhi = new(FileHandle)
	fhi.SbdInstance = make([]byte, len(fh)-HeaderSize)

	r := bytes.NewReader(fh)
	err = binary.Read(r, binary.LittleEndian, fhi)
	return
}
