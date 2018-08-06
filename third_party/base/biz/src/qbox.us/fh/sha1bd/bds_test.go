package sha1bd

import (
	"bytes"
	"testing"
)

func TestDecode(t *testing.T) {

	sample := []byte{22, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	fh, bds := DecodeFh(sample)
	if !bytes.Equal(fh, sample) || bds != [4]uint16{0, 0xffff, 0xffff, 0xffff} {
		t.Fatal("DecodeFh sample:", sample, fh, bds)
	}

	fh2 := append([]byte{1, 0}, sample...)
	fh, bds = DecodeFh(fh2)
	if !bytes.Equal(fh, sample) || bds != [4]uint16{0, 0xffff, 0xffff, 1} {
		t.Fatal("DecodeFh fh2:", fh2, fh, bds)
	}

	fh3 := append([]byte{0, 0, 1, 0, 255, 255, 0, 2, 0}, sample...)
	fh, bds = DecodeFh(fh3)
	if !bytes.Equal(fh, sample) || bds != [4]uint16{0, 1, 0xffff, 2} {
		t.Fatal("DecodeFh fh3:", fh3, fh, bds)
	}
}
