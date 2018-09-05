package fusion

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"qbox.us/api/cdn.v1/model"

	"github.com/qiniu/ctype"
	// ldb "qiniu.com/fusion.v2/fusionline/db"
)

const (
	MGR_DOMAIN = ".qiniudns.com"
)

const (
	MAX_CACHE_TIME = 31536000 // 一年(s)
)

type Domain struct {
	// 用户域名。
	Name string `json:"name" bson:"name"`

	// 域名反写, 用来做域名后缀搜索索引
	ReverseName string `json:"reverseName" bson:"reverseName"`

	// 用户 id。
	Uid uint32 `json:"uid" bson:"uid"`
	// 域名类型
	Type Type `json:"type" bson:"type"`

	// fusion 生成的域名。
	// 用户需要把域名 CNAME 到这个域名。
	Cname string `json:"cname" bson:"cname"`

	// fusion 生成的边缘节点回源域名。
	// 此域名通过 A 记录或者 CNAME 记录连接到用户源站或者中间源。
	SourceCname string `json:"-" bson: "sourceCname"`

	// 回源 Host。
	SourceHost string `json:"sourceHost" bson:"sourceHost"`
	// 源站类型。
	SourceType SourceType `json:"sourceType" bson:"sourceType"`
	// 源站 ip。源是 ip 时不为空。
	SourceIPs []string `json:"sourceIPs,omitempty" bson:"sourceIPs"`
	// 源站域名。源是 domain 时不为空。
	SourceDomain string `json:"sourceDomain,omitempty" bson:"sourceDomain"`
	// 源站的七牛 Bucket。源是 qiniuBucket 时不为空。
	SourceQiniuBucket string `json:"sourceQiniuBucket,omitempty" bson:"sourceQiniuBucket"`
	// 高级源站配置，当源站非七牛时可以使用。
	AdvancedSources []AdvancedSource `json:"advancedSources,omitempty" bson:"advancedSources"`
	// 回源 URL 协议。
	SourceURLScheme string `json:"sourceURLScheme,omitempty" bson:"sourceURLScheme"`

	// url rewrites.
	URLRewrites []model.Rewrite `json:"urlRewrites,omitempty" bson:"urlRewrites"`

	// 添加自定义响应 Header。
	AddRespHeader http.Header `json:"addRespHeader,omitempty" bson:"addRespHeader"`

	// 回源重试 Code。
	SourceRetryCodes []int `json:"sourceRetryCodes" bson:"sourceRetryCodes"`

	// 最大回源速度，每秒字节数。
	MaxSourceRate int `json:"maxSourceRate" bson:"maxSourceRate"`
	// 最大回源并发量。
	MaxSourceConcurrency int `json:"maxSourceConcurrency" bson:"maxSourceConcurrency"`

	// 用户提供的测试 URL Path。
	TestURLPath string `json:"testURLPath" bson:"testURLPath"`

	// 中间源。
	MidSource MidSource `json:"midSource" bson:"midSource"`

	// 线路 Id。
	LineId string `json:"lineId" bson:"lineId"`
	// 线路供应商提供的 cname。
	LineCname string `json:"-" bson:"lineCname"`

	// 域名创建时间。
	CreateAt time.Time `json:"createAt" bson:"createAt"`
	// 域名的最后修改时间。
	ModifyAt time.Time `json:"modifyAt" bson:"modifyAt"`

	// 当前对域名的操作。
	Operation Operation `json:"operation" bson:"operation"`
	// 域名操作的状态。
	State OperatingState `json:"state" bson:"state"`
	// 域名操作状态描述。
	StateDesc string `json:"stateDesc" bson:"stateDesc"`

	// 域名在 CDN 上的操作状态。
	CDNStates []CDNState `json:"-" bson:"cdnStates"`
	// 域名同步CDN配置状态
	SyncCDNStates []CDNState `json:"-" bson:"synccdnStates"`

	// 域名在中间源的操作状态。
	MidSourceState MidSourceState `json:"-" bson:"midSourceState"`

	// 记录 Cname => LineCname 的 DNS 操作状态。
	CnameDNSStates []DNSState `json:"-" bson:"cnameDNSState"`
	// 记录 SourceCname => SourceIPs/SourceDomain 的 DNS 操作状态。
	SourceCnameDNSStates []DNSState `json:"-" bson:"sourceCnameDNSState"`

	// 迁移灰度[1, 100]
	WeightLevel int `json:"weightLevel" bson:"weightLevel"`

	// 黑/白名单
	RefererType RefererType `json:"refererType" bson:"refererType"`
	// 黑白名单列表
	RefererValues []string `json:"refererValues" bson:"refererValues"`
	// 是否允许空Referer
	NullReferer *bool `json:"nullReferer"`

	IgnoreQueryStr *bool          `json:"ignoreParam" bson:"ignoreParam"` // 是否忽略URL中参数部分（URL中"?"后面的部分）
	CacheControls  []CacheControl `json:"cacheControls" bson:"cacheControls"`

	// 是否为泛域名
	GotWildcard bool `json:"gotWildcard" bson:"gotWildcard"`

	// 备案号
	RegisterNo string `json:"registerNo" bson:"registerNo"`

	// 时间戳防盗链开关
	TimeACL *bool `json:"timeACL" bson:"timeACL"`
	// 时间戳防盗链keys
	TimeACLKeys []string `json:"timeACLKeys" bson:"timeACLKeys"`
	// 备注
	Notes       map[NoteType]string `json:"notes" bson:"notes"`
	MutiDomains bool                `json:"mutiDomains"`
	DnsGrey     bool                `json:"dnsGrey"`
	SwitchState string              `json:"switchState"`
	GenCname    bool                `json:"genCname"`

	// CDN同步状态
	SyncCDNState bool `json:"sync_cdn_state" bson:"sync_cdn_state"`

	// 能力
	Abilities map[AbilityType]Ability `json:"abilities" bson:"abilities"`

	// 父域名(只对泛子域名有效)
	PareDomain string `json:"pareDomain" bson:"pareDomain"`
}

