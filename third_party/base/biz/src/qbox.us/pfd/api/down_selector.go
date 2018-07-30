package api

import (
	trackerapi "qbox.us/pfdtracker/api"

	"math"
)

type DownLoad struct {
	LocalIdx  uint32         //always incre, so it's type must be uint32
	RemoteIdx uint32         //always incre, so it's type must be uint32
	LoadVals  map[string]int //key: stg's hostUrl, value:reqCnt
}

func NewDownLoad() *DownLoad {
	return &DownLoad{
		LoadVals: make(map[string]int),
	}
}

func (this *DownLoad) GetLoad(hostUrl string) int {
	load, ok := this.LoadVals[hostUrl]

	if !ok {
		load = 0
		this.LoadVals[hostUrl] = load
	}
	return load
}

func (this *DownLoad) IncreLoad(hostUrl string) {

	if loadVal, ok := this.LoadVals[hostUrl]; ok {
		loadVal++
		this.LoadVals[hostUrl] = loadVal
	}
}

func (this *DownLoad) DecreLoad(hostUrl string) {

	if loadVal, ok := this.LoadVals[hostUrl]; ok {

		loadVal--

		if loadVal < 0 {
			loadVal = 0
		}

		this.LoadVals[hostUrl] = loadVal
	}
}

type DownSelector struct {
	idc          string
	remoteOrders []string // 顺序条件
	Loads        map[uint32]*DownLoad
}

func NewDownSelector(idc string, remoteOrders []string) *DownSelector {
	return &DownSelector{
		idc,
		remoteOrders,
		make(map[uint32]*DownLoad),
	}
}

func (this *DownSelector) GetLoad(dgid uint32) *DownLoad {
	load, ok := this.Loads[dgid]
	if !ok {
		load = NewDownLoad()
		this.Loads[dgid] = load
	}
	return load
}

// 下载时优先选择本机房的机器
// TODO: 优化这段逻辑, 一个推荐思路是rank排序
func (this *DownSelector) SelDiskNode(dgid uint32, isECed bool, dgNode *DGNode, badHostUrls []string) (diskNode *DiskNode, err error) {

	// 策略一: 先选健康(非broken)的节点 再选择非健康的节点
	// 策略二: 在满足策略一的情况下, 先选择非backup的节点,再选择backup的节点
	// 策略三: 在满足策略一、二的情况下, 先选本机房节点 再选其他机房节点
	for _, broken := range []bool{false, true} {
		for _, backup := range []bool{false, true} {
			diskNode, err = this.selLocalDiskNode(dgid, dgNode.LocalNodes, badHostUrls, broken, backup, this.idc)
			if err == nil {
				return
			}
			if len(this.remoteOrders) == 0 {
				diskNode, err = this.selLocalDiskNode(dgid, dgNode.RemoteNodes, badHostUrls, broken, backup, "")
				if err == nil {
					return
				}
			} else {
				for _, idc := range this.remoteOrders {
					diskNode, err = this.selLocalDiskNode(dgid, dgNode.RemoteNodes, badHostUrls, broken, backup, idc)
					if err == nil {
						return
					}
				}
			}

			if err == EDiskNodeNotFind && isECed {
				err = trackerapi.ErrGidECed
				return
			}

		}
	}
	return
}

func (this *DownSelector) selLocalDiskNode(dgid uint32, diskNodes []*DiskNode, badHostUrls []string, broken, backup bool, idc string) (diskNode *DiskNode, err error) {
	var minReqCnt int = math.MaxInt32
	dgLoad := this.GetLoad(dgid)

	// 两组磁盘有各自的轮询 idx
	idx := &dgLoad.LocalIdx
	if len(diskNodes) > 0 && diskNodes[0].Idc != this.idc {
		idx = &dgLoad.RemoteIdx
	}

	for i := *idx; i < *idx+uint32(len(diskNodes)); i++ {
		pos := i % uint32(len(diskNodes))
		node := diskNodes[pos]
		if idc != "" && idc != node.Idc {
			continue
		}
		reqCnt := dgLoad.GetLoad(node.HostUrl)
		if minReqCnt > reqCnt && !inBadHosts(node.HostUrl, badHostUrls) {

			if broken == node.IsRepair && backup == node.IsBackup {
				minReqCnt = reqCnt
				diskNode = node
			}
		}
	}

	if diskNode == nil {
		return nil, EDiskNodeNotFind
	}

	*idx++
	dgLoad.IncreLoad(diskNode.HostUrl)
	return
}
