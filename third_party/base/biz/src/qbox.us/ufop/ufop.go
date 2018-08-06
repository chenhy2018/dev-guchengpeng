package ufop

//QINIU_UAPP是七牛内部的uapp，比如原有的fop操作以ufop方式
//部署，他们对URL的格式有特定要求
const (
	NORMAL_UAPP int = iota
	QINIU_UAPP      = iota
)

const (
	//默认不开启任何外网和内网IP的访问，只允许和gate通信
	NETTYPE_DEFAULT = 0x1
	NETTYPE_LOCAL   = 0x2
	NETTYPE_WIDE    = 0x4
	NETTYPE_FREE    = NETTYPE_LOCAL | NETTYPE_WIDE
)

const (
	DISABLE_DISKCACHE = iota
	ENABLE_DISKCACHE  = iota
)

type UfopInfo struct {
	Id         string `bson:"_id"`
	UfopName   string
	Uid        uint32
	Uapp       string
	DiskCache  int
	CreateTime int64
}

type UappInfo struct {
	Id       string `bson:"_id"`
	UappName string
	Uid      uint32
	//0:private - only uapp owner access
	//1:public - all user can access
	//2:user in the AccList can access
	Mode       byte
	AccList    []uint32
	Bucket     string
	Key        string
	Num        int
	Domains    []string
	Type       int
	NetType    int
	CreateTime int64
	StartTime  int64
}

type DomainInfo struct {
	Id       string `bson:"_id"`
	UappName string
}
