package pfopmq

import (
	"errors"
	"strings"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/pfopmq"
	qconf "qbox.us/qconf/qconfapi"
)

type Client struct {
	Conn *qconf.Client
}

func (r Client) GetPipelineInfo(l rpc.Logger, owner, name string) (*pfopmq.Pipeline, error) {
	var info pfopmq.Pipeline
	id := MakePipelineInfoId(owner, name)
	err := r.Conn.Get(l, &info, id, 0)
	if err != nil {
		return nil, err
	}
	return &info, err
}

const groupPrefix = "pipeline:"
const groupPrefixLen = len(groupPrefix)

var ErrInvalidId = errors.New("invalid id")

func MakePipelineInfoId(owner, name string) string {
	key := owner + ":" + name
	return groupPrefix + key
}

func ParsePipelineInfoId(id string) (owner string, name string, err error) {
	if !strings.HasPrefix(id, groupPrefix) {
		return "", "", ErrInvalidId
	}
	key := id[groupPrefixLen:]
	pos := strings.Index(key, ":")
	if pos < 0 {
		return "", "", ErrInvalidId
	}
	owner = key[:pos]
	name = key[pos+1:]
	return
}