type DomainRegionLine struct {
	//域名
	Name string `json:"name" bson:"name"`
	// key为baseLine.id，value为 范围-运营商 组合
	BaseLines  map[string][]string `json:"baseLines" bson:"baseLines"`
	CreateTime time.Time           `json:"createTime" bson:"createTime"`
	UpdateTime time.Time           `json:"updateTime" bson:"updateTime"`
	// 是否已经与dnspod同步
	DNSSynced       bool   `json:"dnsSynced" bson:"dnsSynced"`
	DefaultBaseLine string `json:"defaultBaseLine" bson:"defaultBaseLine"`
}

type Type string

const (
	TYPE_NORMAL   Type = "normal"   // 普通域名
	TYPE_PAN      Type = "pan"      // 用户泛子域名，只能修改源站
	TYPE_TEST     Type = "test"     // 七牛泛子域名，测试域名
	TYPE_WILDCARD Type = "wildcard" // 泛域名
)

func (t Type) IsValidType() bool {
	switch t {
	case TYPE_NORMAL, TYPE_PAN, TYPE_WILDCARD, TYPE_TEST:
		return true
	case "":
		return true
	}
	return false
}

// [a-zA-Z0-9] || [-.]
func IsDomainName(s string) bool {
	return ctype.IsType(ctype.DOMAIN_CHAR|ctype.COLON, s)
}

type MidSource struct {
	Enabled bool     `json:"enabled" bson:"enabled"`
	Addrs   []string `json:"addrs" bson:"addrs"`
}

type AdvancedSource struct {
	Addr   string `json:"addr" bson:"addr"`
	Weight int    `json:"weight" bson:"weight"`
	Backup bool   `json:"backup" bson:"backup"`
}

type DomainErr struct {
	Id       string    `json:"id" bson:"_id"`
	Name     string    `json:"name" bson:"name"`
	CreateAt time.Time `json:"createAt" bson:"createAt"`
	ErrType  ErrType   `json:"errType" bson:"errType"`
	ErrInfo  string    `json:"errInfo" bson:"errInfo"`
}

type ErrType string

const (
	ERR_TYPE_USER_CONFLICT     ErrType = "userConflict"
	ERR_TYPE_PLATFORM_CONFLICT ErrType = "platformConflict"
	ERR_TYPE_INTERNAL          ErrType = "internal"
	ERR_TYPE_NOICP             ErrType = "no icp"
)

// CDN缓存策略 API结构
type CacheControl struct {
	URLPattern string  `json:"urlPattern" bson:"urlpattern"` // URL匹配规则（正则表达式）
	CacheTime  int64   `json:"cacheTime" bson:"cachetime"`   // 缓存时间，为正数时表示缓存秒数，0 不缓存 -1 遵循源站
	Time       *int64  `json:"time" bson:"time"`             // 新的配置数据,用于向前端显示
	TimeUnit   *int    `json:"timeunit" bson:"timeunit"`     // 新的配置数据,用于向前端显示
	Type       *string `json:"type" bson:"type"`
	Rule       *string `json:"rule" bson:"rule"`
}

