package zone

import (
	"fmt"
	"strconv"
	"strings"
)

type Zone int

const (
	// Kodo
	ZONE_NB  Zone = 0 // 宁波
	ZONE_BC  Zone = 1 // 昌平
	ZONE_LAC Zone = 2 // 北美
	ZONE_GZ  Zone = 3 // 华南
	ZONE_SG  Zone = 4 // 新加坡

	// Kirk
	ZONE_KIRK_BQ  Zone = 1000
	ZONE_KIRK_NQ  Zone = 1001
	ZONE_KIRK_LAC Zone = 1002
	ZONE_KIRK_GQ  Zone = 1003
	ZONE_KIRK_XQ  Zone = 1004

	// Kylin
	ZONE_KYLIN_BJ_1        Zone = 2001
	ZONE_KYLIN_SH_1        Zone = 2011
	ZONE_KYLIN_GZ_0        Zone = 2020
	ZONE_KYLIN_GZ_1        Zone = 2021
	ZONE_KYLIN_GZ_2        Zone = 2022
	ZONE_KYLIN_HK_1        Zone = 2031
	ZONE_KYLIN_TORONTO_1   Zone = 2041
	ZONE_KYLIN_SINGAPORE_1 Zone = 2051

	// Fusion
	// 国内
	ZONE_FUSION_CHINA Zone = 3001
	// 海外分区
	ZONE_FUSION_AMEU Zone = 3002
	ZONE_FUSION_ASIA Zone = 3003
	ZONE_FUSION_SEA  Zone = 3004
	ZONE_FUSION_SA   Zone = 3005
	ZONE_FUSION_OC   Zone = 3006
	// 海外总量
	ZONE_FUSION_FOREIGN Zone = 3007
	// 对新用户海外不分区域的区域
	ZONE_FUSION_NOZONE Zone = 3008

	// QVM
	// 国内
	ZONE_QVM_QD   Zone = 4001 // 青岛
	ZONE_QVM_BJ   Zone = 4002 // 北京
	ZONE_QVM_ZJK  Zone = 4003 // 张家口
	ZONE_QVM_HHHT Zone = 4004 // 呼和浩特
	ZONE_QVM_HZ   Zone = 4005 // 杭州
	ZONE_QVM_SH   Zone = 4006 // 上海
	ZONE_QVM_SZ   Zone = 4007 // 深圳
	ZONE_QVM_HK   Zone = 4008 // 香港
	// 海外
	ZONE_QVM_SINGAPORE Zone = 4009 // 新加坡
	ZONE_QVM_SYDNEY    Zone = 4010 // 悉尼
	ZONE_QVM_KL        Zone = 4011 // 吉隆坡
	ZONE_QVM_JAKARTA   Zone = 4012 // 雅加达
	ZONE_QVM_VIRGINIA  Zone = 4013 // 弗吉尼亚
	ZONE_QVM_SV        Zone = 4014 // 硅谷
	ZONE_QVM_FK        Zone = 4015 // 法兰克福
	ZONE_QVM_DUBAI     Zone = 4016 // 迪拜
	ZONE_QVM_MUMBAI    Zone = 4017 // 孟买
	ZONE_QVM_TOKYO     Zone = 4018 // 东京

	// KE
	ZONE_KE_XQ Zone = 5001
	ZONE_KE_BQ Zone = 5002
	ZONE_KE_JQ Zone = 5003
	ZONE_KE_DQ Zone = 5004
)

var (
	Zones = []Zone{
		ZONE_NB,
		ZONE_BC,
		ZONE_LAC,
		ZONE_GZ,
		ZONE_SG,

		ZONE_KIRK_BQ,
		ZONE_KIRK_NQ,
		ZONE_KIRK_LAC,
		ZONE_KIRK_GQ,
		ZONE_KIRK_XQ,

		ZONE_KYLIN_BJ_1,
		ZONE_KYLIN_SH_1,
		ZONE_KYLIN_GZ_0,
		ZONE_KYLIN_GZ_1,
		ZONE_KYLIN_GZ_2,
		ZONE_KYLIN_HK_1,
		ZONE_KYLIN_TORONTO_1,
		ZONE_KYLIN_SINGAPORE_1,

		ZONE_FUSION_CHINA,
		ZONE_FUSION_AMEU,
		ZONE_FUSION_ASIA,
		ZONE_FUSION_SEA,
		ZONE_FUSION_SA,
		ZONE_FUSION_OC,
		ZONE_FUSION_FOREIGN,
		ZONE_FUSION_NOZONE,

		ZONE_QVM_QD,
		ZONE_QVM_BJ,
		ZONE_QVM_ZJK,
		ZONE_QVM_HHHT,
		ZONE_QVM_HZ,
		ZONE_QVM_SH,
		ZONE_QVM_SZ,
		ZONE_QVM_HK,
		ZONE_QVM_SINGAPORE,
		ZONE_QVM_SYDNEY,
		ZONE_QVM_KL,
		ZONE_QVM_JAKARTA,
		ZONE_QVM_VIRGINIA,
		ZONE_QVM_SV,
		ZONE_QVM_FK,
		ZONE_QVM_DUBAI,
		ZONE_QVM_MUMBAI,
		ZONE_QVM_TOKYO,

		ZONE_KE_XQ,
		ZONE_KE_BQ,
		ZONE_KE_JQ,
		ZONE_KE_DQ,
	}
)

