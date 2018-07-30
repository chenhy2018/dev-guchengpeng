package api

import (
	"io"

	pfdcfg "qbox.us/pfdcfg/api"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
)

type nullPfd struct {
}

func (self nullPfd) Get(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {
	return nil, 0, httputil.NewError(612, "Null pfd")
}

func (self nullPfd) GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error) {
	return 0, httputil.NewError(612, "Null pfd")
}

type nullEbd struct {
}

func (self nullEbd) Get(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {
	return nil, 0, httputil.NewError(612, "Null ebd")
}

func (self nullEbd) GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error) {
	return 0, httputil.NewError(612, "Null pfd")
}

type ebdStater struct {
}

func (self ebdStater) State(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error) {
	return 0, true, nil
}

func (self ebdStater) StateWithGroup(l rpc.Logger, egid string) (group string, dgid uint32, isECed bool, err error) {
	return "", 0, true, nil
}

func (self ebdStater) ForceUpdate(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error) {
	return 0, true, nil
}

var (
	NullPfd   = nullPfd{}
	NullEbd   = nullEbd{}
	EbdStater = ebdStater{}
)
