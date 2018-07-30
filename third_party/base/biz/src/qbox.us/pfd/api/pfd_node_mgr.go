package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	cfgapi "qbox.us/pfdcfg/api"
)

var (
	EDiskMatrixNotFind = errors.New("disk matrix not found")
	EStgNodeNotFind    = errors.New("stg node not found")
	EDiskNodeNotFind   = errors.New("disk node not found")
	EDGNodeNotFind     = errors.New("dg node not found")
)

type DiskNode struct {
	Dgid     uint32
	DiskType cfgapi.DiskType
	ReadOnly uint32
	Weight   uint32
	IsBackup bool
	HostUrl  string
	Idc      string
	LoadId   int64
	IsRepair bool
}

type DGNode struct {
	Idc         string // 客户端所在 idc 标记
	Dgid        uint32
	DiskType    cfgapi.DiskType
	DiskNodes   []*DiskNode // 所有节点：本机房 + 其余机房的节点
	LocalNodes  []*DiskNode // 本机房的节点
	RemoteNodes []*DiskNode // 非本机房的节点
}

func (this *DGNode) add(node *DiskNode) {
	this.DiskNodes = append(this.DiskNodes, node)
	if this.Idc == node.Idc {
		this.LocalNodes = append(this.LocalNodes, node)
	} else {
		this.RemoteNodes = append(this.RemoteNodes, node)
	}
}

func (this *DGNode) getHostUrls() (hostUrls []string) {
	hostUrls = make([]string, 0)
	for _, diskNode := range this.DiskNodes {
		hostUrls = append(hostUrls, diskNode.HostUrl)
	}
	return hostUrls
}

func (this *DGNode) hasBackup() bool {
	for _, node := range this.DiskNodes {
		if node.IsBackup {
			return true
		}
	}
	return false
}

func (this *DGNode) allBackup() bool {
	for _, node := range this.DiskNodes {
		if !node.IsBackup {
			return false
		}
	}
	return true
}

func NewDGNode(idc string, dgid uint32, diskType cfgapi.DiskType) *DGNode {
	return &DGNode{
		Idc:      idc,
		Dgid:     dgid,
		DiskType: diskType,
	}
}

/**
PfdNodeMgr中缓存pfd系统的配置信息和负载信息， 在读/写文件时，根据负载情况，从挑选出一个磁盘节点来进行读/写。
根据pfd系统的体系结构，PfdNodeMgr中有三种节点：
1.disk节点：对应一块磁盘
2.stg节点： 对应一个pfdstg进程，一个该进程下面可以挂12或36块磁盘，每块磁盘分属不同的磁盘组
3.dg节点： 对应一个磁盘组。一个磁盘组由3块磁盘构成，一主二备，这三块磁盘，分别挂在不同的stg节点下
写文件时，客户端只往主磁盘写，由主磁盘同步到两块从盘。
读文件时，客户端可以选择三块盘中的一块进行读取。

PS1：磁盘分为两种，SSD和普通磁盘（DEFAULT）。系统设计约定：
1.一个stg节点下使用类型一致的磁盘，或者全部的SSD，或者全部DEFAULT
2.一个磁盘组中使用类型一致的磁盘

PS2：配置约定：
在pfdcfg的磁盘组配置中，规定第一个stg为主stg（写操作发往该stg），后两个为备stg。
*/
type PfdNodeMgr struct {

	/**
	静态的配置信息，每隔一段时间从pfdcfg中拉取并刷新
	UpDiskNodes: 存储可写的disk节点的数组，数组中相邻的节点属于相邻的stg，只含本机房的 master 节点
	DownDGNodes: 存储dgid和disk节点的对应关系， 供读文件时选取disk节点用，所有节点(包含其余机房)
	*/
	UpDiskNodes map[cfgapi.DiskType][]*DiskNode
	DownDGNodes map[uint32]*DGNode

	/**
	动态的负载和选择信息，不刷新只更新（在做put/get等操作后更新负载和选择信息）
	upSel: 存储文件写操作中，各disk节点的负载和当前选择的节点信息
	downSel: 存储文件读操作中， 各disk节点的负载和当前选择的节点的信息
	*/
	upSel   *UpSelector
	downSel *DownSelector

	idc string

	//lock
	mutex sync.RWMutex
}

func NewPfdNodeMgr(idc string, remoteOrders []string) *PfdNodeMgr {
	return &PfdNodeMgr{
		UpDiskNodes: make(map[cfgapi.DiskType][]*DiskNode),
		DownDGNodes: make(map[uint32]*DGNode),
		upSel:       NewUpSelector(),
		downSel:     NewDownSelector(idc, remoteOrders),
		idc:         idc,
		mutex:       sync.RWMutex{},
	}
}

