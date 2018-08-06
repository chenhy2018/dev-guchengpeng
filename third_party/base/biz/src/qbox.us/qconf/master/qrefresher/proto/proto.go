package proto

import (
	"github.com/qiniu/rpc.v1"
)

// ------------------------------------------------------------------------

type Refresher interface {
	Refresh(l rpc.Logger, id string) (err error)
}

type MultiRefresher interface {
	MultiRefresh(l rpc.Logger, id ...string) (err error)
}

// ------------------------------------------------------------------------

type nilRefresher struct{}

func (r nilRefresher) Refresh(l rpc.Logger, id string) (err error) {
	return nil
}

func (r nilRefresher) MultiRefresh(l rpc.Logger, id ...string) (err error) {
	return nil
}

var NilRefresher nilRefresher

// ------------------------------------------------------------------------
