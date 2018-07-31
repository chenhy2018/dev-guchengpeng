package sha1bd

import (
	"bytes"
	"crypto/sha1"
	"testing"
)

type testInst struct {
	fh   []byte
	ibd  uint16
	ibdc uint16
	etag []byte
}

func TestInstance(t *testing.T) {

	keys := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	h := sha1.New()
	h.Write(bytes.Repeat(keys, 2))
	sha := h.Sum(nil)
	tests := []testInst{
		{
			fh:   append([]byte{1, 2, 22}, keys...),
			ibd:  0,
			ibdc: 513,
			etag: append([]byte{22}, keys...),
		},
		{
			fh:   append(append([]byte{10, 2, 22}, keys...), keys...),
			ibd:  0,
			ibdc: 522,
			etag: append([]byte{0x96}, sha...),
		},
		{
			fh:   append([]byte{2, 1, 0x96}, keys...),
			ibd:  0,
			ibdc: 258,
			etag: append([]byte{0x96}, keys...),
		},
		{
			fh:   append([]byte{3, 2, 255, 255, 255, 255, 0, 1, 2, 22}, keys...),
			ibd:  515,
			ibdc: 513,
			etag: append([]byte{22}, keys...),
		},
		{
			fh:   append(append([]byte{4, 1, 255, 255, 255, 255, 0, 10, 2, 22}, keys...), keys...),
			ibd:  260,
			ibdc: 522,
			etag: append([]byte{0x96}, sha...),
		},
		{
			fh:   []byte{2, 0, 255, 255, 255, 255, 0, 1, 0, 22},
			ibd:  2,
			ibdc: 1,
			etag: []byte{22, 218, 57, 163, 238, 94, 107, 75, 13, 50, 85, 191, 239, 149, 96, 24, 144, 175, 216, 7, 9},
		},
	}

	for i, test := range tests {
		inst := Instance(test.fh)
		if ibd := inst.Ibd(); ibd != test.ibd {
			t.Fatal("TestInstance:", i, "-> ibd ->", ibd, test.ibd)
		}
		if ibdc := inst.Ibdc(); ibdc != test.ibdc {
			t.Fatal("TestInstance:", i, "-> ibdc ->", ibdc, test.ibdc)
		}
		if etag := inst.Etag(); !bytes.Equal(etag, test.etag) {
			t.Fatal("TestInstance:", i, "-> etag ->", etag, test.etag)
		}
	}
}
