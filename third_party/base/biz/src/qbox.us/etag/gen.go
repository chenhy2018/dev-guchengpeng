package etag

import (
	"crypto/sha1"
	"encoding/base64"
	"qbox.us/fh"
)

func Gen(fhb []byte) []byte {

	return fh.Etag(fhb)
}

func GenString(fh []byte) string {

	etag := Gen(fh)
	return base64.URLEncoding.EncodeToString(etag)
}

// fopETag = [0x00, sha1([srcETag, fopVerHigh, fopVerLow, fopDesc])]
func Fop(fh []byte, fopVerHigh, fopVerLow byte, fopDesc string) []byte {

	srcETag := Gen(fh)
	h := sha1.New()
	h.Write(srcETag)
	h.Write([]byte{fopVerHigh, fopVerLow})
	h.Write([]byte(fopDesc))
	v := h.Sum(nil)
	etag := make([]byte, 21)
	etag[0] = 0x00
	copy(etag[1:], v)
	return etag
}

func FopString(fh []byte, fopVerHigh, fopVerLow byte, fopDesc string) string {

	etag := Fop(fh, fopVerHigh, fopVerHigh, fopDesc)
	return base64.URLEncoding.EncodeToString(etag)
}
