package gaea

import (
	"net/http"

	"github.com/qiniu/rpc.v1"

	"qbox.us/biz/api/gaea"
	"qbox.us/biz/api/gaea/enums"
	"qbox.us/biz/utils.v2/log"
)

type OpLogService interface {
	Create(preq *http.Request, uid uint32, ip, userAgent string, opType enums.OpType, op enums.Op,
		extra, bucket string) (err error)
	Query(uid uint32, cluster enums.OpTypeCluster, bucket string, offset, limit int) ([]*gaea.OpLog, error)
}

type oplogService struct {
	logger    log.ReqLogger
	rpcLogger rpc.Logger

	impl *gaea.OpLogService
}

var _ OpLogService = new(oplogService)

func NewOpLogService(host string, t http.RoundTripper, logger log.ReqLogger) OpLogService {
	return &oplogService{
		logger:    logger,
		rpcLogger: log.NewRpcWrapper(logger),
		impl:      gaea.NewOpLogService(host, t),
	}
}

func (s *oplogService) Create(preq *http.Request, uid uint32, ip, userAgent string, opType enums.OpType, op enums.Op,
	extra, bucket string) (err error) {
	err = s.impl.Create(preq, s.rpcLogger, uid, ip, userAgent, opType, op, extra, bucket)
	return
}

func (s *oplogService) Query(uid uint32, cluster enums.OpTypeCluster, bucket string, offset, limit int) (logs []*gaea.OpLog, err error) {
	logs, err = s.impl.Query(s.rpcLogger, uid, cluster, bucket, offset, limit)
	return
}