var zoneNameMap = map[string]Zone{
	"z0":  ZONE_NB,
	"z1":  ZONE_BC,
	"na0": ZONE_LAC,
	"z2":  ZONE_GZ,
	"as0": ZONE_SG,

	"bq":  ZONE_KIRK_BQ,
	"nq":  ZONE_KIRK_NQ,
	"lac": ZONE_KIRK_LAC,
	"gq":  ZONE_KIRK_GQ,
	"xq":  ZONE_KIRK_XQ,

	"kbj1": ZONE_KYLIN_BJ_1,
	"ksh1": ZONE_KYLIN_SH_1,
	"kgz0": ZONE_KYLIN_GZ_0,
	"kgz1": ZONE_KYLIN_GZ_1,
	"kgz2": ZONE_KYLIN_GZ_2,
	"khk1": ZONE_KYLIN_HK_1,
	"kto1": ZONE_KYLIN_TORONTO_1,
	"ksg1": ZONE_KYLIN_SINGAPORE_1,

	"fschina":   ZONE_FUSION_CHINA,
	"fsameu":    ZONE_FUSION_AMEU,
	"fsasia":    ZONE_FUSION_ASIA,
	"fssea":     ZONE_FUSION_SEA,
	"fssa":      ZONE_FUSION_SA,
	"fsoc":      ZONE_FUSION_OC,
	"fsforeign": ZONE_FUSION_FOREIGN,
	"fsnozone":  ZONE_FUSION_NOZONE,

	"qvmqd":   ZONE_QVM_QD,
	"qvmbj":   ZONE_QVM_BJ,
	"qvmzjk":  ZONE_QVM_ZJK,
	"qvmhhht": ZONE_QVM_HHHT,
	"qvmhz":   ZONE_QVM_HZ,
	"qvmsh":   ZONE_QVM_SH,
	"qvmsz":   ZONE_QVM_SZ,
	"qvmhk":   ZONE_QVM_HK,
	"qvmsg":   ZONE_QVM_SINGAPORE,
	"qvmsn":   ZONE_QVM_SYDNEY,
	"qvmkl":   ZONE_QVM_KL,
	"qvmjk":   ZONE_QVM_JAKARTA,
	"qvmvg":   ZONE_QVM_VIRGINIA,
	"qvmsv":   ZONE_QVM_SV,
	"qvmfk":   ZONE_QVM_FK,
	"qvmdb":   ZONE_QVM_DUBAI,
	"qvmmb":   ZONE_QVM_MUMBAI,
	"qvmtky":  ZONE_QVM_TOKYO,

	"kexq": ZONE_KE_XQ,
	"kebq": ZONE_KE_BQ,
	"kejq": ZONE_KE_JQ,
	"kedq": ZONE_KE_DQ,
}

var zoneNameMapRev = map[Zone]string{
	ZONE_NB:  "z0",
	ZONE_BC:  "z1",
	ZONE_LAC: "na0",
	ZONE_GZ:  "z2",
	ZONE_SG:  "as0",

	ZONE_KIRK_BQ:  "bq",
	ZONE_KIRK_NQ:  "nq",
	ZONE_KIRK_LAC: "lac",
	ZONE_KIRK_GQ:  "gq",
	ZONE_KIRK_XQ:  "xq",

	ZONE_KYLIN_BJ_1:        "kbj1",
	ZONE_KYLIN_SH_1:        "ksh1",
	ZONE_KYLIN_GZ_0:        "kgz0",
	ZONE_KYLIN_GZ_1:        "kgz1",
	ZONE_KYLIN_GZ_2:        "kgz2",
	ZONE_KYLIN_HK_1:        "khk1",
	ZONE_KYLIN_TORONTO_1:   "kto1",
	ZONE_KYLIN_SINGAPORE_1: "ksg1",

	ZONE_FUSION_CHINA:   "fschina",
	ZONE_FUSION_AMEU:    "fsameu",
	ZONE_FUSION_ASIA:    "fsasia",
	ZONE_FUSION_SEA:     "fssea",
	ZONE_FUSION_SA:      "fssa",
	ZONE_FUSION_OC:      "fsoc",
	ZONE_FUSION_FOREIGN: "fsforeign",
	ZONE_FUSION_NOZONE:  "fsnozone",

	ZONE_QVM_QD:        "qvmqd",
	ZONE_QVM_BJ:        "qvmbj",
	ZONE_QVM_ZJK:       "qvmzjk",
	ZONE_QVM_HHHT:      "qvmhhht",
	ZONE_QVM_HZ:        "qvmhz",
	ZONE_QVM_SH:        "qvmsh",
	ZONE_QVM_SZ:        "qvmsz",
	ZONE_QVM_HK:        "qvmhk",
	ZONE_QVM_SINGAPORE: "qvmsg",
	ZONE_QVM_SYDNEY:    "qvmsn",
	ZONE_QVM_KL:        "qvmkl",
	ZONE_QVM_JAKARTA:   "qvmjk",
	ZONE_QVM_VIRGINIA:  "qvmvg",
	ZONE_QVM_SV:        "qvmsv",
	ZONE_QVM_FK:        "qvmfk",
	ZONE_QVM_DUBAI:     "qvmdb",
	ZONE_QVM_MUMBAI:    "qvmmb",
	ZONE_QVM_TOKYO:     "qvmtky",

	ZONE_KE_XQ: "kexq",
	ZONE_KE_BQ: "kebq",
	ZONE_KE_JQ: "kejq",
	ZONE_KE_DQ: "kedq",
}