func inBadHosts(hostUrl string, badHostUrls []string) (ret bool) {
	for _, url := range badHostUrls {
		if hostUrl == url {
			return true
		}
	}

	return false
}

func inBadDgid(dgid uint32, badDgids []uint32) (ret bool) {
	for _, id := range badDgids {
		if id == dgid {
			return true
		}
	}

	return false
}

/**
在处理put/alloc这两种写文件操作时，该函数用于挑选出一个disk节点来处理写文件
diskType: 磁盘类型，指明需要挑选SSD磁盘还是普通磁盘
badHostUrls: stg黑名单，指明哪些stg已经被验证不可用
*/
func (this *PfdNodeMgr) SelectUpDisk(xl *xlog.Logger, diskType cfgapi.DiskType, badHostUrls []string, badDgids []uint32) (diskNode *DiskNode, err error) {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	upDiskNodes, ok := this.UpDiskNodes[diskType]
	if !ok || len(upDiskNodes) == 0 {
		return nil, EDiskMatrixNotFind
	}

	diskNode, err = this.upSel.SelDiskNode(xl, upDiskNodes, badHostUrls, badDgids)
	return
}

func (this *PfdNodeMgr) ReleaseUpDisk(diskNode *DiskNode) (err error) {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.upSel.DecreLoad(diskNode.LoadId)

	return nil
}

func (this *PfdNodeMgr) ReleaseDownDisk(diskNode *DiskNode) (err error) {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	dgLoad := this.downSel.GetLoad(diskNode.Dgid)
	dgLoad.DecreLoad(diskNode.HostUrl)

	return nil
}

func (this *PfdNodeMgr) LoadDGInfo(dgInfo *cfgapi.DiskGroupInfo) (err error) {

	dgNode := NewDGNode(this.idc, dgInfo.Dgid, dgInfo.DiskType)

	for i, hostUrls := range dgInfo.Hosts {
		//做一个保护，以防IsBackup数组没有Hosts数组长
		hostUrl := hostUrls[0]
		isBackup := false
		if dgInfo.IsBackup != nil && len(dgInfo.IsBackup) > i {
			isBackup = dgInfo.IsBackup[i]
		}
		repair := false
		if dgInfo.Repair != nil && len(dgInfo.Repair) > i {
			repair = dgInfo.Repair[i]
		}
		var idc string
		if len(dgInfo.Idc) > i { // 保持兼容
			idc = dgInfo.Idc[i]
		}

		diskNode := DiskNode{
			dgInfo.Dgid,
			dgInfo.DiskType,
			dgInfo.ReadOnly,
			dgInfo.Weight,
			isBackup,
			hostUrl,
			idc,
			0,
			repair,
		}
		dgNode.add(&diskNode)
	}

	err = this.saveDGNode(dgNode)
	return
}

type stgNode struct {
	hostUrl   string
	diskNodes []*DiskNode
}

func newStgNode(hostUrl string) *stgNode {
	return &stgNode{
		hostUrl,
		make([]*DiskNode, 0),
	}
}

type diskMatrix struct {
	diskType cfgapi.DiskType
	stgNodes []*stgNode
}

func newDiskMatrix(diskType cfgapi.DiskType) *diskMatrix {
	return &diskMatrix{
		diskType,
		make([]*stgNode, 0),
	}
}

func maxDiskNodeNum(stgNodes []*stgNode) int {
	maxNum := 0
	for _, node := range stgNodes {
		if maxNum < len(node.diskNodes) {
			maxNum = len(node.diskNodes)
		}
	}
	return maxNum
}

func loadId(dgid uint32, instanceId uint32) int64 {

	return int64(dgid)<<32 + int64(instanceId)
}

