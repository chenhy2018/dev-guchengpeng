package ufop

import (
	"net/http"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/brpc"
)

type BClient struct {
	host string
	Conn brpc.Client
}

func NewBClient(host string, t http.RoundTripper) *BClient {
	conn := brpc.Client{&http.Client{Transport: t}}
	return &BClient{host, conn}
}

type AclEntry struct {
	Ufop    string   `bson:"ufop"`
	AclMode byte     `bson:"acl_mode"`
	AclList []uint32 `bson:"acl_list"`
}

type ListallRet struct {
	Entries []AclEntry `bson:"entries"`
}

func (b *BClient) Listall(l rpc.Logger) (ret ListallRet, err error) {
	err = b.Conn.Call(l, &ret, b.host+"/listall")
	return
}

type ListUfopsRet struct {
	Ufops []string `bson:"ufops"`
}

func (b *BClient) ListUfops(l rpc.Logger) (ret ListUfopsRet, err error) {
	err = b.Conn.Call(l, &ret, b.host+"/list/ufops")
	return
}