func NewZone(name string) (Zone, error) {
	if zone, ok := zoneNameMap[name]; ok {
		return zone, nil
	} else {
		// For compatibility of old code
		z, err := strconv.Atoi(strings.TrimPrefix(name, "z"))
		if err != nil {
			return ZONE_NB, err
		}
		zone = Zone(z)
		if !zone.IsValid() {
			return ZONE_NB, fmt.Errorf("invalid zone: %s", name)
		}
		return zone, nil
	}
}

func (z Zone) String() string {
	return strconv.Itoa(int(z))
}

func (z Zone) Name() string {
	return zoneNameMapRev[z]
}

func (z Zone) Humanize() string {
	switch z {
	case ZONE_NB:
		return "宁波机房"
	case ZONE_BC:
		return "昌平机房"
	case ZONE_LAC:
		return "北美机房"
	case ZONE_GZ:
		return "华南机房"
	case ZONE_SG:
		return "新加坡机房"

	case ZONE_KIRK_BQ:
		return "容器云华北1区"
	case ZONE_KIRK_NQ:
		return "容器云华东1区"
	case ZONE_KIRK_LAC:
		return "容器云北美1区"
	case ZONE_KIRK_GQ:
		return "容器云华南1区"
	case ZONE_KIRK_XQ:
		return "容器云华东2区"

	// 融合计算
	case ZONE_KYLIN_BJ_1:
		return "北京一区"
	case ZONE_KYLIN_SH_1:
		return "上海一区"
	case ZONE_KYLIN_GZ_0:
		return "广州零区"
	case ZONE_KYLIN_GZ_1:
		return "广州一区"
	case ZONE_KYLIN_GZ_2:
		return "广州二区"
	case ZONE_KYLIN_HK_1:
		return "香港一区"
	case ZONE_KYLIN_TORONTO_1:
		return "北美一区"
	case ZONE_KYLIN_SINGAPORE_1:
		return "新加坡一区"

	// Fusion
	case ZONE_FUSION_CHINA:
		return "中国大陆"
	case ZONE_FUSION_AMEU:
		return "欧洲/北美洲"
	case ZONE_FUSION_ASIA:
		return "亚洲（除中国/印度/东南亚）"
	case ZONE_FUSION_SEA:
		return "亚洲（东南亚/印度）"
	case ZONE_FUSION_SA:
		return "南美洲"
	case ZONE_FUSION_OC:
		return "大洋洲与其他"
	case ZONE_FUSION_FOREIGN:
		return "海外未分区"
	case ZONE_FUSION_NOZONE:
		return "网宿海外不分区"

		// QVM
	case ZONE_QVM_QD:
		return "华北一区"
	case ZONE_QVM_BJ:
		return "华北二区"
	case ZONE_QVM_ZJK:
		return "华北三区"
	case ZONE_QVM_HHHT:
		return "华北五区"
	case ZONE_QVM_HZ:
		return "华东一区"
	case ZONE_QVM_SH:
		return "华东二区"
	case ZONE_QVM_SZ:
		return "华南一区"
	case ZONE_QVM_HK:
		return "香港一区"
	case ZONE_QVM_SINGAPORE:
		return "亚太东南一区"
	case ZONE_QVM_SYDNEY:
		return "亚太东南二区"
	case ZONE_QVM_KL:
		return "亚太东南三区"
	case ZONE_QVM_JAKARTA:
		return "亚太东南五区"
	case ZONE_QVM_VIRGINIA:
		return "美国东部一区"
	case ZONE_QVM_SV:
		return "美国西部一区"
	case ZONE_QVM_FK:
		return "欧洲中部一区"
	case ZONE_QVM_DUBAI:
		return "中东东部一区"
	case ZONE_QVM_MUMBAI:
		return "亚太南部一区"
	case ZONE_QVM_TOKYO:
		return "亚太东北一区"

	// KE
	case ZONE_KE_XQ:
		return "华东一区"
	case ZONE_KE_BQ:
		return "华北一区"
	case ZONE_KE_JQ:
		return "华东二区"
	case ZONE_KE_DQ:
		return "华南一区"

	}
	return fmt.Sprintf("未命名(%d)", z)
}

func (z Zone) IsValid() bool {
	_, ok := zoneNameMapRev[z]
	return ok
}