type CDNState struct {
	CDNProvider CDNProvider    `json:"cdnProvider" bson:"cdnProvider"`
	TaskId      string         `json:"taskId" bson:"taskId"`
	State       OperatingState `json:"state" bson:"state"`
	StateDesc   string         `json:"stateDesc" bson:"stateDesc"`
}

type MidSourceState struct {
	TaskId string         `json:"taskId" bson:"taskId"`
	State  OperatingState `json:"state" bson:"state"`
}

type DNSState struct {
	DNSProvider DNSProvider    `json:"dnsProvider" bson:"dnsProvider"`
	TaskId      string         `json:"taskId" bson:"taskId"`
	State       OperatingState `json:"state" bson:"state"`
}

const (
	DNSProviderDnspod = "dnspod"
)

const (
	DNSRecordTypeCNAME = "CNAME"
	DNSRecordTypeA     = "A"
)

const (
	RecordValueNil = "nil.a.com"
)

type Operation string

const (
	OperationCreateDomain          Operation = "create_domain"
	OperationDeleteDomain                    = "delete_domain"
	OperationModifySource                    = "modify_source"
	OperationModifyReferer                   = "modify_referer"
	OperationModifyCache                     = "modify_cache"
	OperationModifyNote                      = "modify_note"
	OperationModifyAbility                   = "modify_ability"
	OperationTransfer                        = "transfer"
	OperationRecord                          = "record"
	OperationSwitch                          = "switch"
	OperationBackup                          = "backup"
	OperationSyncFromCDN                     = "sync_from_cdn"
	OperationVerifyCDN                       = "verify_cdn"
	OperationSwitchMidsrc                    = "switch_midsrc"
	OperationFreezeDomain                    = "freeze_domain"
	OperationUnFreezeDomain                  = "unfreeze_domain"
	OperationRegionAdjust                    = "region_adjust"
	OperationModifyTimeACL                   = "modify_timeacl"
	OperationModifyHttpsCrt                  = "modify_https_crt"
	OperationTransferUser                    = "transfer_user"
	OperationInternalCorrectSource           = "correct_source"
)

type OperatingState string

const (
	OperatingStateProcessing OperatingState = "processing"
	OperatingStateFailed                    = "failed"
	OperatingStateSuccess                   = "success"
	OperatingStateFrozen                    = "frozen"
	OperatingStateDeleted                   = "deleted"
)

func ValidOperatingState(s string) (OperatingState, bool) {
	state := OperatingState(s)
	switch state {
	case OperatingStateProcessing, OperatingStateFailed, OperatingStateSuccess, OperatingStateFrozen, OperatingStateDeleted:
		return state, true
	}
	return "", false
}

type RefreshState string

const (
	RefreshStateProcessing RefreshState = "processing"
	RefreshStateFailed                  = "failed"
	RefreshStateFailure                 = "failure"
	RefreshStateSuccess                 = "success"
	RefreshStateUnknow                  = "unknow"
	RefreshStateWaitting                = "waitting"
	RefreshStateFetch                   = "fetch"
)

type SourceType string

const (
	SourceTypeDomain      SourceType = "domain"
	SourceTypeIP                     = "ip"
	SourceTypeQiniuBucket            = "qiniuBucket"
	SourceTypeAdvanced               = "advanced"
)

func ValidSourceType(s string) (SourceType, bool) {
	sourceType := SourceType(s)
	switch sourceType {
	case SourceTypeDomain, SourceTypeIP, SourceTypeQiniuBucket, SourceTypeAdvanced:
		return sourceType, true
	}
	return "", false
}

type GeoCover string

const (
	GeoCoverChina   GeoCover = "china"
	GeoCoverForeign          = "foreign"
	GeoCoverGlobal           = "global"
)

func ValidGeoCover(s string) (GeoCover, bool) {
	geoCover := GeoCover(s)
	switch geoCover {
	case GeoCoverChina, GeoCoverForeign, GeoCoverGlobal:
		return geoCover, true
	}
	return "", false
}

type Protocol string

const (
	ProtocolHTTP  = "http"
	ProtocolHTTPS = "https"
)

func ValidProtocl(s string) (Protocol, bool) {
	protocol := Protocol(s)
	switch protocol {
	case ProtocolHTTP, ProtocolHTTPS:
		return protocol, true
	}
	return "", false
}

