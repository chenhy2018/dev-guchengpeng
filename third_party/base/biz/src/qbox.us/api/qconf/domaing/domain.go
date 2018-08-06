package domaing

import (
	"errors"
	"strings"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/one/domain"
	qconf "qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

type Info struct {
	Phy              string `json:"phy" bson:"phy"`
	Tbl              string `json:"tbl" bson:"tbl"`
	Uid              uint32 `json:"uid" bson:"uid"`
	Itbl             uint32 `json:"itbl" bson:"itbl"`
	Refresh          bool   `json:"refresh" bson:"refresh"`
	Global           bool   `json:"global" bson:"global"`
	Domain           string `json:"domain" bson:"domain"`
	domain.AntiLeech `json:"antileech,omitempty" bson:"antileech,omitempty"`
}

type Getter interface {
	Get(l rpc.Logger, domain string) (ret Info, err error)
}

// ------------------------------------------------------------------------

type Client struct {
	Conn *qconf.Client
}

func (r Client) Get(l rpc.Logger, domain string) (ret Info, err error) {

	err = r.Conn.Get(l, &ret, MakeId(domain), qconf.Cache_NoSuchEntry)
	return
}

// ------------------------------------------------------------------------
const groupPrefix = "domain:"
const groupPrefixLen = len(groupPrefix)

var ErrInvalidGroup = errors.New("invalid group")

func MakeId(domain string) string {
	return groupPrefix + domain
}

func ParseId(id string) (domain string, err error) {
	if !strings.HasPrefix(id, groupPrefix) {
		return "", ErrInvalidGroup
	}
	return id[groupPrefixLen:], nil
}

// ------------------------------------------------------------------------
