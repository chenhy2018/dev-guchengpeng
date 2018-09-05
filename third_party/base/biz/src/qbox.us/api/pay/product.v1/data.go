package product

import (
	"net/url"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleDataSync struct {
	Host   string
	Client *rpc.Client
}

func NewHandleDataSync(host string, client *rpc.Client) *HandleDataSync {
	return &HandleDataSync{host, client}
}

func (r HandleDataSync) Set(logger rpc.Logger, params ReqDataSync) (err error) {
	value := url.Values{}
	value.Add("product", params.Product.ToString())
	value.Add("zone", params.Zone.String())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v1/data_sync/set", map[string][]string(value))
	return
}

func (r HandleDataSync) Status(logger rpc.Logger, params ReqDataSync) (data RespDataSyncStatus, err error) {
	value := url.Values{}
	value.Add("product", params.Product.ToString())
	value.Add("zone", params.Zone.String())
	err = r.Client.Call(logger, &data, r.Host+"/v1/data_sync/status?"+value.Encode())
	return
}

func (r HandleDataSync) List(logger rpc.Logger) (data []RespDataSyncStatus, err error) {
	err = r.Client.Call(logger, &data, r.Host+"/v1/data_sync/list")
	return
}
