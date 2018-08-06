package proto

import (
	"errors"
	"io"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

var ErrInvalidFh = errors.New("invalid fh")

/* -----------------------------------------------------------------------------

v1: fh[1] < 0x80 && len(fh) % 20 == 3:

type Sha1bdV1Fh struct {
	upidc uint16
	tag uint8  // chunkBits+largeFileFlag
	sha1Array [][20]byte
}

v2: fh[1] < 0x80 && len(fh) % 20 == 10:

type Sha1bdV2Fh struct {
	upibd [3]uint16  // 6B
	unused uint8     // 1B
	upidc uint16     // 2B
	tag uint8        // 1B
	sha1Array [][20]byte // n*20B
}

v4: fh[1] == 0x96 && fh[0] == 4:

type UrlbdFh struct {
	ver uint8		// 1B ver=4
	tag uint8		// 1B tag=0x96
	url string
}

v5: fh[1] == 0x96 && fh[0] == 5:

type PfdFh struct {
	ver uint8         // 1B ver=5
	tag uint8         // 1B tag=0x96
	upibd uint16      // 2B 保留
	gid [12]byte      // 12B dgid+round
	offset int64      // 8B
	fsize int64       // 8B
	fid uint64        // 8B
	hash [20]byte     // 20B
}

// ---------------------------------------------------------------------------*/

type Handle interface {
	Ibd() uint16
	Ibdc() uint16
	Etag() []byte
	Sha1(xl *xlog.Logger, getter *Getter, fsize int64) ([]byte, error)
	Source(xl *xlog.Logger, getter *Getter, fsize int64) (Source, error)
}

type Fsizer interface {
	Fsize() int64
}

// -----------------------------------------------------------------------------

type Source interface {
	io.ReadSeeker
	RangeRead(w io.Writer, from, to int64) (err error)
	QueryFhandle() (fh []byte, err error)      // TODO: remove
	UploadedSize() (uploaded int64, err error) // TODO: remove
}

// -----------------------------------------------------------------------------

type ReaderGetter interface {
	Get(l rpc.Logger, fh []byte, from, to int64) (io.ReadCloser, int64, error)
}

type ReaderGetter2 interface {
	Get(l rpc.Logger, fh []byte, from, to, fsize int64) (io.ReadCloser, int64, error)
}

type CommonGetter interface {
	Get(xl rpc.Logger, fh []byte, w io.Writer, from, to int64) (int64, error)
}

type Sha1bdGetter interface {
	Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) error
}

type Getter struct {
	Sha1bd Sha1bdGetter
	Pfd    CommonGetter
	Oss    CommonGetter
}

// -----------------------------------------------------------------------------
