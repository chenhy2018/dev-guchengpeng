package sha1bd

import (
	"crypto/sha1"
	"encoding/binary"

	"github.com/qiniu/xlog.v1"
	"qbox.us/fh/proto"
)

type Instance []byte

// -----------------------------------------------------------------------------

func (p Instance) Ibd() uint16 {

	_, bds := DecodeFh([]byte(p))
	return bds[0]
}

func (p Instance) Ibdc() uint16 {

	fh := []byte(p)

	var ibdc uint16
	switch len(fh) % 20 {
	case 10:
		fh = fh[7:]
		fallthrough
	case 3:
		ibdc = binary.LittleEndian.Uint16(fh[:2])
	}
	return ibdc
}

func (p Instance) Etag() []byte {

	fh := []byte(p)
	switch len(fh) % 20 {
	case 3:
		fh = fh[2:]
	case 10:
		fh = fh[9:]
	}

	if len(fh) > 21 {
		h := sha1.New()
		h.Write(fh[1:])
		return append([]byte{0x80 | fh[0]}, h.Sum(nil)...)
	}
	if len(fh) == 1 {
		return append([]byte{fh[0]}, zeroSha1...)
	}
	return fh
}

var zeroSha1 = sha1.New().Sum(nil)

func (p Instance) Sha1(xl *xlog.Logger, getter *proto.Getter, fsize int64) ([]byte, error) {

	if fsize == 0 {
		return zeroSha1, nil
	}

	if fsize <= 1<<22 {
		return p.Etag()[1:], nil
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

	return NewStream(xl, getter.Sha1bd, []byte(p), fsize)
}
