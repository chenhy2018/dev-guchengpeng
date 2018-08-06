package services

import (
	"net/http"

	"github.com/qiniu/rpc.v1"
)

type OpLogService interface {
	CreateOpLog(params CreateOpLogIn) error
}

type opLogService struct {
	host      string
	client    rpc.Client
	reqLogger rpc.Logger
}

func NewOpLogService(host string, adminOauth http.RoundTripper, reqLogger rpc.Logger) OpLogService {
	return &opLogService{
		host: host,
		client: rpc.Client{
			&http.Client{Transport: adminOauth},
		},
		reqLogger: reqLogger,
	}
}
