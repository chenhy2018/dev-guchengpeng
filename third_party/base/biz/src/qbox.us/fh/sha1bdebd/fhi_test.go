package sha1bdebd

import (
	"crypto/rand"

	"github.com/stretchr/testify.v1/require"

	"qbox.us/fh/fhver"
	"qbox.us/pfd/api/types"

	"testing"
)

func TestFileHandle(t *testing.T) {
	sbdInstance := make([]byte, 10)
	rand.Read(sbdInstance)

	fhi := &FileHandle{
		Gid:         types.NewGid(1),
		Fsize:       12,
		Fid:         ^uint64(0),
		SbdInstance: sbdInstance,
	}
	fh := EncodeFh(fhi)

	fhi2, err := DecodeFh(fh)
	require.NoError(t, err)
	require.Equal(t, &FileHandle{
		Ver:         fhver.FhSha1bdEbd,
		Tag:         0x96,
		Gid:         fhi.Gid,
		Fsize:       fhi.Fsize,
		Fid:         fhi.Fid,
		SbdInstance: fhi.SbdInstance,
	}, fhi2)

	fh2 := EncodeFh(fhi2)
	require.Equal(t, fh, fh2)
}
