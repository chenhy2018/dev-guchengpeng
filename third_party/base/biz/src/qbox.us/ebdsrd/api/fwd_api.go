package api

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

var (
	ErrInvalidLength = errors.New("invalid length")
)

type FwdClient struct {
	rpc.Client
	Fwdward
}

type Fwdward interface {
	FwdPut(l rpc.Logger, host string, sid uint64, fids []uint64) error
	FwdDelete(l rpc.Logger, host string, sid uint64) error
}

func (client *FwdClient) FwdPut(l rpc.Logger, host string, item SidReverseItem) error {
	sid := item.Sid
	fidsstr := EncodeFids(item.Fids)
	params := map[string][]string{
		"sid":     {strconv.FormatUint(sid, 10)},
		"fidsstr": {fidsstr},
	}
	url := fmt.Sprintf("%v/put", host)
	return client.CallWithForm(l, nil, url, params)
}

func (client *FwdClient) FwdDelete(l rpc.Logger, host string, sid uint64) error {
	url := fmt.Sprintf("%v/delete/%v", host, sid)
	err := client.Call(l, nil, url)
	return err
}

func EncodeFids(fids []uint64) string {
	n := len(fids) * 8
	b := make([]byte, n)
	for i, fid := range fids {
		binary.LittleEndian.PutUint64(b[i*8:], fid)
	}
	if n == 0 {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func DecodeFids(fidstr string) ([]uint64, error) {
	if len(fidstr) == 0 {
		return []uint64{}, nil
	}
	b, err := base64.URLEncoding.DecodeString(fidstr)
	fids := make([]uint64, 0, len(b)%8)
	if err != nil {
		return fids, err
	}
	if len(b)%8 != 0 {
		return fids, ErrInvalidLength
	}
	for {
		fid := binary.LittleEndian.Uint64(b[:8])
		fids = append(fids, fid)
		b = b[8:]
		if len(b) == 0 {
			break
		}
	}
	return fids, nil
}
