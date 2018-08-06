package akg

import (
	"errors"
	"strings"

	"github.com/qiniu/rpc.v1"
	qconf "qbox.us/qconf/qconfapi"
	account "qiniu.com/auth/proto.v1"
)

// ------------------------------------------------------------------------

type Info account.AccessInfo

type Client struct {
	Conn *qconf.Client
}

func (r Client) Get(l rpc.Logger, accessKey string) (ret Info, err error) {

	err = r.Conn.Get(l, &ret, MakeId(accessKey), qconf.Cache_NoSuchEntry)
	return
}

// ------------------------------------------------------------------------

const groupPrefix = "ak:"
const groupPrefixLen = len(groupPrefix)

var ErrInvalidGroup = errors.New("invalid group")

func MakeId(key string) string {
	return groupPrefix + key
}

func ParseId(id string) (key string, err error) {
	if !strings.HasPrefix(id, groupPrefix) {
		return "", ErrInvalidGroup
	}
	return id[groupPrefixLen:], nil
}

// ------------------------------------------------------------------------
