package iam

import (
	"strings"

	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"qbox.us/iam/entity"
	qconf "qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

type Client struct {
	Conn *qconf.Client
}

func (r Client) GetIamQConfInfo(l rpc.Logger, accessKey string) (info *entity.QConfInfo, err error) {
	id := MakeId(accessKey)
	info = &entity.QConfInfo{}
	err = r.Conn.Get(l, info, id, 0)
	return
}

// ------------------------------------------------------------------------

const (
	groupPrefix    = "iamQConfInfo:"
	groupPrefixLen = len(groupPrefix)
)

var ErrInvalidGroup = errors.New("invalid group")

func MakeId(accessKey string) string {
	return groupPrefix + accessKey
}

func ParseId(id string) (accessKey string, err error) {
	if !strings.HasPrefix(id, groupPrefix) {
		return "", ErrInvalidGroup
	}
	accessKey = strings.TrimSpace(id[groupPrefixLen:])
	return accessKey, nil
}

// ------------------------------------------------------------------------
