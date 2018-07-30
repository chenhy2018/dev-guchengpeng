package api

import (
	"io"

	"github.com/qiniu/encoding/binary"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"qbox.us/errors"
	pfdstgapi "qbox.us/pfdstg/api"
)

const (
	StatusNoSuchEntry = 612
	StatusNothingTodo = 202
)

var (
	ErrNoSuchEntry = httputil.NewError(StatusNoSuchEntry, "no such entry")
	ErrNothingTodo = httputil.NewError(StatusNothingTodo, "nothing todo")
	ErrCrc64       = errors.New("unmatched crc64")
)

var (
	O_RepairBytes  = []byte{0x6f, 0x64, 0x65, 0x2e}
	O_MigrateBytes = []byte{0x2e, 0x55, 0x54, 0x43}
	O_RecycleBytes = []byte{0x52, 0x65, 0x63, 0x79}
	O_CheckBytes   = []byte{0x03, 0xfa, 0xfb, 0x40}
)

// ---------------------------------------------------------
const (
	ValidateLength = 16 // 8个字节的version字段+8个字节的crc64字段。
)

// StripeBuildInfo如果新增字段，必须放在Crc64字段之前, Version之后
type StripeBuildInfo struct {
	Version  int64
	Sid      uint64 // stripe id
	Dgid     uint32
	Psectors [N + M]uint64
	OldSuids [N + M]uint64
	Migrate  pfdstgapi.MigrateInfo
	Crc64    uint64
}

func ReadStripeBuildInfo(r io.Reader) (sbi *StripeBuildInfo, err error) {
	r = NewCrc64StreamReader(r)
	var version int64
	if err, version = CheckVersion(r); err != nil {
		return
	}
	sbi = new(StripeBuildInfo)
	sbi.Version = version
	if err = binary.Read(r, binary.LittleEndian, &sbi.Sid); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &sbi.Dgid); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &sbi.Psectors); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &sbi.OldSuids); err != nil {
		return
	}
	m, err := pfdstgapi.DecodeMigrateInfo(r)
	if err != nil {
		return
	}
	sbi.Migrate = *m
	actualCrc64 := r.(*Crc64StreamReader).Sum64()
	if err = binary.Read(r, binary.LittleEndian, &sbi.Crc64); err != nil {
		return
	}
	if actualCrc64 != sbi.Crc64 {
		err = ErrCrc64
	}
	return
}

func EncodeStripeBuildInfo(w io.Writer, sbi *StripeBuildInfo) (err error) {
	w = NewCrc64StreamWriter(w)
	if err = binary.Write(w, binary.LittleEndian, CurrentVersion()); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, &sbi.Sid); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, &sbi.Dgid); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, &sbi.Psectors); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, &sbi.OldSuids); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, &sbi.Migrate); err != nil {
		return
	}
	_, err = w.(*Crc64StreamWriter).WriteSum64()
	return
}

// ---------------------------------------------------------
// StripeRepairInfo如果新增字段，必须放在Crc64字段之前, Version之后
type StripeRepairInfo struct {
	Version  int64
	Sid      uint64
	Bads     [M]int8 // 值范围：-1..N+M-1；其中 0..N+M-1 是正常值，-1 表示没坏
	Psectors [N + M]uint64
	OldSuids [N + M]uint64
	Crc32s   [N + M]uint32
	Crc64    uint64
}

func ReadStripeRepairInfo(r io.Reader) (sri *StripeRepairInfo, err error) {
	var version int64
	r = NewCrc64StreamReader(r)
	sri = new(StripeRepairInfo)
	if err, version = CheckVersion(r); err != nil {
		return
	}
	sri.Version = version
	if err = binary.Read(r, binary.LittleEndian, &sri.Sid); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &sri.Bads); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &sri.Psectors); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &sri.OldSuids); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &sri.Crc32s); err != nil {
		return
	}
	actualCrc64 := r.(*Crc64StreamReader).Sum64()
	if err = binary.Read(r, binary.LittleEndian, &sri.Crc64); err != nil {
		return
	}
	if actualCrc64 != sri.Crc64 {
		err = ErrCrc64
	}
	return
}

func EncodeStripeRepairInfo(w io.Writer, sri *StripeRepairInfo) (err error) {
	w = NewCrc64StreamWriter(w)
	if err = binary.Write(w, binary.LittleEndian, CurrentVersion()); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, &sri.Sid); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, &sri.Bads); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, &sri.Psectors); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, &sri.OldSuids); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, &sri.Crc32s); err != nil {
		return
	}
	_, err = w.(*Crc64StreamWriter).WriteSum64()
	return
}

