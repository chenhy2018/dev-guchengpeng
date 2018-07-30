package enums

import (
	"fmt"
)

const (
	OpTypeClusterAll     OpTypeCluster = 0
	OpTypeClusterAccount OpTypeCluster = 1
	OpTypeClusterBucket  OpTypeCluster = 2
)

type OpTypeCluster uint32 // 中文对照表: src/controllers/setting.go [RouteOplog]

func OpTypeClusterMap() map[OpTypeCluster]string {
	return map[OpTypeCluster]string{
		OpTypeClusterAll:     OpTypeClusterAll.Humanize(),
		OpTypeClusterAccount: OpTypeClusterAccount.Humanize(),
		OpTypeClusterBucket:  OpTypeClusterBucket.Humanize(),
	}
}

func (optc OpTypeCluster) Valid() bool {
	return optc >= OpTypeClusterAll && optc <= OpTypeClusterBucket
}

func (optc OpTypeCluster) Humanize() string {
	switch optc {
	case OpTypeClusterAll:
		return "全部"
	case OpTypeClusterAccount:
		return "账户"
	case OpTypeClusterBucket:
		return "空间"
	default:
		return fmt.Sprintf("无效的操作日志状态: %d", optc)
	}
}
