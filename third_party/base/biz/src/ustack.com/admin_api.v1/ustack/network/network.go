package network

import (
	"github.com/qiniu/rpc.v2"
	"ustack.com/admin_api.v1/ustack"
)

type Client struct {
	ProjectId string
	Conn      ustack.Conn
}

func New(services ustack.Services, project string) Client {
	conn, ok := services.Find("neutron")
	if !ok {
		panic("network api not found")
	}

	return Client{
		ProjectId: project,
		Conn:      conn,
	}
}

func fakeError(err error) bool {
	if rpc.HttpCodeOf(err)/100 == 2 {
		return true
	}
	return false
}

// ------------------------------------------------------------------
// 获取 quota 设置

type Quota struct {
	Qta struct {
		Floatingip        int `json:"floatingip,omitempty"`
		Network           int `json:"network,omitempty"`
		Port              int `json:"port,omitempty"`
		PortPerVm         int `json:"port_per_vm,omitempty"`
		PPTPConnection    int `json:"pptpconnection,omitempty"`
		Router            int `json:"route,omitempty"`
		SecurityGroup     int `json:"security_group,omitempty"`
		SecurityGroupRule int `json:"security_group_rule,omitempty"`
		Subnet            int `json:"subnet,omitempty"`
		VPNUser           int `json:"vpnuser,omitempty"`
		Loadbalancer      int `json:"loadbalancer,omitempty"`
		Pool              int `json:"pool,omitempty"`
		Listener          int `json:"listener,omitempty"`
		L7policy          int `json:"l7policy,omitempty"`
	} `json:"quota"`
}

func (p *Client) GetQuota(l rpc.Logger, projectId string) (ret Quota, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/quotas/"+projectId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// ------------------------------------------------------------------
// 修改 quota 设置

func (p *Client) PutQuota(l rpc.Logger, projectId string, args Quota) (ret Quota, err error) {
	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2.0/quotas/"+projectId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}
