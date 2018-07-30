package types

import (
	"qbox.us/fh/fhver"
	"qbox.us/fh/sha1bdebd"
	pfd "qbox.us/pfd/api/types"
)

const (
	Upibd = 32767
)

// 目前有两种情况
// 1. "qbox.us/fh/pfd"，版本号为5、6
// 2. "qbox.us/fh/sha1bdebd"， 版本号为7
// 对于所有情况，最后总会得到6版本的fh
func DecodeFh(fh []byte) (fhi *pfd.FileHandle, err error) {
	switch fhver.FhVer(fh) {
	case fhver.FhPfd, fhver.FhPfdV2:
		return pfd.DecodeFh(fh)
	case fhver.FhSha1bdEbd:
		fhiv7, err := sha1bdebd.DecodeFh(fh)
		if err != nil {
			return nil, err
		}
		fhi = &pfd.FileHandle{
			Ver:    fhver.FhPfdV2,
			Tag:    pfd.FileHandle_ChunkBits,
			Upibd:  Upibd,
			Gid:    fhiv7.Gid,
			Offset: -1,
			Fsize:  fhiv7.Fsize,
			Fid:    fhiv7.Fid,
		}
		copy(fhi.Hash[:], fhiv7.SbdInstance.Etag()[1:])
		return fhi, nil
	default:
		panic("unknown fhver")
	}
}
