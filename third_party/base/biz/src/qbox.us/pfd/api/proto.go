package api

import (
	"io"

	"github.com/qiniu/rpc.v1"

	cfgapi "qbox.us/pfdcfg/api"
)

type PfdPutGetter interface {
	GetType(l rpc.Logger, fh []byte) (cfgapi.DiskType, error)
	Put(l rpc.Logger, r io.Reader, fsize int64) ([]byte, error)
	Get(l rpc.Logger, fh []byte, from, to int64) (io.ReadCloser, int64, error)
}

type DgInfoer interface {
	ListDgs(l rpc.Logger, guid string) ([]*cfgapi.DiskGroupInfo, error)
	GetDGInfo(l rpc.Logger, guid string, dgid uint32) (dgInfo *cfgapi.DiskGroupInfo, err error)
}
