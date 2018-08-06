package fhver

import "bytes"

// -----------------------------------------------------------------------------

const (
	FhUnknown   = 0
	FhSha1bdV1  = 1
	FhSha1bdV2  = 2
	_           = 3 // 过时，已经废弃不用
	FhUrlbd     = 4
	FhPfd       = 5
	FhPfdV2     = 6 // 跟 FhPfd 的差别是这个版本保证 fh 不会被多个资源引用
	FhSha1bdEbd = 7
	FhOssBD     = 8
	// 新版本Fh如果支持删除，必须支持保护功能，需要修改"IsFileCanBeDeleted"、"ProtectFh"、"UnProtectFh"三个函数
)

func FhVer(fh []byte) int {

	tag := fh[1]
	if tag < 0x80 {
		switch len(fh) % 20 {
		case 3:
			return FhSha1bdV1
		case 10:
			return FhSha1bdV2
		}
		return FhUnknown
	}
	if tag == 0x96 {
		return int(fh[0])
	}
	return FhUnknown
}

// -----------------------------------------------------------------------------

// 检查一个fh对应的文件能否被删除，当前只有FhPfdV2可以被删除
func IsFileCanBeDeleted(fh []byte) bool {
	ver := FhVer(fh)
	return ver == FhPfdV2 || ver == FhOssBD
}

// 被保护的fh不可被物理删除，当前只有FhPfdV2支持保护，成功保护之后protected=true
func ProtectFh(fh []byte) (newFh []byte, protected bool) {
	if FhVer(fh) != FhPfdV2 {
		return
	}
	newFh = make([]byte, len(fh))
	copy(newFh, fh)
	newFh[0] = FhPfd
	protected = true
	return
}

// 只有被保护的fh才可以取消保护
func UnProtectFh(fh []byte) (newFh []byte) {
	newFh = make([]byte, len(fh))
	ver := FhVer(fh)
	if ver == FhOssBD {
		copy(newFh, fh)
		return
	}
	if ver != FhPfd {
		panic("only FhPfd can be unprotected")
	}
	copy(newFh, fh)
	newFh[0] = FhPfdV2
	return
}

// -----------------------------------------------------------------------------
// 检查fh是否相同，这里认为降级之后的fh和未降级的是相同的
func IsFhEqual(fh1, fh2 []byte) bool {
	pfh1, ok := ProtectFh(fh1)
	if !ok {
		pfh1 = fh1
	}
	pfh2, ok := ProtectFh(fh2)
	if !ok {
		pfh2 = fh2
	}
	return bytes.Equal(pfh1, pfh2)
}