func (this *PfdNodeMgr) LoadPfdNodes(xl *xlog.Logger, dgis []*cfgapi.DiskGroupInfo) {

	diskMatrixs := make(map[cfgapi.DiskType]*diskMatrix)
	dgNodes := make(map[uint32]*DGNode)

	// 关于 DiskType 和 Host 的约定:
	//   1. 相同 Stg 的 Host 统一使用一致类型的磁盘
	//   2. 为了在同一台物理机上混合部署不同类型的磁盘, 只需同时部署多个 Stg 既可
	for _, dgi := range dgis {
		dgid := dgi.Dgid

		var dgNode *DGNode = nil
		var ok bool = false

		if dgNode, ok = dgNodes[dgid]; !ok {
			dgNode = NewDGNode(this.idc, dgid, dgi.DiskType)
			dgNodes[dgid] = dgNode
		}

		for j, hostUrls := range dgi.Hosts {
			hostUrl := hostUrls[0]

			//做一个保护，以防IsBackup数组没有Hosts数组长
			var repair, isBackup bool = false, false
			if dgi.IsBackup != nil && len(dgi.IsBackup) > j {
				isBackup = dgi.IsBackup[j]
			}
			if dgi.Repair != nil && len(dgi.Repair) > j {
				repair = dgi.Repair[j]
			}
			var idc string
			if len(dgi.Idc) > j { // 保持兼容
				idc = dgi.Idc[j]
			}
			diskNode := DiskNode{
				dgi.Dgid,
				dgi.DiskType,
				dgi.ReadOnly,
				dgi.Weight,
				isBackup,
				hostUrl,
				idc,
				loadId(dgi.Dgid, 0),
				repair,
			}
			dgNode.add(&diskNode)

			//默认第一个host是主host， 其他host都是备host。只有主host才可写
			//slave正在修盘的, 也不能进行写操作
			hasRepair := false
			if j == 0 {
				for _, repair := range dgi.Repair {
					if repair {
						hasRepair = true
						break
					}
				}
			}
			if j == 0 && 0 == diskNode.ReadOnly && !hasRepair {

				var diskMatrix *diskMatrix = nil
				if diskMatrix, ok = diskMatrixs[diskNode.DiskType]; !ok {
					diskMatrix = newDiskMatrix(diskNode.DiskType)
					diskMatrixs[diskNode.DiskType] = diskMatrix
				}

				var stgNode *stgNode
				for _, node := range diskMatrix.stgNodes {
					if node.hostUrl == hostUrl {
						stgNode = node
						break
					}
				}

				if stgNode == nil {
					stgNode = newStgNode(hostUrl)
					diskMatrix.stgNodes = append(diskMatrix.stgNodes, stgNode)
				}

				stgNode.diskNodes = append(stgNode.diskNodes, &diskNode)
			}
		}
	}

	// 根据 Weight 再调整磁盘节点
	for _, diskMatrix := range diskMatrixs {
		for _, stgNode := range diskMatrix.stgNodes {
			var max uint32
			for _, node := range stgNode.diskNodes {
				if max < node.Weight {
					max = node.Weight
				}
			}
			nodes := stgNode.diskNodes
			for i := uint32(1); i <= max; i++ {
				for _, node := range stgNode.diskNodes {
					if node.Weight >= i {
						clone := *node
						clone.LoadId = loadId(clone.Dgid, i)
						nodes = append(nodes, &clone)
					}
				}
			}
			stgNode.diskNodes = nodes
		}
	}

	//将磁盘矩阵中的磁盘节点，归并到一个磁盘节点数组中。该磁盘数组用于up操作时，选择负载最小的磁盘节点用。数组中相邻的磁盘节点，属于相邻的stg节点。
	mapUpDiskNodes := make(map[cfgapi.DiskType][]*DiskNode, 0)
	for diskType, diskMatrix := range diskMatrixs {

		var upDiskNodes []*DiskNode
		maxDiskNodeNum := maxDiskNodeNum(diskMatrix.stgNodes)
		for j := 0; j < maxDiskNodeNum; j++ {

			for i := 0; i < len(diskMatrix.stgNodes); i++ {
				stgNode := diskMatrix.stgNodes[i]
				if len(stgNode.diskNodes) <= j {
					continue
				}
				diskNode := stgNode.diskNodes[j]
				if diskNode.Idc != "" && this.idc != "" && diskNode.Idc != this.idc { // 只能上传到本地机房的副本，判断非空用于保证兼容
					continue
				}
				upDiskNodes = append(upDiskNodes, diskNode)
			}
		}
		mapUpDiskNodes[diskType] = upDiskNodes
	}

	this.mutex.Lock()

	this.UpDiskNodes = mapUpDiskNodes
	this.DownDGNodes = dgNodes

	this.mutex.Unlock()

	this.printPfdNodeMgr(xl, dgis)
}

