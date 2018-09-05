package product

import (
	"github.com/qiniu/rpc.v1"
)

type HandleBase struct {
	Host   string
	Client *rpc.Client
}

func NewHandleBase(host string, client *rpc.Client) *HandleBase {
	return &HandleBase{host, client}
}
