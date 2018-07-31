package etag

import (
	"bytes"
	"crypto/sha1"
	"testing"
)

func Test(t *testing.T) {

	b := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	fh1 := append([]byte{1, 0, 0x16}, b...)
	fh2 := append([]byte{1, 0, 0x96}, b...)
	fh3 := append(fh1, b...)
	fh4 := append([]byte{1, 0, 255, 255, 255, 255, 0}, fh1...)
	fh5 := append([]byte{1, 0, 255, 255, 255, 255, 0}, fh2...)
	fh6 := append(fh4, b...)

	etag1 := Gen(fh1)
	etag2 := Gen(fh2)
	etag3 := Gen(fh3)
	etag4 := Gen(fh4)
	etag5 := Gen(fh4)
	etag6 := Gen(fh4)
	if len(etag1) != 21 || len(etag2) != 21 || len(etag3) != 21 ||
		len(etag4) != 21 || len(etag5) != 21 || len(etag6) != 21 {
		t.Fatal("error length")
	}

	if !bytes.Equal(etag1, append([]byte{0x16}, b...)) {
		t.Fatal("error etag")
	}

	if !bytes.Equal(etag2, append([]byte{0x96}, b...)) {
		t.Fatal("error etag")
	}

	h := sha1.New()
	h.Write(b)
	h.Write(b)
	if !bytes.Equal(etag3, append([]byte{0x96}, h.Sum(nil)...)) {
		t.Fatal("error etag")
	}

	if GenString(fh1) != GenString(fh4) {
		t.Fatal("version incompatible")
	}

	if GenString(fh2) != GenString(fh5) {
		t.Fatal("version incompatible")
	}

	if GenString(fh3) != GenString(fh6) {
		t.Fatal("version incompatible")
	}
}
