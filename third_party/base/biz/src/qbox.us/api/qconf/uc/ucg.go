package uc

import (
	"github.com/qiniu/rpc.v1"
	"qbox.us/admin_api/uc"
	qconf "qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

type Info struct {
	Items []uc.GroupItem `bson:"items" json:"items"`
}

// ------------------------------------------------------------------------

type Client struct {
	Conn *qconf.Client
}

func (r Client) Get(l rpc.Logger, grp, name string) (val string, err error) {

	var ret struct {
		Val string `bson:"v"`
	}
	err = r.Conn.Get(l, &ret, "uc:"+grp+":"+name, 0)
	val = ret.Val
	return
}

// ------------------------------------------------------------------------
