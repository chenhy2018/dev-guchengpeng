package wallet_stat

import (
	"net/url"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleStatStatus struct {
	Host   string
	Client *rpc.Client
}

func NewHandleStatStatus(host string, client *rpc.Client) *HandleStatStatus {
	return &HandleStatStatus{host, client}
}

type ModelStatStatus struct {
	Job  Job    `json:"job"`
	Date Second `json:"date"`
}

type ReqStatStatus struct {
	Job Job `json:"job"`
}

func (r HandleStatStatus) Get(logger rpc.Logger, req ReqStatStatus) (resp ModelStatStatus, err error) {
	value := url.Values{}
	value.Add("job", req.Job.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v1/statstatus/get?"+value.Encode())
	return
}
