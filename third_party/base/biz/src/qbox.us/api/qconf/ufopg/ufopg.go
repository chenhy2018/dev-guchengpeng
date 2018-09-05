package ufopg

import (
	"github.com/qiniu/rpc.v1"
	"qbox.us/admin_api/one/ufop"
	qconf "qbox.us/qconf/qconfapi"
)

type AclEntry ufop.AclEntry

type Client struct {
	Conn *qconf.Client
}

//---------------------------------------------------------

type ListRet struct {
	Entries []AclEntry `bson:"entries"`
}

func (r Client) List(l rpc.Logger) (ret ListRet, err error) {

	err = r.Conn.Get(l, &ret, "ufopg:all", 0)
	return
}

//---------------------------------------------------------

type ListNameRet struct {
	Ufops []string `bson:"ufops"`
}

func (r Client) ListName(l rpc.Logger) (ret ListNameRet, err error) {

	err = r.Conn.Get(l, &ret, "ufopg:name", 0)
	return
}

//---------------------------------------------------------

const groupPrefix = "ufopg:"

// Get id list for refreshing cache.
func MakeId() []string {
	return []string{
		groupPrefix + "all",
		groupPrefix + "name",
	}
}
