package appg

import (
	"errors"
	"strconv"
	"strings"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/one/access"
	qconf "qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

type Client struct {
	Conn *qconf.Client
}

func (r Client) Get(l rpc.Logger, app string, uid uint32) (ret access.AppInfo, err error) {

	err = r.Conn.Get(l, &ret, MakeId(app, uid), 0)
	return
}

func (r Client) GetAkSk(l rpc.Logger, uid uint32) (ak, sk string, err error) {

	info, err := r.Get(l, "default", uid)
	if err != nil {
		return
	}

	ak = info.Key
	sk = info.Secret
	return
}

// ------------------------------------------------------------------------

const groupPrefix = "app:"
const groupPrefixLen = len(groupPrefix)

var (
	ErrInvalidGroup = errors.New("invalid group")
	ErrInvalidId    = errors.New("invalid id")
)

func MakeId(app string, uid uint32) string {
	return groupPrefix + strconv.FormatUint(uint64(uid), 36) + ":" + app
}

func ParseId(id string) (app string, uid uint32, err error) {
	if !strings.HasPrefix(id, groupPrefix) {
		err = ErrInvalidGroup
		return
	}
	id = id[groupPrefixLen:]
	idx := strings.Index(id, ":")
	if idx == -1 {
		err = ErrInvalidId
		return
	}

	app = id[idx+1:]

	uid6, err := strconv.ParseUint(id[:idx], 36, 64)
	if err != nil {
		err = ErrInvalidId
		return
	}
	uid = uint32(uid6)
	return
}

// ------------------------------------------------------------------------