type Platform string

const (
	PlatformWeb      Platform = "web"
	PlatformDownload          = "download"
	PlatformVOD               = "vod"
)

func ValidPlatform(s string) (Platform, bool) {
	platform := Platform(s)
	switch platform {
	case PlatformWeb, PlatformDownload, PlatformVOD:
		return platform, true
	}
	return "", false
}

type PlatformLevel uint64

const (
	PlatformLevelNone PlatformLevel = 0
	PlatformLevel10                 = 10
	PlatformLevel20                 = 20
)

func ValidPlatformLevel(i uint64) (PlatformLevel, bool) {
	level := PlatformLevel(i)
	switch level {
	case PlatformLevelNone, PlatformLevel10, PlatformLevel20:
		return level, true
	}
	return 0, false
}

const (
	StatusEnable = "enable"
	StatusUnable = "unable"
)

const (
	AlarmRefreshRecord  = "AlarmRefreshRecord"
	AlarmPrefetchRecord = "AlarmPrefetchRecord"
	AlarmUserLimit      = "AlarmUserLimit"
	AlarmCdnLimit       = "AlarmCdnLimit"
)

const (
	IsDirYes = "yes"
	IsDirNo  = "no"
)

type OpType string

const (
	OP_TYPE_UPDATE_SOURCE  OpType = "update_source"
	OP_TYPE_UPDATE_REFERER OpType = "update_referer"
	OP_TYPE_UPDATE_CACHE   OpType = "update_cache"
	OP_TYPE_UPDATE_TIMEACL OpType = "update_timeacl"
)

type DomainTask struct {
	Domain      string `bson:"domain" json:"domain"`
	TaskId      string `bson:"taskId" json:"taskId"`
	NextOpType  OpType `bson:"next_op_type" json:"next_op_type"`
	NextOpParam string `bson:"next_op_param" json:"next_op_param"` // param json string
}

type DNSProvider string

const (
	DNSProviderDNSPod DNSProvider = "dnspod"
)

type HttpsCrt struct {
	Name         string    `json:"name" bson:"name"`
	Uid          uint32    `json:"uid" bson:"uid"`
	ServerKey    string    `json:"serverKey" bson:"serverKey"`
	ServerCrt    string    `json:"serverCrt" bson:"serverCrt"`
	FirstMidCrt  string    `json:"firstCrt" bson:"firstCrt"`
	SecondMidCrt string    `json:"secondCrt" bson:"secondCrt"`
	CaCrt        string    `json:"caCrt" bson:"caCrt"` // 暂时作为用户上传的证书字段
	CreateAt     time.Time `json:"createAt" bson:"createAt"`
	ModifyAt     time.Time `json:"modifyAt" bson:"modifyAt"`
	UseOwnChain  bool      `json:"useOwnChain" bson:"useOwnChain"`
}

func (h HttpsCrt) GetCrt(typ HttpsCrtType) string {
	switch typ {
	case HTTPS_CRT_TYPE_SVR_KEY:
		return h.ServerKey
	case HTTPS_CRT_TYPE_SVR_CRT:
		return h.ServerCrt
	case HTTPS_CRT_TYPE_CA_CRT:
		return h.CaCrt
	}
	return ""
}

func (h HttpsCrt) GetCrtName(typ HttpsCrtType) string {
	switch typ {
	case HTTPS_CRT_TYPE_SVR_KEY:
		return "server.key"
	case HTTPS_CRT_TYPE_SVR_CRT:
		return "server.crt"
	case HTTPS_CRT_TYPE_CA_CRT:
		return "ca.crt"
	case HTTPS_CRT_TYPE_ALL:
		return "crt.zip"
	}
	return ""
}

type NoteType string

const (
	// 缓存设置
	NOTETYPE_CACHE NoteType = "cache"
	// UA防盗链
	NOTETYPE_REFERER_UA NoteType = "referer_ua"
	// referer防盗链
	NOTETYPE_REFERER_REFERER NoteType = "referer_referer"
	// 时间戳防盗链
	NOTETYPE_REFERER_TIMESTAMP NoteType = "referer_timestamp"
	// ip过滤
	NOTETYPE_REFERER_IPFILTER NoteType = "referer_ipfilter"
	// 回源配置
	NOTETYPE_SOURCE NoteType = "source"
	// 限速配置
	NOTETYPE_SPEED_LIMIT NoteType = "speed_limit"
	// 防攻击配置
	NOTETYPE_ANTI_ATTACK NoteType = "anti_attack"
	// 动态加速
	NOTETYPE_DYNAMIC_ACC NoteType = "dynamic_acc"
	// 其他特殊配置
	NOTETYPE_OTHER NoteType = "other"
	// 域名说明
	NOTETYPE_NOTE NoteType = "note"
)

