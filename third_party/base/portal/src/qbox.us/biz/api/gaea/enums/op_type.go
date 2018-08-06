package enums

const (
	OpTypeLogin    OpType = 0
	OpTypeAccount  OpType = 1
	OpTypeBucket   OpType = 2
	OpTypeDomain   OpType = 3
	OpTypeKey      OpType = 4
	OpTypeLogFile  OpType = 5
	OpTypeFop      OpType = 6
	OpTypeBwList   OpType = 7
	OpTypeRecharge OpType = 8
	OpTypeThird    OpType = 9  // 第三账户相关
	OpTypeService  OpType = 10 // 第三方服务市场
	OpTypeUFOP     OpType = 11 // UFOP相关
	OpTypePili     OpType = 12 // pili相关
)

type OpType uint32 // 中文对照表: src/website/assets/js/setting/oplog.js [cnmap.opType]

func OpTypeList() []OpType {
	return []OpType{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
}

func (o OpType) Valid() bool {
	return o >= OpTypeLogin && o <= OpTypePili
}

func (o OpType) Humanize() string {
	switch o {
	case OpTypeLogin:
		return "登录"
	case OpTypeAccount:
		return "账户"
	case OpTypeBucket:
		return "空间"
	case OpTypeDomain:
		return "域名"
	case OpTypeKey:
		return "密钥"
	case OpTypeLogFile:
		return "日志功能"
	case OpTypeFop:
		return "数据处理"
	case OpTypeBwList:
		return "黑白名单"
	case OpTypeRecharge:
		return "充值"
	case OpTypeThird:
		return "第三方账户"
	case OpTypeService:
		return "服务市场"
	case OpTypeUFOP:
		return "应用开发"
	case OpTypePili:
		return "直播服务"
	default:
		return "错误的OpType值"
	}
}
