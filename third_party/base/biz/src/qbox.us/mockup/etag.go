package mockup

import (
	"crypto/sha1"
	"encoding/base64"
	"io"
)

func CalSha1(r io.Reader) (sha1Code []byte, err error) {
	h := sha1.New()
	_, err = io.Copy(h, r)
	if err != nil {
		return
	}
	sha1Code = h.Sum(nil)
	return
}

//just for size < 4m
func GetEtag(r io.Reader) (etag string, err error) {
	sha1Buf := make([]byte, 0, 21)
	var sha1Code []byte
	sha1Buf = append(sha1Buf, 0x16)
	sha1Code, err = CalSha1(r)
	if err != nil {
		return
	}
	sha1Buf = append(sha1Buf, sha1Code...)
	etag = base64.URLEncoding.EncodeToString(sha1Buf)
	return
}
