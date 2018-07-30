package cc

import (
	"github.com/qiniu/rpc.v2"
	"strconv"
)

type NetworkInfo struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	RateLimit int    `json:"rate_limit"`
	Shared    bool   `json:"shared"`
	External  bool   `json:"external"`
}

type ListNetworksRet []NetworkInfo

func (r *Service) ListNetworks(l rpc.Logger) (ret ListNetworksRet, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/networks")
	return
}

// -------------------------------------------------------------------------------

type SubnetInfo struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	NetworkId string `json:"network_id"`
	Shared    bool   `json:"shared"`
	Cidr      string `json:"cidr"`
}

type ListSubnetRet []SubnetInfo

func (r *Service) ListSubnets(l rpc.Logger) (ret ListSubnetRet, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/subnets")
	return
}

// -------------------------------------------------------------------------------

type AddrPair struct {
	IpAddr string `json:"ip_address"`
}

type Port struct {
	Status              string     `json:"status"`
	Name                string     `json:"name"`
	AllowedAddrPairs    []AddrPair `json:"allowed_address_pairs"`
	DisableAntiSpoofing bool       `json:"binding:disable_anti_spoofing"`
	NetworkId           string     `json:"network_id"`
	TenantId            string     `json:"tenant_id"`
	CreatedAt           string     `json:"created_at"`
	MacAddress          string     `json:"mac_address"`
	Id                  string     `json:"id"`
	FixedIps            []FixedIp  `json:"fixed_ips"`
	SecGrps             []string   `json:"security_groups"`
	DeviceId            string     `json:"device_id"`
	DeviceOwner         string     `json:"device_owner"`
}

type ListPortsRet struct {
	Ports []Port `json:"ports"`
}

// -------------------------------------------------------------------------------
// 列出所有网卡

func (r *Service) ListPorts(l rpc.Logger) (ret ListPortsRet, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/ports")
	return
}

// -------------------------------------------------------------------------------
// 创建虚拟网卡

func (r *Service) CreatePort(l rpc.Logger, name, networkId string) (ret Port, err error) {
	params := map[string][]string{
		"name":       {name},
		"network_id": {networkId},
	}
	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/ports", params)
	return
}

// -------------------------------------------------------------------------------
// 删除虚拟网卡

func (r *Service) DeletePort(l rpc.Logger, portId string) (err error) {
	err = r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/ports/"+portId)
	return
}

// -------------------------------------------------------------------------------
// 修改网卡名称

func (r *Service) PortRename(l rpc.Logger, portId, name string) (ret Port, err error) {
	params := map[string][]string{
		"name": {name},
	}
	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/ports/"+portId+"/rename", params)
	return
}

// -------------------------------------------------------------------------------
// 启用或关闭虚拟网卡的安全限制

func (r *Service) PortAntiSpoofing(l rpc.Logger, portId string, disable bool) (ret Port, err error) {
	params := map[string][]string{
		"disable": {strconv.FormatBool(disable)},
	}
	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/ports/"+portId+"/antispoofing", params)
	return
}

// -------------------------------------------------------------------------------
// 修改网卡 allowed_adress_pairs

func (r *Service) SetPortAllowedAddrPairs(l rpc.Logger, portId, ipAddr string) (ret Port, err error) {
	params := map[string][]string{
		"ipaddress": {ipAddr},
	}
	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/ports/"+portId+"/allowedaddrpairs", params)
	return
}

// -------------------------------------------------------------------------------

func (r *Service) SetPortSecGrp(l rpc.Logger, portId, secGrpId string) (ret Port, err error) {
	params := map[string][]string{
		"security_group_id": {secGrpId},
	}
	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/ports/"+portId, params)
	return
}
