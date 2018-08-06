package types

import (
	"bytes"
	"crypto/sha1"
	"reflect"
	"syscall"
	"testing"

	"qbox.us/fh/fhver"
)

func doTestFh(t *testing.T, fhinfo *FileHandle) {

	fh := EncodeFh(fhinfo)
	t.Log(fh)

	fhinfo2, err := DecodeFh(fh)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(fhinfo, fhinfo2) {
		t.Fatalf("fhinfo1: %#v, fhinfo2: %#v\n", fhinfo, fhinfo2)
	}

	fh2 := EncodeFh(fhinfo2)
	if !bytes.Equal(fh, fh2) {
		t.Fatalf("fh1: %v, fh2: %v", fh, fh2)
	}
}

func TestFh(t *testing.T) {

	hash := sha1.New()
	hash.Write([]byte("hello"))
	fhinfo := &FileHandle{
		Ver:    fhver.FhPfdV2,
		Tag:    FileHandle_ChunkBits,
		Upibd:  65535,
		Offset: 31,
		Fsize:  3 * 1024 * 1024,
		Gid:    Gid{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1},
		Fid:    1001234,
		Hash:   [20]byte{},
	}
	doTestFh(t, fhinfo)
	doTestFh(t, &FileHandle{
		Ver: fhver.FhPfd,
		Tag: FileHandle_ChunkBits,
	})
}

func TestInvalid(t *testing.T) {

	var (
		fh  []byte
		err error
	)

	fh = make([]byte, 0)
	_, err = DecodeFh(fh)
	if err != syscall.EINVAL {
		t.Fatal("err != syscall.EINVAL", "error:", err)
	}

	fh = nil
	_, err = DecodeFh(fh)
	if err != syscall.EINVAL {
		t.Fatal("err != syscall.EINVAL", "error:", err)
	}
}
