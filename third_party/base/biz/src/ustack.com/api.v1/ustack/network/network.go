// neuturn
package network

import (
	"github.com/qiniu/rpc.v2"
	"ustack.com/api.v1/ustack"
)

// --------------------------------------------------

type Client struct {
	ProjectId string
	Conn      ustack.Conn
}

func New(services ustack.Services, project string) Client {

	conn, ok := services.Find("network")
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

// --------------------------------------------------
// 列出所有可用的网络

type Network struct {
	Status       string   `json:"status"`
	Subnets      []string `json:"subnets"`
	Name         string   `json:"name"`
	AdminStateUp bool     `json:"admin_state_up"`
	TenantId     string   `json:"tenant_id"`
	CreatedAt    string   `json:"created_at"`
	RateLimit    int      `json:"uos:rate_limit"`
	External     bool     `json:"router:external"`
	Shared       bool     `json:"shared"`
	Id           string   `json:"id"`
}

type ListNetworksRet struct {
	Networks []Network `json:"networks"`
}

func (p Client) ListNetworks(l rpc.Logger) (ret *ListNetworksRet, err error) {

	ret = &ListNetworksRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2.0/networks")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 列出所有子网

type AllocationPool struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Subnet struct {
	Id              string           `json:"id"`
	Name            string           `json:"name"`
	EnableDhcp      bool             `json:"enable_dhcp"`
	NetworkId       string           `json:"network_id"`
	TenantId        string           `json:"tenant_id"`
	CreatedAt       string           `json:"created_at"`
	DnsNameservers  []string         `json:"dns_nameservers"`
	Ipv6RaMode      bool             `json:"ipv6_ra_mode"`
	AllocPools      []AllocationPool `json:"allocation_pools"`
	GatewayIp       string           `json:"gateway_ip"`
	Shared          bool             `json:"shared"`
	IpVersion       int              `json:"ip_version"`
	Cidr            string           `json:"cidr"`
	Ipv6AddressMode bool             `json:"ipv6_address_mode"`
}

type ListSubnetsRet struct {
	Subnets []Subnet `json:"subnets"`
}

func (p Client) ListSubnets(l rpc.Logger) (ret *ListSubnetsRet, err error) {
	ret = &ListSubnetsRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2.0/subnets")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看所有虚拟网卡

type FixedIp struct {
	SubnetId  string `json:"subnet_id"`
	IpAddress string `json:"ip_address"`
}

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

type PortRet struct {
	Port Port `json:"port"`
}

type ListPortsRet struct {
	Ports []Port `json:"ports"`
}

func (p Client) ListPorts(l rpc.Logger) (ret ListPortsRet, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/ports")
	if err != nil && fakeError(err) {
		err = nil
		return
	}
	return
}

// --------------------------------------------------
// 创建虚拟网卡

func (p Client) CreatePort(l rpc.Logger, name, networkId string) (ret Port, err error) {
	type createPortArgs struct {
		Port struct {
			Name      string `json:"name"`
			NetworkId string `json:"network_id"`
		} `json:"port"`
	}

	var args createPortArgs
	args.Port.Name = name
	args.Port.NetworkId = networkId
	var uret PortRet
	err = p.Conn.CallWithJson(l, &uret, "POST", "/v2.0/ports", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	ret = uret.Port
	return
}

// --------------------------------------------------
// 删除虚拟网卡

func (p Client) DeletePort(l rpc.Logger, portId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2.0/ports/"+portId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 修改网卡名称

func (p Client) UpdatePortName(l rpc.Logger, portId, name string) (ret Port, err error) {
	type portNameArgs struct {
		Port struct {
			Name string `json:"name"`
		} `json:"port"`
	}
	var args portNameArgs
	args.Port.Name = name
	var uret PortRet
	err = p.Conn.CallWithJson(l, &uret, "PUT", "/v2.0/ports/"+portId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	ret = uret.Port
	return
}

// --------------------------------------------------
// 启用或关闭虚拟网卡的安全限制

func (p Client) PortAntiSpoofing(l rpc.Logger, portId string, disable bool) (ret Port, err error) {
	type antiSpoofingArgs struct {
		Port struct {
			DisableAntiSpoofing bool `json:"binding:disable_anti_spoofing"`
		} `json:"port"`
	}

	var args antiSpoofingArgs
	args.Port.DisableAntiSpoofing = disable
	var uret PortRet
	err = p.Conn.CallWithJson(l, &uret, "PUT", "/v2.0/ports/"+portId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	ret = uret.Port
	return
}

// --------------------------------------------------
// 修改虚拟网卡的安全组

func (p Client) SetPortSecGrp(l rpc.Logger, portId, secGrpId string) (ret Port, err error) {
	type setPortSecGrpArgs struct {
		Port struct {
			SecGrp []string `json:"security_groups"`
		} `json:"port"`
	}

	type setPortSecGrpRet struct {
		Port Port `json:"port"`
	}

	args := setPortSecGrpArgs{}
	args.Port.SecGrp = []string{secGrpId}
	uret := setPortSecGrpRet{}
	err = p.Conn.CallWithJson(l, &uret, "PUT", "/v2.0/ports/"+portId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	ret = uret.Port
	return
}

// --------------------------------------------------
// 修改网卡 allowed_address_pairs

func (p Client) SetPortAllowedAddressPairs(l rpc.Logger, portId, ipAddr string) (ret Port, err error) {
	type addressPairArgs struct {
		Port struct {
			AllowedAddrPairs []AddrPair `json:"allowed_address_pairs"`
		} `json:"port"`
	}
	args := addressPairArgs{}
	args.Port.AllowedAddrPairs = append(args.Port.AllowedAddrPairs, AddrPair{
		IpAddr: ipAddr,
	})
	var uret PortRet
	err = p.Conn.CallWithJson(l, &uret, "PUT", "/v2.0/ports/"+portId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	ret = uret.Port
	return
}

// --------------------------------------------------
// 创建公网 IP

type Floatingip struct {
	Id                string `json:"id"`
	Name              string `json:"uos:name"`
	RateLimit         int    `json:"rate_limit"`
	RouterId          string `json:"router_id"`
	Status            string `json:"status"`
	TenantId          string `json:"tenant_id"`
	CreatedAt         string `json:"created_at"`
	FloatingNetworkId string `json:"floating_network_id"`
	FixedIpAddress    string `json:"fixed_ip_address"`
	FloatingipAddress string `json:"floating_ip_address"`
	PortId            string `json:"port_id"`
	Provider          string `json:"uos:service_provider"`
}

type FloatingipRet struct {
	Floatingip Floatingip `json:"floatingip"`
}

func (p Client) CreateFloatingip(l rpc.Logger, networkId string,
	rateLimit int, name, provider string) (ret *FloatingipRet, err error) {

	type createFipArgs struct {
		Floatingip struct {
			FloatingNetwordId string `json:"floating_network_id"`
			RateLimit         int    `json:"rate_limit"`
			Name              string `json:"uos:name"`
			Provider          string `json:"uos:service_provider"` // CHINATELECOM or CHINAUNICOM
		} `json:"floatingip"`
	}

	var args createFipArgs
	args.Floatingip.FloatingNetwordId = networkId
	args.Floatingip.RateLimit = rateLimit
	args.Floatingip.Name = name
	args.Floatingip.Provider = provider

	ret = &FloatingipRet{}
	err = p.Conn.CallWithJson(l, ret, "POST", "/v2.0/floatingips", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看单个 IP 详情

func (p Client) FloatingipInfo(l rpc.Logger, id string) (ret *FloatingipRet, err error) {

	ret = &FloatingipRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2.0/floatingips/"+id)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 绑定 IP 到网卡

func (p Client) BindFloatingip(l rpc.Logger, ipId, portId string) (ret *FloatingipRet, err error) {

	type bindArgs struct {
		Floatingip struct {
			PortId string `json:"port_id"`
		} `json:"floatingip"`
	}

	var args bindArgs
	args.Floatingip.PortId = portId

	ret = &FloatingipRet{}
	err = p.Conn.CallWithJson(l, ret, "PUT", "/v2.0/floatingips/"+ipId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 解绑 IP 到网卡

func (p Client) UnbindFloatingip(l rpc.Logger, ipId string) (ret *FloatingipRet, err error) {

	type unbindArgs struct {
		Floatingip struct {
			PortId interface{} `json:"port_id"`
		} `json:"floatingip"`
	}

	var args unbindArgs
	args.Floatingip.PortId = nil

	ret = &FloatingipRet{}
	err = p.Conn.CallWithJson(l, ret, "PUT", "/v2.0/floatingips/"+ipId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除公网 IP

func (p Client) DeleteFloatingip(l rpc.Logger, id string) (err error) {

	err = p.Conn.Call(l, nil, "DELETE", "/v2.0/floatingips/"+id)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 列出所有公网 IP

type ListFloatingipsRet struct {
	Floatingips []Floatingip `json:"floatingips"`
}

func (p Client) ListFloatingips(l rpc.Logger) (ret *ListFloatingipsRet, err error) {

	ret = &ListFloatingipsRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2.0/floatingips")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 修改公网IP带宽

func (p Client) AdjustFipRatelimit(l rpc.Logger, fipId string, rateLimit int) (ret *Floatingip, err error) {

	type AdjustArgs struct {
		Ratelimit int `json:"rate_limit"`
	}

	args := AdjustArgs{}
	args.Ratelimit = rateLimit
	ret = &Floatingip{}
	err = p.Conn.CallWithJson(l, ret, "PUT", "/v2.0/uos_resources/"+fipId+"/update_floatingip_ratelimit", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看所有安全组

type SecurityRule struct {
	RemoteGroupId   string `json:"remote_group_id,omitempty"`
	Direction       string `json:"direction,omitempty"`
	RemoteIpPrefix  string `json:"remote_ip_prefix,omitempty"`
	Protocol        string `json:"protocol,omitempty"`
	PortRangeMax    int    `json:"port_range_max,omitempty"`
	PortRangeMin    int    `json:"port_range_min,omitempty"`
	Ethertype       string `json:"ethertype,omitempty"`
	Id              string `json:"id,omitempty"`
	SecurityGroupId string `json:"security_group_id,omitempty"`
}

type SecurityGroup struct {
	Description string         `json:"description"`
	Rules       []SecurityRule `json:"security_group_rules"`
	Name        string         `json:"name"`
	Id          string         `json:"id"`
	CreatedAt   string         `json:"created_at"`
}

type SecurityGroupRet struct {
	SecGrp SecurityGroup `json:"security_group"`
}

type ListSecGrpsRet struct {
	SecGrps []SecurityGroup `json:"security_groups"`
}

func (p Client) ListSecurityGroups(l rpc.Logger) (ret ListSecGrpsRet, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/security-groups")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 创建安全组

func (p Client) CreateSecurityGroup(l rpc.Logger, name, desc string) (ret SecurityGroup, err error) {
	type createSecGrpsArgs struct {
		SecGroup struct {
			Desc string `json:"description"`
			Name string `json:"name"`
		} `json:"security_group"`
	}

	args := createSecGrpsArgs{}
	args.SecGroup.Desc = desc
	args.SecGroup.Name = name
	uret := SecurityGroupRet{}

	err = p.Conn.CallWithJson(l, &uret, "POST", "/v2.0/security-groups", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	ret = uret.SecGrp
	return
}

// --------------------------------------------------
// 查看安全组

func (p Client) SecurityGroupInfo(l rpc.Logger, secGrpId string) (ret SecurityGroup, err error) {
	uret := SecurityGroupRet{}
	err = p.Conn.Call(l, &uret, "GET", "/v2.0/security-groups/"+secGrpId)
	if err != nil && fakeError(err) {
		err = nil
	}
	ret = uret.SecGrp
	return
}

// --------------------------------------------------
// 删除安全组

func (p Client) DeleteSecurityGroup(l rpc.Logger, secGrpId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2.0/security-groups/"+secGrpId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 创建安全组规则

func (p Client) CreateSecurityRule(l rpc.Logger, args SecurityRule) (ret SecurityRule, err error) {
	type secGrpRule struct {
		SecRule SecurityRule `json:"security_group_rule"`
	}

	params := secGrpRule{}
	params.SecRule = args
	uret := secGrpRule{}
	err = p.Conn.CallWithJson(l, &uret, "POST", "/v2.0/security-group-rules", params)
	if err != nil && fakeError(err) {
		err = nil
	}
	ret = uret.SecRule
	return
}

// --------------------------------------------------
// 查看安全组规则

func (p Client) SecurityRuleInfo(l rpc.Logger, secRuleId string) (ret SecurityRule, err error) {
	type secGrpRule struct {
		SecRule SecurityRule `json:"security_group_rule"`
	}
	uret := secGrpRule{}
	err = p.Conn.Call(l, &uret, "GET", "/v2.0/security-group-rules/"+secRuleId)
	if err != nil && fakeError(err) {
		err = nil
	}
	ret = uret.SecRule
	return
}

// --------------------------------------------------
// 删除安全组规则

func (p Client) DeleteSecurityRule(l rpc.Logger, secRuleId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2.0/security-group-rules/"+secRuleId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}
