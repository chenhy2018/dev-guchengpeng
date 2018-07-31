package stgapi

import (
	"errors"
	"strconv"
	"strings"
	"syscall"

	"github.com/qiniu/rpc.v1"
	"qbox.us/qconf/qconfapi"
)

type Client struct {
	Conn *qconfapi.Client
}

type hostRet struct {
	Hosts [2]string `bson:"hosts"`
}

func (r Client) Host(l rpc.Logger, guid string, diskId uint32) (host string, err error) {

	var ret hostRet
	err = r.Conn.Get(l, &ret, MakeDiskId(guid, diskId), 0)
	if err != nil {
		return
	}
	if ret.Hosts[0] == "" && ret.Hosts[1] == "" {
		err = syscall.ENOENT
		return
	}
	return ret.Hosts[0], err
}

func (r Client) EcHost(l rpc.Logger, guid string, diskId uint32) (echost string, err error) {

	var ret hostRet
	err = r.Conn.Get(l, &ret, MakeDiskId(guid, diskId), 0)
	if err != nil {
		return
	}
	if ret.Hosts[0] == "" && ret.Hosts[1] == "" {
		err = syscall.ENOENT
		return
	}
	return ret.Hosts[1], err
}

const groupPrefix = "edisk:"
const groupPrefixLen = len(groupPrefix)

var ErrInvalidGroup = errors.New("invalid group")

func MakeDiskId(guid string, diskId uint32) string {
	return groupPrefix + guid + ":" + strconv.FormatUint(uint64(diskId), 36)
}

func ParseDiskId(s string) (guid string, diskId uint32, err error) {
	if !strings.HasPrefix(s, groupPrefix) {
		err = ErrInvalidGroup
		return
	}
	q := strings.Split(s[groupPrefixLen:], ":")
	guid = q[0]
	diskId64, err := strconv.ParseUint(q[1], 36, 32)
	return guid, uint32(diskId64), err
}
