package api

import (
	"math"

	"github.com/qiniu/xlog.v1"
)

func (this *UpSelector) GetLoad(id int64) int {

	load, ok := this.Loads[id]

	if !ok {
		load = 0
		this.Loads[id] = load
	}
	return load
}

func (this *UpSelector) IncreLoad(id int64) {

	if loadVal, ok := this.Loads[id]; ok {
		loadVal++
		this.Loads[id] = loadVal
	}
}

func (this *UpSelector) DecreLoad(id int64) {

	if loadVal, ok := this.Loads[id]; ok {

		loadVal--

		if loadVal < 0 {
			loadVal = 0
		}

		this.Loads[id] = loadVal
	}
}

type UpSelector struct {
	diskIdx uint32 //always incre, so it's type must be uint32

	Loads map[int64]int //key: dgid(loadId), value: reqCnt
}

func NewUpSelector() *UpSelector {
	return &UpSelector{
		0,
		make(map[int64]int),
	}
}

const (
	MAX_DISK_NODE_CHOOSE_NUM = 36
)

func (this *UpSelector) SelDiskNode(xl *xlog.Logger, upDiskNodes []*DiskNode, badHostUrls []string, badDgids []uint32) (diskNode *DiskNode, err error) {

	var selDiskNode *DiskNode
	var minLoad int = math.MaxInt32

	xl.Debugf("len(upDiskNodes):%v diskIdx:%v Loads:%v", len(upDiskNodes), this.diskIdx, this.Loads)

	//选择策略：从上次选择的磁盘节点的下一个节点开始，依次扫描upDiskNodes数组，选择负载最小的磁盘节点。扫描次数限制为最多扫36个不在badHostUrls中的节点，同时，扫描过程中如果发现某节点负载为0，则选择之。
	for i, j := this.diskIdx, 0; i < this.diskIdx+uint32(len(upDiskNodes)) && j < MAX_DISK_NODE_CHOOSE_NUM; i = i + 1 {
		pos := i % uint32(len(upDiskNodes))
		diskNode := upDiskNodes[pos]
		load := this.GetLoad(diskNode.LoadId)

		xl.Debugf("load:%v minLoad:%v diskNode:%v", load, minLoad, *diskNode)
		if !inBadHosts(diskNode.HostUrl, badHostUrls) && !inBadDgid(diskNode.Dgid, badDgids) {
			j++
			if minLoad > load {
				minLoad = load
				selDiskNode = diskNode
			}
		}

		if minLoad == 0 {
			break
		}
	}

	if selDiskNode == nil {
		return nil, EStgNodeNotFind
	}

	this.diskIdx++
	this.IncreLoad(selDiskNode.LoadId)
	return selDiskNode, nil
}
