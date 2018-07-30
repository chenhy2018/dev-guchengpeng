package floatip

import (
	"errors"
	"strconv"
	"strings"

	"github.com/qiniu/rpc.v2"
	"ustack.com/admin_api.v1/ustack"
	"ustack.com/api.v1/ustack/network"
)

type Client struct {
	ProjectId string
	Conn      ustack.Conn
}

func New(services ustack.Services, project string) (cli Client, err error) {

	conn, ok := services.Find("neutron")
	if !ok {
		err = errors.New("neutron api not found")
		return
	}

	cli = Client{
		ProjectId: project,
		Conn:      conn,
	}
	return
}

func isFakeError(err error) bool {
	if rpc.HttpCodeOf(err)/200 == 2 {
		return true
	}
	return false
}

// -------------------------------------------------------------------
// 获取所有公网ip的信息

func (p Client) GetFloatingips(l rpc.Logger) (ret network.ListFloatingipsRet, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/floatingips")
	if err != nil && isFakeError(err) {
		err = nil
	}
	return
}

// -------------------------------------------------------------------
// 获取某个公网ip的信息

func (p Client) GetFloatingipInfo(l rpc.Logger, floatingipId string) (ret network.FloatingipRet, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/floatingips/"+floatingipId)
	if err != nil && isFakeError(err) {
		err = nil
	}
	return
}

// -------------------------------------------------------------------
// 获取公网ip统计信息

type FipCount struct {
	Telecom int `json:"telecom"`
	Unicom  int `json:"unicom"`
	Total   int `json:"total"`
}

type GetSubnetsRet struct {
	Subnets []struct {
		Cidr string `json:"cidr"`
	} `json:"subnets"`
}

func (p Client) FloatingipCount(l rpc.Logger, telPrefixes, uniPrefixes []string) (ret FipCount, err error) {
	subnetRet := &GetSubnetsRet{}
	err = p.Conn.Call(l, subnetRet, "GET", "/v2.0/subnets")
	if err != nil && !isFakeError(err) {
		return
	}
	err = nil

	var telecom int
	var unicom int
	var total int
	var count int

	for _, subnet := range subnetRet.Subnets {
		matched := false

		for _, tel := range telPrefixes {
			if strings.HasPrefix(subnet.Cidr, tel) {
				count, err = cidrToCount(subnet.Cidr)
				if err != nil {
					return
				}
				telecom += count
				total += count
				matched = true
				break
			}
		}

		if matched {
			continue
		}

		for _, uni := range uniPrefixes {
			if strings.HasPrefix(subnet.Cidr, uni) {
				count, err = cidrToCount(subnet.Cidr)
				if err != nil {
					return
				}
				unicom += count
				total += count
				matched = true
				break
			}
		}
	}

	ret.Total = total
	ret.Telecom = telecom
	ret.Unicom = unicom
	return
}

func cidrToCount(cidr string) (count int, err error) {
	maskStr := strings.Split(cidr, "/")[1]
	mask, err1 := strconv.Atoi(maskStr)
	if err1 != nil {
		err = errors.New("invalid cidr format:" + cidr)
		return
	}
	bits := uint(32 - mask)
	count = 1<<bits - 3
	return
}

// -------------------------------------------------------------------
// 获取公网IP使用量信息

type FipUsed struct {
	TelcomUsed int
	UnicomUsed int
	TotalUsed  int
}

func (p Client) FloatingipUsed(l rpc.Logger, telPrefixes, uniPrefixes []string) (ret FipUsed, err error) {
	var uret network.ListFloatingipsRet
	err = p.Conn.Call(l, &uret, "GET", "/v2.0/floatingips?uos_ext=1")
	if err != nil && !isFakeError(err) {
		return
	}
	err = nil

	for _, fip := range uret.Floatingips {
		matched := false

		for _, tel := range telPrefixes {
			if strings.HasPrefix(fip.FloatingipAddress, tel) {
				ret.TelcomUsed += 1
				matched = true
				break
			}
		}

		if matched {
			continue
		}

		for _, uni := range uniPrefixes {
			if strings.HasPrefix(fip.FloatingipAddress, uni) {
				ret.UnicomUsed += 1
				matched = true
				break
			}
		}

		if !matched {
			err = errors.New("unrecognized floating ip address: " + fip.FloatingipAddress)
			return
		}
	}

	ret.TotalUsed = len(uret.Floatingips)
	return
}
