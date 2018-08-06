package uidg

import (
	"errors"
	"strconv"
	"strings"

	"github.com/qiniu/rpc.v1"
	qconf "qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

type Info struct {
	Utype uint32 `bson:"utype" json:"utype"`
}

// ------------------------------------------------------------------------

type Client struct {
	Conn *qconf.Client
}

func (r Client) Get(l rpc.Logger, uid uint32) (ret Info, err error) {

	err = r.Conn.Get(l, &ret, MakeId(uid), 0)
	return
}

func (r Client) GetUtype(l rpc.Logger, uid uint32) (utype uint32, err error) {

	var ret Info
	err = r.Conn.Get(l, &ret, MakeId(uid), 0)
	utype = ret.Utype
	return
}

// ------------------------------------------------------------------------

const groupPrefix = "uidg:"
const groupPrefixLen = len(groupPrefix)

var ErrInvalidGroup = errors.New("invalid group")

func MakeId(uid uint32) string {
	return groupPrefix + strconv.FormatUint(uint64(uid), 36)
}

func ParseId(id string) (uid uint32, err error) {
	if !strings.HasPrefix(id, groupPrefix) {
		return 0, ErrInvalidGroup
	}
	v, err := strconv.ParseUint(id[groupPrefixLen:], 36, 32)
	return uint32(v), err
}

// ------------------------------------------------------------------------
