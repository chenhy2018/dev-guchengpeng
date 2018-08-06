package uapp

import (
	"github.com/qiniu/ctype"
	"strings"
)

const (
	NORMAL_UAPP int = iota // 用户 uapp, 以非七牛 fop 的请求规格请求
	QINIU_UAPP      = iota // 七牛内部 uapp，使用原有 fop 的请求规格
)

const (
	// 默认不开启任何外网和内网IP的访问，只允许和gate通信
	NETTYPE_DEFAULT = 0x1
	NETTYPE_LOCAL   = 0x2
	NETTYPE_WIDE    = 0x4
	NETTYPE_FREE    = NETTYPE_LOCAL | NETTYPE_WIDE
)

const (
	DISABLE_DISKCACHE = iota
	ENABLE_DISKCACHE  = iota

	MAX_UAPP_LENGTH = 64
)

//-----------------------------------------------------------------------------------------------

type UappInfo struct {
	Name  string `json:"name" bson:"name"`
	Owner uint32 `json:"owner" bson:"owner"`

	// 0: NORMAL_UAPP [default]
	// 1: QINIU_UAPP
	ReqType uint32 `json:"req_type" bson:"req_type"`

	// 0: daemon service [default]
	// 1: one-shoot service
	SrvType uint32 `json:"srv_type" bson:"srv_type"`

	// 0: private - only uapp owner access [default]
	// 1: public - all user can access
	// 2: protected - user in the AccList can access
	AclMode byte `json:"acl_mode" bson:"acl_mode"`

	// qiniu:<bucket>:<key> or an url
	ImageURL string `json:"image_url" bson:"image_url"`
	ImgVer   int    `json:"img_ver" bson:"img_ver"`
	Desc     string `json:"desc" bson:"desc"`

	AccList    []uint32 `json:"acc_list" bson:"acl_list"`
	NetType    int      `json:"net_type" bson:"net_type"`
	CreateTime int64    `json:"create_time" bson:"create_time"`

	// instance number quota
	InstQuota uint32   `json:"inst_quota" bson:"inst_quota"`
	RsCap     Resource `json:"rs_cap" bson:"rs_cap"`
}

// is uapp name valid?
func IsNameValid(uappname string) bool {

	return len(uappname) <= MAX_UAPP_LENGTH &&
		uappname != "" && !strings.Contains(uappname, ":") &&
		ctype.IsType(ctype.XMLSYMBOL_NEXT_CHAR, uappname)
}

//-----------------------------------------------------------------------------------------------

type Resource struct {
	Mem  uint64 `json:"mem" bson:"mem"`   // 内存资源 - byte
	Net  uint64 `json:"net" bson:"net"`   // 网络流量 - bps
	Disk uint64 `json:"disk" bson:"disk"` // 磁盘用量 - byte
	Iops uint64 `json:"iops" bson:"iops"` // Iops - iops
	Cpu  uint64 `json:"cpu" bson:"cpu"`   // Cpu share - relative numeric
}

func (r *Resource) Sub(r1 Resource) *Resource {

	r.Mem -= r1.Mem
	r.Net -= r1.Net
	r.Disk -= r1.Disk
	r.Iops -= r1.Iops
	r.Cpu -= r1.Cpu

	return r
}

func (r *Resource) ImmutableSub(r1 Resource) Resource {

	var ret Resource
	ret.Mem = r.Mem - r1.Mem
	ret.Net = r.Net - r1.Net
	ret.Disk = r.Disk - r1.Disk
	ret.Iops = r.Iops - r1.Iops
	ret.Cpu = r.Cpu - r1.Cpu
	return ret
}

func (r *Resource) Add(r1 Resource) {

	r.Mem += r1.Mem
	r.Net += r1.Net
	r.Disk += r1.Disk
	r.Iops += r1.Iops
	r.Cpu += r1.Cpu
}

func (r *Resource) Enough(r1 Resource) bool {

	f := func(r, r1 uint64) bool {
		return r >= r1
	}
	return f(r.Mem, r1.Mem) &&
		f(r.Net, r1.Net) &&
		f(r.Disk, r1.Disk) &&
		f(r.Iops, r1.Iops) && f(r.Cpu, r1.Cpu)
}

func (r *Resource) Cap() uint64 {

	s := []uint64{r.Mem, r.Net, r.Disk, r.Iops, r.Cpu}
	min := s[0]
	for i := 1; i < len(s); i++ {
		if min > s[i] {
			min = s[i]
		}
	}
	return min
}

var (
	MemWeight  = 0.4
	CpuWeight  = 0.4
	NetWeight  = 0.1
	IopsWeight = 0.05
	DiskWeight = 0.05
)

func (r *Resource) Score() (score float64) {

	score += float64(r.Mem) * MemWeight
	score += float64(r.Cpu) * CpuWeight
	score += float64(r.Net) * NetWeight
	score += float64(r.Iops) * IopsWeight
	score += float64(r.Disk) * DiskWeight

	return
}

// please make sure r > used
func (r *Resource) LeftScore(used Resource) (score float64) {

	return r.Sub(used).Score()
}

// please make sure r > used
func (r *Resource) LeftPercentScore(used Resource) (score float64) {

	score += (float64(r.Mem-used.Mem) / float64(r.Mem)) * MemWeight
	score += (float64(r.Cpu-used.Cpu) / float64(r.Cpu)) * CpuWeight
	score += (float64(r.Net-used.Net) / float64(r.Net)) * NetWeight
	score += (float64(r.Iops-used.Iops) / float64(r.Iops)) * IopsWeight
	score += (float64(r.Disk-used.Disk) / float64(r.Disk)) * DiskWeight

	return
}

//----------------------------------------------------------------------------------------------------

const (
	STATE_INVALID   InsState = iota
	STATE_STOPPED   InsState = iota
	STATE_RUNNING   InsState = iota
	STATE_DELETED   InsState = iota
	STATE_STARTING  InsState = iota // 表明该实例正在启动，但还未确认其已经处于正确运行的状态
	STATE_KILLED    InsState = iota // 表明该实例被系统杀死，可能因为资源使用超限
	STATE_UNKNOWN   InsState = iota // 表明当前无法查询该实例的真实状态
	STATE_MIGRATING InsState = iota
)

type InsState int

func (us InsState) String() string {

	switch us {
	case STATE_STOPPED:
		return "stopped"
	case STATE_RUNNING:
		return "running"
	case STATE_DELETED:
		return "delete"
	case STATE_STARTING:
		return "starting"
	case STATE_KILLED:
		return "killed"
	case STATE_UNKNOWN:
		return "unknown"
	case STATE_MIGRATING:
		return "migrating"
	case STATE_INVALID:
		fallthrough
	default:
		return "invalid"
	}
	return ""
}

//----------------------------------------------------------------------------------------------------
// 把cc的资源单位转换成标准单位

// 1mem == 1MB
func StdMemUnit(mem uint64) uint64 {

	return (1 << 20) * mem
}

// 1disk == 1GB
func StdDiskUnit(disk uint64) uint64 {

	return (1 << 30) * disk
}

func StdCpuShareUnit(cpushare uint64) uint64 {

	return cpushare
}
