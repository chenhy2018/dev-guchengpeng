package fhver

import (
	"crypto/sha1"
	"testing"

	"github.com/stretchr/testify.v1/require"
	"qbox.us/fh/ossbd"
)

func Test(t *testing.T) {

	fhs := [][]byte{
		[]byte{0, 0, 22},
		[]byte{0, 0, 22, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		[]byte{0, 0, 255, 255, 255, 255, 0, 2, 0, 22},
		[]byte{0, 0, 255, 255, 255, 255, 0, 2, 0, 22, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		[]byte{0, 127, 255, 255, 255, 255, 0, 2, 0, 22, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		[]byte{1, 128, 255, 255, 255, 255, 0, 2, 0, 22, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		[]byte{3, 22, 1, 0, 230, 141, 0, 0, 0, 0, 0, 0, 56, 233, 207, 113, 216, 245, 113, 157, 125, 17, 99, 128, 154, 155, 203, 30, 243, 113, 223, 25},
		[]byte{3, 22, 1, 0, 230, 141, 0, 0, 0, 0, 0, 0, 56, 233, 207, 113, 216, 245, 113, 157, 125, 17, 99, 128, 154, 155, 203, 30, 243, 113, 223, 25, 102, 0, 0, 0, 0, 0, 0, 0},
		[]byte{4, 22},
		[]byte{4, 150},
	}
	vers := []int{1, 1, 2, 2, 2, 0, 0, 0, 0, 4}
	for i := 0; i < len(fhs); i++ {
		if FhVer(fhs[i]) != vers[i] {
			t.Fatal("TestFh: unexpected ver =>", i, fhs[i], FhVer(fhs[i]), vers[i])
		}
	}
}

func TestProtectFh(t *testing.T) {
	// new FhPfdV2
	oldFh := []byte{FhPfdV2, 0x96, 22}
	require.Equal(t, FhPfdV2, FhVer(oldFh), "fhver")
	// protect FhPfdV2
	newFh, ok := ProtectFh(oldFh)
	require.Equal(t, true, ok, "protected")
	require.Equal(t, []byte{FhPfd, 0x96, 22}, newFh, "new fh")
	// unprotect
	unFh := UnProtectFh(newFh)
	require.Equal(t, oldFh, unFh, "unprotect")
	// protect other
	old := []byte{FhPfd, 0x96, 22}
	_, ok = ProtectFh(old)
	require.Equal(t, false, ok, "other fh should not protect")
	// protect ossbd
	s := sha1.Sum([]byte("a"))
	fh := ossbd.NewInstance(1, s[:], "bucket", []byte("key"))
	n, ok := ProtectFh(fh)
	require.False(t, ok)
	require.Equal(t, []byte(nil), n)
	require.Equal(t, []byte(fh), UnProtectFh(fh))
}
