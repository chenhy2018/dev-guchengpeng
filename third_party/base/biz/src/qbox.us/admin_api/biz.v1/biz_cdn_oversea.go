package biz

import (
	"github.com/qiniu/rpc.v1"
)

type CdnOverseaData struct {
	Uid    uint32 `json:"uid,string"` // 用户
	Bucket string `json:"bucket"`     // bucket
	Domain string `json:"domain"`     // 计费域名
}

// 返回状态为“计费”的域名
func (s *BizService) CdnOverseaListActive(l rpc.Logger) (res []*CdnOverseaData, err error) {
	err = s.rpc.CallWithJson(l, &res, s.host+"/admin/cdn/oversea/active", nil)
	return
}