func (n NoteType) IsValid() bool {
	switch n {
	case NOTETYPE_CACHE, NOTETYPE_NOTE, NOTETYPE_OTHER, NOTETYPE_SOURCE,
		NOTETYPE_ANTI_ATTACK, NOTETYPE_DYNAMIC_ACC, NOTETYPE_REFERER_IPFILTER,
		NOTETYPE_REFERER_REFERER, NOTETYPE_REFERER_TIMESTAMP, NOTETYPE_REFERER_UA,
		NOTETYPE_SPEED_LIMIT:
		return true
	default:
		return false
	}
	return false
}

// 域名能力类型
type AbilityType string

const (
	// 日志粒度
	ABILITY_TYPE_LOG_GRANULARITY AbilityType = "log_granularity"
	// 日志延迟
	ABILITY_TYPE_LOG_LATENCY AbilityType = "log_latency"
	// 流量API数据粒度
	ABILITY_TYPE_TRANSFER_GRANULARITY AbilityType = "transfer_granularity"
	// 流量API数据延迟
	ABILITY_TYPE_TRANSFER_LATENCY   AbilityType = "transfer_latency"
	ABILITY_TYPE_LOG_AGGRESORT                  = "log_aggresort"
	ABILITY_TYPE_LOG_AGGRESORTV2                = "log_aggresortv2"
	ABILITY_TYPE_LOG_AGGRESORTV3                = "log_aggresortv3"
	ABILITY_TYPE_LOG_CODE_ANALYSIS              = "code_analysis"
	ABILITY_TYPE_LOG_COVER_ANALYSIS             = "cover_analysis"
	ABILITY_TYPE_LOG_AGGRETIMER                 = "AggreTimer"
)

func (a AbilityType) IsValid() bool {
	switch a {
	case ABILITY_TYPE_LOG_GRANULARITY, ABILITY_TYPE_LOG_LATENCY,
		ABILITY_TYPE_TRANSFER_GRANULARITY, ABILITY_TYPE_TRANSFER_LATENCY, ABILITY_TYPE_LOG_AGGRESORT, ABILITY_TYPE_LOG_AGGRESORTV2, ABILITY_TYPE_LOG_AGGRESORTV3,
		ABILITY_TYPE_LOG_CODE_ANALYSIS, ABILITY_TYPE_LOG_COVER_ANALYSIS, ABILITY_TYPE_LOG_AGGRETIMER:
		return true
	default:
		return false
	}
	return false
}

// 域名能力值
type AbilityValue string

// 具体能力值映射到自己配置的数值上
const (
	ABILITY_A_PLUS  AbilityValue = "A+"
	ABILITY_A       AbilityValue = "A"
	ABILITY_A_MINUS AbilityValue = "A-"
	ABILITY_B_PLUS  AbilityValue = "B+"
	ABILITY_B       AbilityValue = "B"
	ABILITY_B_MINUS AbilityValue = "B-"
	ABILITY_C_PLUS  AbilityValue = "C+"
	ABILITY_C       AbilityValue = "C"
	ABILITY_C_MINUS AbilityValue = "C-"
)

type Ability struct {
	Config       string       `bson:"config" json:"config"`
	Enable       bool         `bson:"enable" json:"enable"`
	AbilityValue AbilityValue `bson:"abilityValue" json:"abilityValue"`
}

func (a Ability) IsValid() bool {
	switch a.AbilityValue {
	case "":
		return true
	case ABILITY_A, ABILITY_A_PLUS, ABILITY_A_MINUS,
		ABILITY_B, ABILITY_B_PLUS, ABILITY_B_MINUS,
		ABILITY_C, ABILITY_C_PLUS, ABILITY_C_MINUS:
		return true
	}
	return false
}

const (
	BASE string = "0123456789abcdefghijklmnopqrstuvwxyz"
)

func GetRandomUUID() (u string) {
	u = strconv.FormatInt(time.Now().Unix(), 36)
	if len(u) > 4 {
		u = u[:4]
	}

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 5; i++ {
		u += string(BASE[rand.Int63()%36])
	}

	return u
}

func GetWildcardTestPrefix() string {
	return "fusion_test_" + GetRandomUUID()
}
