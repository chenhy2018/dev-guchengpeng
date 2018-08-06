package v3

import (
	"encoding/json"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/pay/pay"
)

type HandleResourceGroup struct {
	Host   string
	Client *rpc.Client
}

func NewHandleResourceGroup(host string, client *rpc.Client) *HandleResourceGroup {
	return &HandleResourceGroup{host, client}
}

type ReqSetResourceGroup struct {
	Id        string            `json:"id"`
	Type      int               `json:"type"`
	RawGroups []json.RawMessage `json:"groups"`
}

func (r HandleResourceGroup) Set(logger rpc.Logger, req ReqSetResourceGroup) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/resourcegroup/set", req)
	return
}

type ReqCreateResourceGroup struct {
	BaseId    string            `json:"base_id"`
	Item      pay.Item          `json:"item"`
	Type      int               `json:"type"`
	RawGroups []json.RawMessage `json:"groups"`
}

type RespCreateResourceGroup struct {
	ResourceGroupId string `json:"resource_group_id"`
}

func (r HandleResourceGroup) Create(logger rpc.Logger, req ReqCreateResourceGroup) (resp RespCreateResourceGroup, err error) {
	err = r.Client.CallWithJson(logger, &resp, r.Host+"/v3/resourcegroup/create", req)
	return
}

type ReqCloneResourceGroup struct {
	Id string `json:"id"`
}

func (r HandleResourceGroup) Clone(logger rpc.Logger, req ReqCloneResourceGroup) (resp RespCreateResourceGroup, err error) {
	err = r.Client.CallWithJson(logger, &resp, r.Host+"/v3/resourcegroup/clone", req)
	return
}
