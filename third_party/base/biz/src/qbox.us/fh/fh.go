package fh

import (
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/fh/fhver"
	"qbox.us/fh/ossbd"
	"qbox.us/fh/pfd"
	"qbox.us/fh/proto"
	"qbox.us/fh/sha1bd"
	"qbox.us/fh/sha1bdebd"
)

func New(fh []byte) proto.Handle {

	switch fhver.FhVer(fh) {
	case fhver.FhSha1bdV1, fhver.FhSha1bdV2:
		return sha1bd.Instance(fh)
	case fhver.FhPfd, fhver.FhPfdV2:
		return pfd.Instance(fh)
	case fhver.FhSha1bdEbd:
		return sha1bdebd.NewInstance(fh)
	case fhver.FhOssBD:
		return ossbd.Instance(fh)
	}

	log.Error("fh.New: unknown fh =>", fh)
	panic("unknown fh")
}

// -----------------------------------------------------------------------------

func Ibd(fh []byte) uint16 {

	return New(fh).Ibd()
}

func Ibdc(fh []byte) uint16 {

	return New(fh).Ibdc()
}

func Etag(fh []byte) []byte {

	return New(fh).Etag()
}

// fh must be Fsizer.
func Fsize(fh []byte) int64 {

	return New(fh).(proto.Fsizer).Fsize()
}

func Sha1(xl *xlog.Logger, getter *proto.Getter, fh []byte, fsize int64) ([]byte, error) {

	return New(fh).Sha1(xl, getter, fsize)
}

func Source(xl *xlog.Logger, getter *proto.Getter, fh []byte, fsize int64) (proto.Source, error) {

	return New(fh).Source(xl, getter, fsize)
}