func (this *PfdNodeMgr) printPfdNodeMgr(xl *xlog.Logger, dgis []*cfgapi.DiskGroupInfo) {

	if log.Ldebug < log.Std.Level {
		return
	}
	if len(this.DownDGNodes) < 10 { // 小于10个dgid，认为是本地调试环境，不打印
		return
	}

	bytes, err := json.Marshal(dgis)
	if err != nil {
		err = errors.New("encode dgis to json failed.")
		xl.Errorf("err:%v", err)
	} else {
		xl.Debugf("dgis:\r\n%v", string(bytes[:]))
	}

	upDiskNodesStr := ""
	for diskType, diskNodes := range this.UpDiskNodes {

		disksStr := ""
		for j, diskNode := range diskNodes {
			disksStr += fmt.Sprintf("{index:%v,dgid:%v,weight:%v,disktype:%v,readonly:%v,isbackup:%v,hosturl:'%v'},\r\n", j, diskNode.Dgid, diskNode.Weight, diskNode.DiskType, diskNode.ReadOnly, diskNode.IsBackup, diskNode.HostUrl)
		}
		upDiskNodesStr += fmt.Sprintf("{diskType:%v, disksStr:[\r\n%v]}\r\n,", diskType, disksStr)
	}

	upDiskNodesJson := fmt.Sprintf("{[\r\n%v]}\r\n", upDiskNodesStr)

	dgsStr := ""
	for dgid, dgNode := range this.DownDGNodes {

		disksStr := ""
		for i, diskNode := range dgNode.DiskNodes {
			diskStr := fmt.Sprintf("{index:%v,dgid:%v,disktype:%v,readonly:%v,isbackup:%v,hosturl:'%v'},\r\n", i, diskNode.Dgid, diskNode.DiskType, diskNode.ReadOnly, diskNode.IsBackup, diskNode.HostUrl)
			disksStr += diskStr
		}
		dgStr := fmt.Sprintf("{dgid:%v,disknodes:[\r\n%v]}\r\n", dgid, disksStr)

		dgsStr += dgStr
	}

	dgnodesJson := fmt.Sprintf("{[\r\n%v]}\r\n", dgsStr)

	//selector:
	upSelJson := ""
	downSelJson := ""
	this.mutex.RLock()
	for k, v := range this.upSel.Loads {
		upSelJson += fmt.Sprintf("dgid:%v,load:%v ", k, v)
	}
	for k, v := range this.downSel.Loads {
		downSelJson += fmt.Sprintf("dgid:%v,load:%v ", k, v)
	}
	this.mutex.RUnlock()

	xl.Debugf("UPDISKNODES:\r\n%v\r\n\r\nDGS:\r\n%v\r\nupSel:%v\r\ndownSel:%v\r\n", upDiskNodesJson, dgnodesJson, upSelJson, downSelJson)
}

func (this *PfdNodeMgr) GetDownDiskNode(dgid uint32, isECed bool, badHostUrls []string) (diskNode *DiskNode, err error) {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	dgNode := this.DownDGNodes[dgid]
	if dgNode == nil {
		return nil, EDGNodeNotFind
	}

	diskNode, err = this.downSel.SelDiskNode(dgid, isECed, dgNode, badHostUrls)
	return diskNode, err
}

/**
只将dg信息保存到供下载用的缓存中，而不保存到上传用节点缓存，因为上传用节点缓存可以通过固定时间间隔获取到
*/
func (this *PfdNodeMgr) saveDGNode(dgNode *DGNode) (err error) {

	if dgNode == nil {
		return errors.New("dg node param is nil")
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.DownDGNodes[dgNode.Dgid] = dgNode
	return nil
}

func (this *PfdNodeMgr) GetType(dgid uint32) (dt cfgapi.DiskType, ok bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	if dgNode, ok := this.DownDGNodes[dgid]; ok {
		return dgNode.DiskType, true
	}

	return cfgapi.DEFAULT, false
}

func (this *PfdNodeMgr) UsedAllHost(dgid uint32, usedHostUrls []string) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	dgNode := this.DownDGNodes[dgid]
	if dgNode == nil {
		return true
	}

	for _, diskNode := range dgNode.DiskNodes {
		hostUrl := diskNode.HostUrl
		var isUrlUsed bool = false
		for _, url := range usedHostUrls {
			if hostUrl == url {
				isUrlUsed = true
			}
		}
		if !isUrlUsed {
			return false
		}
	}

	return true
}

func (this *PfdNodeMgr) GetMasterHostUrl(dgid uint32) (hostUrl string, hasBackup bool, err error) {
	hostUrl, _, hasBackup, err = this.GetMasterHostUrlAndIdc(dgid)
	return
}

func (this *PfdNodeMgr) GetMasterHostUrlAndIdc(dgid uint32) (hostUrl string, idc string, hasBackup bool, err error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	dgNode := this.DownDGNodes[dgid]
	if dgNode == nil {
		return "", "", false, EDGNodeNotFind
	}

	return dgNode.DiskNodes[0].HostUrl, dgNode.DiskNodes[0].Idc, dgNode.hasBackup(), nil
}

func (this *PfdNodeMgr) AllBackup(dgid uint32) (bool, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	dgNode := this.DownDGNodes[dgid]
	if dgNode == nil {
		return false, EDGNodeNotFind
	}
	return dgNode.allBackup(), nil
}
