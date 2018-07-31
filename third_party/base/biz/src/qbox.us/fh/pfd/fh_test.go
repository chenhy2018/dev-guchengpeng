package pfd

import (
	"bytes"
	"testing"

	"qbox.us/pfd/api/types"
)

func TestInstance(t *testing.T) {

	fhs := make([][]byte, 4)
	e0s := []byte{0x16, 0x16, 0x96, 0x96}
	etags := make([][]byte, 4)
	for i := range fhs {
		fhi := &types.FileHandle{
			Ver:   5,
			Tag:   0x96,
			Upibd: uint16(i + 1),
			Fsize: int64(i+1) * 2 * 1024 * 1024,
			Hash:  [20]byte{byte(1 * i), byte(2 * i), byte(3 * i)},
		}
		fhs[i] = types.EncodeFh(fhi)
		etags[i] = append([]byte{e0s[i]}, fhi.Hash[:]...)
	}

	for i, fh := range fhs {
		inst := Instance(fh)
		if ibd := inst.Ibd(); ibd != uint16(i+1) {
			t.Fatal("TestInstance:", i, "-> ibd ->", ibd, i+1)
		}
		if ibdc := inst.Ibdc(); ibdc != 0 {
			t.Fatal("TestInstance:", i, "-> ibdc ->", ibdc, 0)
		}
		if etag := inst.Etag(); !bytes.Equal(etag, etags[i]) {
			t.Fatal("TestInstance:", i, "-> etag ->", etag, etags[i])
		}
		if fsize := inst.Fsize(); fsize != int64(i+1)*2*1024*1024 {
			t.Fatal("TestInstance:", i, "-> fsize ->", fsize, int64(i+1)*2*1024*1024)
		}
	}
}
