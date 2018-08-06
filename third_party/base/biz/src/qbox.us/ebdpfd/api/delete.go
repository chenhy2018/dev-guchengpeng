package api

import (
	"github.com/qiniu/rpc.v1"

	ebdtypes "qbox.us/ebd/api/types"
	"qbox.us/errors"
	"qbox.us/pfd/api/types"
	"qbox.us/pfdtracker/stater"
)

var ErrEgidEcing = errors.New("egid is ECing")

type Deleter interface {
	Delete(l rpc.Logger, fh []byte) (err error)
}

type DeleterChooser interface {
	Choose(group string) Deleter
}

type Delete struct {
	s   stater.EntryStater
	pfd Deleter
	ebd DeleterChooser
}

func NewDeleter(s stater.EntryStater, pfd Deleter, ebd DeleterChooser) *Delete {
	return &Delete{s: s, pfd: pfd, ebd: ebd}
}

func (self *Delete) Delete(l rpc.Logger, fh []byte) (err error) {

	fhi, err := ebdtypes.DecodeFh(fh)
	if err != nil {
		return
	}
	egid := types.EncodeGid(fhi.Gid)
	entry, err := self.s.StateEntry(l, egid)
	if err != nil {
		return
	}
	if !entry.EC {
		if entry.Ecing == 0 {
			return self.pfd.Delete(l, fh)
		} else {
			return ErrEgidEcing
		}
	}
	err = self.ebd.Choose(entry.Group).Delete(l, fh)
	return
}
