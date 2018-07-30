package sha1bdebd

import (
	"encoding/binary"

	"github.com/qiniu/xlog.v1"
	"qbox.us/fh/proto"
	"qbox.us/fh/sha1bd"
	"qbox.us/fh/stream"
)

type Instance struct {
	fh          []byte
	sbdInstance proto.Handle
}

func NewInstance(fh []byte) Instance {
	sbdInstance := sha1bd.Instance(fh[HeaderSize:])
	return Instance{fh: fh, sbdInstance: sbdInstance}
}

func (p Instance) Fsize() int64 {

	return int64(binary.LittleEndian.Uint64(p.fh[14:22]))
}

func (p Instance) Ibd() uint16 {
	return p.sbdInstance.Ibd()
}

func (p Instance) Ibdc() uint16 {
	return p.sbdInstance.Ibdc()
}

func (p Instance) Etag() []byte {
	return p.sbdInstance.Etag()
}

func (p Instance) Sha1(xl *xlog.Logger, getter *proto.Getter, fsize int64) ([]byte, error) {
	return p.sbdInstance.Sha1(xl, getter, fsize)
}

var ReadFromEbd = map[uint16]bool{}

func (p Instance) Source(xl *xlog.Logger, getter *proto.Getter, fsize int64) (proto.Source, error) {
	if ReadFromEbd[p.Ibd()] {
		r := stream.New(xl, getter.Pfd, p.fh, fsize)
		return r, nil
	}
	return p.sbdInstance.Source(xl, getter, fsize)
}
