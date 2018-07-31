// 具体实现参考 "qbox.us/pfd/api/types".FileHandle
package pfd

import (
	"crypto/sha1"
	"encoding/binary"

	"qbox.us/fh/proto"
	"qbox.us/fh/stream"

	"github.com/qiniu/xlog.v1"
)

const largeFileBits = 22
const largeFileSize = 1 << largeFileBits

type Instance []byte

func (p Instance) Fsize() int64 {

	return int64(binary.LittleEndian.Uint64(p[24:32]))
}

func (p Instance) Ibd() uint16 {

	return binary.LittleEndian.Uint16(p[2:4])
}

func (p Instance) Ibdc() uint16 {

	return 0
}

func (p Instance) Etag() []byte {

	etag := make([]byte, 21)
	etag[0] = largeFileBits
	if p.Fsize() > largeFileSize {
		etag[0] |= 0x80
	}
	copy(etag[1:], p[40:40+20])
	return etag
}

var zeroSha1 = sha1.New().Sum(nil)

func (p Instance) Sha1(xl *xlog.Logger, getter *proto.Getter, fsize int64) ([]byte, error) {

	if fsize == 0 {
		return zeroSha1, nil
	}

	if etag := p.Etag(); etag[0]&0x80 == 0 {
		return etag[1:], nil
	}

	r, err := p.Source(xl, getter, fsize)
	if err != nil {
		return nil, err
	}

	h := sha1.New()
	if err = r.RangeRead(h, 0, fsize); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (p Instance) Source(xl *xlog.Logger, getter *proto.Getter, fsize int64) (r proto.Source, err error) {

	fh := []byte(p)
	r = stream.New(xl, getter.Pfd, fh, fsize)
	return
}
