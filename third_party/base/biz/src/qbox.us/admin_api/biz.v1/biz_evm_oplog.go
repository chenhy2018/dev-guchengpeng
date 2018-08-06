package biz

import (
	"net/url"
	"strconv"
	"time"

	"github.com/qiniu/rpc.v1"
	"labix.org/v2/mgo/bson"
)

type EvmOpType int

const (
	EvmOpCreate EvmOpType = iota + 1
	EvmOpUpdate
	EvmOpShutdown
	EvmOpDelete
	EvmOpStop
	EvmOpReboot
	EvmOpStart
	EvmOpUpdateCostMode // 更新计费类型
)

func (t EvmOpType) String() string {
	switch t {
	case EvmOpCreate:
		return "CREATE"
	case EvmOpUpdate:
		return "UPDATE"
	case EvmOpShutdown:
		return "SHUTDOWN"
	case EvmOpDelete:
		return "DELETE"
	case EvmOpStop:
		return "STOP"
	case EvmOpReboot:
		return "REBOOT"
	case EvmOpStart:
		return "START"
	case EvmOpUpdateCostMode:
		return "UPDATECOSTMODE"
	default:
		return "N/A"
	}
}

type EvmCostOpType int

const (
	EvmDegrade EvmCostOpType = iota + 1 // 配置更新为降低配置时，对应的对计费的操作类型
	EvmRenew                            // 续费操作
)

func (t EvmCostOpType) String() string {
	switch t {
	case EvmDegrade:
		return "DEGRADE"
	case EvmRenew:
		return "RENEW"
	default:
		return "N/A"
	}
}

type EvmRecType int

const (
	EvmRecTypeCompute EvmRecType = iota + 1
	EvmRecTypeVolume
	EvmRecTypeFloatingip
	EvmRecTypeSnapshot
	EvmRecTypeListener //负载均衡监听器
)

func (t EvmRecType) String() string {
	switch t {
	case EvmRecTypeCompute:
		return "compute"
	case EvmRecTypeVolume:
		return "volume"
	case EvmRecTypeFloatingip:
		return "floatingip"
	case EvmRecTypeSnapshot:
		return "snapshot"
	case EvmRecTypeListener:
		return "listener"
	default:
		return "N/A"
	}
}

type EvmCostModeType int

const (
	EvmOndemand EvmCostModeType = iota + 1
	EvmBymonth
	EvmByyear
)

func (t EvmCostModeType) String() string {
	switch t {
	case EvmOndemand:
		return "ondemand"
	case EvmBymonth:
		return "bymonth"
	case EvmByyear:
		return "byyear"
	default:
		return "N/A"
	}
}

type EvmOplogQueryArgs struct {
	StartTime    time.Time       // optional
	EndTime      time.Time       // optional
	Owner        uint32          // optional
	LastId       string          // optional
	ResourceType EvmRecType      // optional
	CostMode     EvmCostModeType // optional
	Limit        int             // optional, but start_time, end_time, limit must got one
}

type EvmOplogs struct {
	Records []EvmOplogRecord `json:"records"`
	LastId  string           `json:"last_id"`
}

type EvmOplogRecord struct {
	Id           bson.ObjectId   `json:"id"`
	Uid          uint32          `json:"uid"`
	Op           EvmOpType       `json:"op"`
	CostOp       EvmCostOpType   `json:"cost_op"` //与计费相关的op
	Args         string          `json:"args"`
	ResourceId   string          `json:"resource_id"`
	ResourceType EvmRecType      `json:"resource_type"`
	CostMode     EvmCostModeType `json:"cost_mode"`
	Amount       int             `json:"amount"`
	CreatedAt    time.Time       `json:"created_at"`
}

func (s *BizService) ListEvmOplogs(l rpc.Logger, in EvmOplogQueryArgs) (
	ret EvmOplogs, err error) {

	v := url.Values{
		"start_time":    {in.StartTime.Format(time.RFC3339)},
		"end_time":      {in.EndTime.Format(time.RFC3339)},
		"owner":         {strconv.FormatInt(int64(in.Owner), 10)},
		"last_id":       {in.LastId},
		"resource_type": {strconv.FormatInt(int64(in.ResourceType), 10)},
		"cost_mode":     {strconv.FormatInt(int64(in.CostMode), 10)},
		"limit":         {strconv.FormatInt(int64(in.Limit), 10)},
	}

	err = s.rpc.GetCall(l, &ret, s.host+"/evm/oplogs?"+v.Encode())
	return
}
