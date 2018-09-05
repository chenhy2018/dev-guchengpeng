package compute

import (
	"errors"
	"strings"

	"github.com/qiniu/rpc.v2"
	"ustack.com/admin_api.v1/ustack"
)

type Client struct {
	ProjectId string
	Conn      ustack.Conn
}

func New(services ustack.Services, project string) Client {

	conn, ok := services.Find("nova")
	if !ok {
		panic("compute api not found")
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
// 获取系统中所有hypervisor的总计统计信息

type GetHypervisorStatisticsRet struct {
	HypervisorStatistics struct {
		Count              uint32 `json:"count"`
		VcpusUsed          uint32 `json:"vcpus_used"`
		MemoryMb           uint32 `json:"memory_mb"`
		CurrentWorkload    uint32 `json:"current_workload"`
		Vcpus              uint32 `json:"vcpus"`
		RunningVms         uint32 `json:"running_vms"`
		FreeDiskGb         int32  `json:"free_disk_gb"`
		DiskAvailableLeast uint32 `json:"disk_available_least"`
		LocalGb            uint32 `json:"local_gb"`
		FreeRamMb          uint32 `json:"free_ram_mb"`
		MemoryMbUsed       uint32 `json:"memory_mb_used"`
	} `json:"hypervisor_statistics"`
}

func (p Client) GetHypervisorStatistics(l rpc.Logger) (ret *GetHypervisorStatisticsRet, err error) {
	ret = &GetHypervisorStatisticsRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/os-hypervisors/statistics")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// ------------------------------------------------------------------
// 统计所有 active 状态的虚机使用的内存总和（ustack没有提供这个接口）

func (p Client) GetMemUsedReal(l rpc.Logger) (MemGB float32, err error) {
	type serverInfo struct {
		Status string `json:"status"`
		Flavor struct {
			Name string `json:"name"`
		} `flavor`
	}

	type serversRet struct {
		Servers []serverInfo `json:"servers"`
	}

	ret := serversRet{}
	err = p.Conn.Call(l, &ret, "GET", "/v2/"+p.ProjectId+"/servers/detail?all_tenants=1")
	if err != nil {
		if fakeError(err) {
			err = nil
		} else {
			return
		}
	}

	for _, svr := range ret.Servers {
		if strings.EqualFold(svr.Status, "active") {
			var mem float32
			_, mem, err = FlavorConvert(svr.Flavor.Name)
			if err != nil {
				return
			}
			MemGB += mem
		}
	}

	return
}

func FlavorConvert(flavor string) (vcpu int, memGb float32, err error) {
	switch flavor {
	case "micro-1":
		return 1, 0.5, nil
	case "micro-2":
		return 1, 1, nil
	case "standard-1":
		return 1, 2, nil
	case "standard-2":
		return 2, 4, nil
	case "standard-4":
		return 4, 8, nil
	case "standard-8":
		return 8, 16, nil
	case "standard-12":
		return 12, 24, nil
	case "standard-16":
		return 16, 32, nil
	case "memory-1":
		return 1, 4, nil
	case "memory-2":
		return 2, 8, nil
	case "memory-4":
		return 4, 16, nil
	case "memory-8":
		return 8, 32, nil
	case "memory-12":
		return 12, 48, nil
	case "compute-2":
		return 2, 2, nil
	case "compute-4":
		return 4, 4, nil
	case "compute-8":
		return 8, 8, nil
	case "compute-12":
		return 12, 12, nil
	case "memory-16":
		return 16, 64, nil
	default:
		err = errors.New("unknown flavor")
		return
	}
}

// ------------------------------------------------------------------
// 获取 quota 设置

type Quota struct {
	QuotaSet struct {
		Cores              int    `json:"cores,omitempty"`
		FloatingIps        int    `json:"floating_ips,omitempty"`
		Id                 string `json:"id,omitempty"`
		Instances          int    `json:"instances,omitempty"`
		KeyPairs           int    `json:"key_pairs,omitempty"`
		Ram                int32  `json:"ram,omitempty"`
		SecurityGroupRules int    `json:"security_group_rules,omitempty"`
		SecurityGroups     int    `json:"security_groups,omitempty"`
	} `json:"quota_set"`
}

func (p *Client) GetQuota(l rpc.Logger, projectId string) (ret Quota, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2/"+p.ProjectId+"/os-quota-sets/"+projectId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// ------------------------------------------------------------------
// 修改 quota 设置

func (p *Client) PutQuota(l rpc.Logger, projectId string, args Quota) (ret Quota, err error) {
	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2/"+p.ProjectId+"/os-quota-sets/"+projectId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}
