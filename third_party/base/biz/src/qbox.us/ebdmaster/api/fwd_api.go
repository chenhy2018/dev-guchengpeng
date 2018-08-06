package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

const FidInfoLen = 28 + 2 /*使用1个字节标记文件是否被删除+1个字节标记文件文件是否可以被删除*/

type FidInfo struct {
	Fid     uint64
	Fsize   int64
	Sid     uint64
	Soff    uint32
	Deleted uint8 //delete=0时, 文件是未删除状态。
	Lock    uint8 //Lock为1的时候, 文件不能被删除
}

func EncodeFidInfos(w io.Writer, infos []FidInfo) (err error) {
	w = NewCrc64StreamWriter(w)
	if err = binary.Write(w, binary.LittleEndian, CurrentVersion()); err != nil {
		return
	}
	b := make([]byte, FidInfoLen)
	for _, info := range infos {
		binary.LittleEndian.PutUint64(b, info.Fid)
		binary.LittleEndian.PutUint64(b[8:], uint64(info.Fsize))
		binary.LittleEndian.PutUint64(b[16:], info.Sid)
		binary.LittleEndian.PutUint32(b[24:], info.Soff)
		b[28] = byte(info.Deleted)
		b[29] = byte(info.Lock)
		if _, err = w.Write(b); err != nil {
			return
		}
	}
	_, err = w.(*Crc64StreamWriter).WriteSum64()
	return
}

func DecodeFidInfos(r io.Reader, n int) ([]FidInfo, error) {
	r = NewCrc64StreamReader(r)
	if err, _ := CheckVersion(r); err != nil {
		return nil, err
	}
	b := make([]byte, FidInfoLen)
	infos := make([]FidInfo, n)
	for i := 0; i < n; i++ {
		_, err := io.ReadFull(r, b)
		if err != nil {
			return nil, err
		}
		info := &infos[i]
		info.Fid = binary.LittleEndian.Uint64(b)
		info.Fsize = int64(binary.LittleEndian.Uint64(b[8:]))
		info.Sid = binary.LittleEndian.Uint64(b[16:])
		info.Soff = binary.LittleEndian.Uint32(b[24:])
		info.Deleted = uint8(b[28])
		info.Lock = uint8(b[29])
	}
	actualCrc64 := r.(*Crc64StreamReader).Sum64()
	expectCrc64 := uint64(0)
	if err := binary.Read(r, binary.LittleEndian, &expectCrc64); err != nil {
		return nil, err
	}
	if actualCrc64 != expectCrc64 {
		return nil, ErrCrc64
	}
	return infos, nil
}

// ----------------------------------------------------------------------------

type Forwarder interface {
	FwdMigrate(l rpc.Logger, host string, sid uint64, psects *[N + M]uint64, crc32s *[N + M]uint32, fids []FidInfo) error
	FwdRepair(l rpc.Logger, host string, sid uint64, psects *[N + M]uint64, crc32s *[N + M]uint32, bads [M]int8) error
	FwdRecycle(l rpc.Logger, host string, sid uint64, psects *[N + M]uint64, crc32s *[N + M]uint32, fids []FidInfo) error
	FwdDelete(l rpc.Logger, host string, fid uint64) error
}

// -----------------------------------------------------------------------------

type FwdClient struct {
	rpc.Client
}

func (c FwdClient) FwdMigrate(l rpc.Logger, host string, sid uint64, psects *[N + M]uint64, crc32s *[N + M]uint32, fids []FidInfo) error {

	psect := EncodePsects(psects)
	crc32 := EncodeCrc32s(crc32s)
	length := len(fids) * FidInfoLen + ValidateLength // 8字节版本号+8字节crc64
	buf := bytes.NewBuffer(make([]byte, 0, length))
	err := EncodeFidInfos(buf, fids)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%v/fwd/migrate?sid=%v&psect=%v&crc32s=%v", host, sid, psect, crc32)
	return c.CallWith(l, nil, url, "application/octet-stream", buf, length)
}

func (c FwdClient) FwdRepair(l rpc.Logger, host string, sid uint64, psects *[N + M]uint64, crc32s *[N + M]uint32, bads [M]int8) error {

	params := map[string][]string{
		"sid":    []string{strconv.FormatUint(sid, 10)},
		"psect":  []string{EncodePsects(psects)},
		"crc32s": []string{EncodeCrc32s(crc32s)},
		"bads":   []string{EncodeBads(bads)},
	}
	return c.CallWithForm(l, nil, host+"/fwd/repair", params)
}

func (c FwdClient) FwdRecycle(l rpc.Logger, host string, sid uint64, psects *[N + M]uint64, crc32s *[N + M]uint32, fids []FidInfo) error {

	psect := EncodePsects(psects)
	crc32 := EncodeCrc32s(crc32s)
	length := len(fids) * FidInfoLen + ValidateLength
	buf := bytes.NewBuffer(make([]byte, 0, length))
	err := EncodeFidInfos(buf, fids)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%v/fwd/recycle?sid=%v&psect=%v&crc32s=%v", host, sid, psect, crc32)
	return c.CallWith(l, nil, url, "application/octet-stream", buf, length)
}

func (c FwdClient) FwdDelete(l rpc.Logger, host string, fid uint64) error {

	url := fmt.Sprintf("%v/fwd/delete?fid=%v", host, fid)
	return c.Call(l, nil, url)
}