// -----------------------------------------------------------------------------
// StripeRecycleBuildInfo如果新增字段，必须放在Crc64字段之前, Version之后
type StripeRecycleBuildInfo struct {
	Version       int64
	Sid           uint64
	Psectors      [N + M]uint64
	OldSuids      [N + M]uint64
	FirstFidBegin int64
	LastFidEnd    int64
	Fids          []uint64
	FileInfo      []FileInfo // 存储的文件信息
	Crc64         uint64
}

func ReadStripeRecycleInfo(r io.Reader) (srbi *StripeRecycleBuildInfo, err error) {
	var version int64
	r = NewCrc64StreamReader(r)
	srbi = new(StripeRecycleBuildInfo)
	if err, version = CheckVersion(r); err != nil {
		return
	}
	srbi.Version = version
	if err = binary.Read(r, binary.LittleEndian, &srbi.Sid); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &srbi.Psectors); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &srbi.OldSuids); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &srbi.FirstFidBegin); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &srbi.LastFidEnd); err != nil {
		return
	}

	var fileNumber uint32
	if err = binary.Read(r, binary.LittleEndian, &fileNumber); err != nil {
		return
	}
	fids := make([]uint64, int(fileNumber))
	for i := 0; i < int(fileNumber); i++ {
		if err = binary.Read(r, binary.LittleEndian, &fids[i]); err != nil {
			return
		}
	}
	srbi.Fids = fids
	fileInfos := make([]FileInfo, int(fileNumber))
	for i := 0; i < int(fileNumber); i++ {
		if err = binary.Read(r, binary.LittleEndian, &fileInfos[i].Soff); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &fileInfos[i].Fsize); err != nil {
			return
		}
		var pNumber uint32
		if err = binary.Read(r, binary.LittleEndian, &pNumber); err != nil {
			return
		}
		fileInfos[i].Suids = make([]uint64, pNumber)
		fileInfos[i].Psectors = make([]uint64, int(pNumber))
		if err = binary.Read(r, binary.LittleEndian, &fileInfos[i].Suids); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &fileInfos[i].Psectors); err != nil {
			return
		}
	}
	srbi.FileInfo = fileInfos
	actualCrc64 := r.(*Crc64StreamReader).Sum64()
	if err = binary.Read(r, binary.LittleEndian, &srbi.Crc64); err != nil {
		return
	}
	if actualCrc64 != srbi.Crc64 {
		err = ErrCrc64
	}
	return
}

func EncodeStripeRecycleInfo(w io.Writer, srbi *StripeRecycleBuildInfo) (err error) {
	w = NewCrc64StreamWriter(w)
	if err = binary.Write(w, binary.LittleEndian, CurrentVersion()); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, srbi.Sid); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, srbi.Psectors); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, srbi.OldSuids); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, srbi.FirstFidBegin); err != nil {
		return
	}
	if err = binary.Write(w, binary.LittleEndian, srbi.LastFidEnd); err != nil {
		return
	}
	var fileNumber = uint32(len(srbi.FileInfo))
	if err = binary.Write(w, binary.LittleEndian, fileNumber); err != nil {
		return
	}
	for i := 0; i < int(fileNumber); i++ {
		if err = binary.Write(w, binary.LittleEndian, srbi.Fids[i]); err != nil {
			return
		}
	}
	for i := 0; i < int(fileNumber); i++ {
		if err = binary.Write(w, binary.LittleEndian, srbi.FileInfo[i].Soff); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, srbi.FileInfo[i].Fsize); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, uint32(len(srbi.FileInfo[i].Suids))); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, srbi.FileInfo[i].Suids); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, srbi.FileInfo[i].Psectors); err != nil {
			return
		}
	}
	_, err = w.(*Crc64StreamWriter).WriteSum64()
	return
}

// -----------------------------------------------------------------------------

func (r Client) Acquire(l rpc.Logger) (typ []byte, rc io.ReadCloser, err error) {

	resp, err := rpc.DefaultClient.PostEx(l, r.Host+"/acquire")
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}
	typ = make([]byte, 4)
	_, err = io.ReadFull(resp.Body, typ)
	if err != nil {
		resp.Body.Close()
		return
	}
	rc = resp.Body
	return
}
