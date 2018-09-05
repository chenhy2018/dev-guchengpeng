package biz

import (
	"github.com/qiniu/rpc.v1"
)

type BizService struct {
	host string
	rpc  *rpc.Client
}

func NewBizService(host string, client *rpc.Client) *BizService {
	return &BizService{host, client}
}
