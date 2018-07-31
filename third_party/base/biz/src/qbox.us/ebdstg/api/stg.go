package api

import (
	"fmt"
	"io"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
)

// ---------------------------------------------------------

const (
	ErrCodeDuplicated = 222
)

var (
	ErrChecksumError  = httputil.NewError(802, "crc32 not matched")
	ErrInvalidPsect   = httputil.NewError(803, "invalid psector")
	ErrPsectUsed      = httputil.NewError(804, "psector used")
	ErrSuidNotMatched = httputil.NewError(810, "suid not matched")
	ErrNoSuchDisk     = httputil.NewError(811, "no such disk")
	ErrBrokenDisk     = httputil.NewError(812, "broken disk")

	ErrDuplicated          = httputil.NewError(ErrCodeDuplicated, "data with same suid exists")
	ErrUnmatchedDuplicated = httputil.NewError(400, "unmatched data with same suid")

	ErrDataTooLarge  = errors.New("data too large")
	ErrDataOverflow  = errors.New("data overflow")
	ErrOutofBoundary = errors.New("out of boundary")
	ErrInvalidSuid   = errors.New("invalid suid")
	ErrInvalidIsect  = errors.New("invalid isector")
)

var (
	DefaultClient rpc.Client = rpc.DefaultClient
)

func ErrDiskIoError(msg string) (err error) {
	return httputil.NewError(801, msg)
}

const (
	ErrReasonUnknown      = uint32(0)
	ErrReasonDiskIoError  = uint32(1)
	ErrReasonNetworkError = uint32(2)
	ErrReasonPsectUsed    = uint32(3)
	ErrReasonNoSuchDisk   = uint32(4)
	ErrReasonChecksum     = uint32(5)
	ErrReasonSuidNotMatch = uint32(6)
	ErrReasonBrokenDisk   = uint32(7)
)

func Reason(err error) uint32 {
	err = errors.Err(err)
	if _, ok := err.(*rpc.ErrorInfo); !ok {
		return ErrReasonNetworkError
	}
	code, _ := httputil.DetectError(err)
	if code == 801 {
		return ErrReasonDiskIoError
	} else if code == 804 {
		return ErrReasonPsectUsed
	} else if code == 811 {
		return ErrReasonNoSuchDisk
	} else if code == 802 {
		return ErrReasonChecksum
	} else if code == 810 {
		return ErrReasonSuidNotMatch
	} else if code == 812 {
		return ErrReasonBrokenDisk
	} else {
		return ErrReasonUnknown
	}
}

// ---------------------------------------------------------

func Put(
	host string, l rpc.Logger,
	psect uint64, crc32, size uint32, suid, oldSuid uint64, body io.Reader) (err error) {

	url := fmt.Sprintf("%v/put/%v/crc32/%v/suid/%v/oldsuid/%v", host, psect, crc32, suid, oldSuid)
	err = DefaultClient.CallWith(l, nil, url, "application/octet-stream", body, int(size))
	if err != nil {
		if httputil.DetectCode(err) == ErrCodeDuplicated {
			err = nil
		}
	}
	return
}

func Recycleget(
	host string, l rpc.Logger,
	psect uint64, off, size uint32, suid uint64) (body io.ReadCloser, err error) {
	url := fmt.Sprintf("%v/recycleget/%v/off/%v/n/%v/suid/%v", host, psect, off, size, suid)
	return Client{DefaultClient}.get(url, l)
}

func Repairget(
	host string, l rpc.Logger,
	psect uint64, off, size uint32, suid uint64) (body io.ReadCloser, err error) {
	url := fmt.Sprintf("%v/repairget/%v/off/%v/n/%v/suid/%v", host, psect, off, size, suid)
	return Client{DefaultClient}.get(url, l)
}

func Get(
	host string, l rpc.Logger,
	psect uint64, off, size uint32, suid uint64) (body io.ReadCloser, err error) {
	url := fmt.Sprintf("%v/get/%v/off/%v/n/%v/suid/%v", host, psect, off, size, suid)
	return Client{DefaultClient}.get(url, l)
}

func ExtendFormat(l rpc.Logger, host string) (err error) {
	url := fmt.Sprintf("%v/extend/format", host)
	err = rpc.DefaultClient.Call(l, nil, url)
	return
}

// ---------------------------------------------------------
func Reopen(l rpc.Logger, host string, idx int) (err error) {
	url := fmt.Sprintf("%v/reopen/%v", host, idx)
	err = rpc.DefaultClient.Call(l, nil, url)
	return
}

func Broken(l rpc.Logger, host string, idx int) (err error) {
	url := fmt.Sprintf("%v/broken/%v", host, idx)
	err = rpc.DefaultClient.Call(l, nil, url)
	return
}

type Client struct {
	rpc.Client
}

func (c Client) Get(
	host string, l rpc.Logger,
	psect uint64, off, size uint32, suid uint64) (body io.ReadCloser, err error) {
	url := fmt.Sprintf("%v/get/%v/off/%v/n/%v/suid/%v", host, psect, off, size, suid)
	return c.get(url, l)
}

func (c Client) get(url string, l rpc.Logger) (body io.ReadCloser, err error) {

	resp, err := c.PostEx(l, url)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}

	return resp.Body, nil
}
