package stater

import (
	"errors"
	"strings"

	"time"

	"github.com/qiniu/rpc.v1"
	"qbox.us/qconf/qconfapi"
)

type Stater interface {
	State(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error)
	StateWithGroup(l rpc.Logger, egid string) (group string, dgid uint32, isECed bool, err error)
	ForceUpdate(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error)
}

// Stater 并不是一个标准的go的interface
// 一旦有人想加StateWithGroup， 就一定还会想加其他方法比如StateWithECtime，会发现接口永远不够用。
// 直接修改Stater是一个牵连甚广的工程，EntryStater保证不会有人再继续增加新的方法了。
type EntryStater interface {
	Stater
	StateEntry(l rpc.Logger, egid string) (entry Entry, err error)
}

const groupPrefix = "egid:"
const groupPrefixLen = len(groupPrefix)

func MakeId(egid string) (id string) {
	return groupPrefix + egid
}

func ParseId(id string) (egid string, err error) {
	if !strings.HasPrefix(id, groupPrefix) {
		return "", errors.New("invalid group")
	}
	return id[groupPrefixLen:], nil
}

// ===========================================================

type Entry struct {
	Group  string    `bson:"group"`
	Dgid   uint32    `bson:"dgid"`
	EC     bool      `bson:"ec"`
	ECTime time.Time `bson:"ectime,omitempty"`
	Ecing  int32     `bson:"ecing"`
}

func NewGidStater(qconf *qconfapi.Config) *GidStater {
	cli := qconfapi.New(qconf)
	return &GidStater{qconfcli: cli}
}

type GidStater struct {
	qconfcli *qconfapi.Client
}

func (self *GidStater) State(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error) {

	id := MakeId(egid)
	var ret Entry
	err = self.qconfcli.Get(l, &ret, id, qconfapi.Cache_Normal)
	if err != nil {
		return
	}
	return ret.Dgid, ret.EC, nil
}

func (self *GidStater) StateEntry(l rpc.Logger, egid string) (entry Entry, err error) {
	id := MakeId(egid)
	err = self.qconfcli.Get(l, &entry, id, qconfapi.Cache_Normal)
	if err != nil {
		return
	}
	return
}

func (self *GidStater) StateWithECtime(l rpc.Logger, egid string) (dgid uint32, isECed bool, ecTime time.Time, err error) {

	id := MakeId(egid)
	var ret Entry
	err = self.qconfcli.Get(l, &ret, id, qconfapi.Cache_Normal)
	if err != nil {
		return
	}
	return ret.Dgid, ret.EC, ret.ECTime, nil
}

func (self *GidStater) StateWithGroup(l rpc.Logger, egid string) (group string, dgid uint32, isECed bool, err error) {
	id := MakeId(egid)
	var ret Entry
	err = self.qconfcli.Get(l, &ret, id, qconfapi.Cache_Normal)
	if err != nil {
		return
	}
	return ret.Group, ret.Dgid, ret.EC, nil
}

func (self *GidStater) ForceUpdate(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error) {

	id := MakeId(egid)
	var ret Entry
	err = self.qconfcli.GetFromMaster(l, &ret, id, qconfapi.Cache_Normal)
	if err != nil {
		return
	}
	return ret.Dgid, ret.EC, nil
}
