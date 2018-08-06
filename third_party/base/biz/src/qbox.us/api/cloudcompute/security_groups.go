package cc

import (
	"strconv"

	"github.com/qiniu/rpc.v2"
)

type SecurityRule struct {
	Direction       string `json:"direction,omitempty"`
	RemoteIpPrefix  string `json:"remote_ip_prefix,omitempty"`
	Protocol        string `json:"protocol,omitempty"`
	PortRangeMax    int    `json:"port_range_max,omitempty"`
	PortRangeMin    int    `json:"port_range_min,omitempty"`
	Ethertype       string `json:"ethertype,omitempty"`
	Id              string `json:"id,omitempty"`
	RemoteGroupId   string `json:"remote_group_id,omitempty"`
	SecurityGroupId string `json:"security_group_id,omitempty"`
}

type SecurityGroup struct {
	Description string         `json:"description"`
	Rules       []SecurityRule `json:"security_group_rules"`
	Name        string         `json:"name"`
	Id          string         `json:"id"`
	CreatedAt   string         `json:"created_at"`
}

type ListSecGrpsRet struct {
	SecGrps []SecurityGroup `json:"security_groups"`
}

// ----------------------------------------------------------------------

func (r *Service) ListSecurityGroups(l rpc.Logger) (ret ListSecGrpsRet, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/securitygroups")
	return
}

// ----------------------------------------------------------------------

func (r *Service) CreateSecurityGroup(l rpc.Logger, name, desc string) (ret SecurityGroup, err error) {
	params := map[string][]string{
		"name":        {name},
		"description": {desc},
	}

	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/securitygroups", params)
	return
}

// ----------------------------------------------------------------------

func (r *Service) GetSecurityGroup(l rpc.Logger, secGrpId string) (ret SecurityGroup, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/securitygroups/"+secGrpId)
	return
}

// ----------------------------------------------------------------------

func (r *Service) DeleteSecurityGroup(l rpc.Logger, secGrpId string) error {
	return r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/securitygroups/"+secGrpId)
}

// ----------------------------------------------------------------------
// 创建安全组规则
//
// direction 为必填参数，小写（大小写敏感），可取值为 ingress，egress；
// 若只填 direction，则 ethertype 默认为 IPv4，所有协议所有端口都开放；
//
// protocol 可取 tcp/udp/icmp，大小写不敏感；
//
// remote_ip_prefix 格式为 192.168.10.0/24，如果只是开放给某台主机，
// 可以写成 192.168.10.1/32 (示例)。

type CreateSecRuleArgs struct {
	Direction       string `json:"direction"`
	RemoteIpPrefix  string `json:"remote_ip_prefix"`
	Protocol        string `json:"protocol"`
	PortRangeMax    int    `json:"port_range_max"`
	PortRangeMin    int    `json:"port_range_min"`
	Ethertype       string `json:"ethertype"`
	RemoteGroupId   string `json:"remote_group_id"`
	SecurityGroupId string `json:"security_group_id"`
}

func (r *Service) CreateSecurityRule(l rpc.Logger, args CreateSecRuleArgs) (ret SecurityRule, err error) {
	params := map[string][]string{
		"direction":         {args.Direction},
		"remote_ip_prefix":  {args.RemoteIpPrefix},
		"protocol":          {args.Protocol},
		"port_range_max":    {strconv.Itoa(args.PortRangeMax)},
		"port_range_min":    {strconv.Itoa(args.PortRangeMin)},
		"ethertype":         {args.Ethertype},
		"remote_group_id":   {args.RemoteGroupId},
		"security_group_id": {args.SecurityGroupId},
	}

	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/securityrules", params)
	return
}

// ----------------------------------------------------------------------

func (r *Service) GetSecurityRule(l rpc.Logger, secRuleId string) (ret SecurityRule, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/securityrules/"+secRuleId)
	return
}

// ----------------------------------------------------------------------

func (r *Service) DeleteSecurityRule(l rpc.Logger, secRuleId string) error {
	return r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/securityrules/"+secRuleId)
}

// ----------------------------------------------------------------------
