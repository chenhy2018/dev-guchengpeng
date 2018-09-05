package fusioncdn

import "time"

type Task struct {
	Id                string             `json:"id" bson:"_id"`
	MqId              string             `json:"mqId" bson:"mqId"` //  mq队列中messgae的id
	CdnProvider       string             `json:"cdnProvider" bson:"cdnProvider"`
	Domain            string             `json:"domain" bson:"domain"`
	CallbackUrl       string             `json:"callbackURL" bson:"callbackURL"`
	Status            string             `json:"status" bson:"status"`
	CMD               string             `json:"cmd" bson:"cmd"`
	Manual            bool               `json:"manual" bson:"manual"`
	ManualType        string             `json:"manualType"`
	SwitchManual      bool               `json:"switchManual" bson:"switchManual"`
	UpdateTime        time.Time          `json:"updateTime" bson:"updateTime"`
	WangsuUpdateState string             `bson: "wangsuUpdateState"`
	Configuration     QiniuConfiguration `json:"config" bson:"config"`
	RequestId         string             `bson:"requestId"`
	ExtroInfo         string             `json:"extroInfo"`
}

const (
	ErrOK = 200
)

// CDN缓存策略
type CacheControl struct {
	URLPattern string `bson:"urlpattern" json:"urlPattern"` // URL匹配规则（正则表达式）
	CacheTime  int64  `bson:"cachetime" json:"cacheTime"`   // 缓存时间，为正数时表示缓存秒数，0 不缓存 -1 遵循源站
}

type QiniuConfiguration struct {
	SourceType      string         `bson:"sourceType" json:"sourceType"`
	SourceIPs       string         `bson:"sourceIPs" json:"sourceIPs"`
	SourceDomain    string         `bson:"sourceDomain" json:"sourceDomain"`
	SourceURLScheme string         `bson:"sourceURLScheme" json:"sourceURLScheme"`
	LineCname       string         `json:"lineCname" bson:"lineCname"`
	Platform        string         `bson:"platform" json:"platform"`
	GeoCover        string         `bson:"geoCover"  json:"geoCover"`
	PlatformLevel   uint64         `json:"platformLevel"`
	QiniuPrivate    bool           `bson:"qiniuPrivate" json:"qiniuPrivate"`
	Protocol        string         `bson:"protocol" json:"protocol"`
	TestURL         string         `bson:"testURL" json:"testURL"`
	RegisterNo      string         `bson:"registerNo" json:"registerNo"`
	RefererType     string         `json:"refererType"`
	RefererValue    string         `json:"refererValue"`
	NullReferer     bool           `json:"nullReferer"`                    // 是否允许空Referer, true 允许, false 不允许
	ForwardHost     string         `json:"forwardHost"`                    // CDN回源HTTP请求中header Host的值，默认为访问域名
	IgnoreQueryStr  bool           `bson:"ignoreParam" json:"ignoreParam"` // 是否忽略URL中参数部分（URL中"?"后面的部分）
	CacheControls   []CacheControl `json:"cacheControls"`
	TimeACL         bool           `json:"timeACL" bson:"timeACL"`
	TimeACLKeys     []string       `json:"timeACLKeys" bson:"timeACLKeys"`
}

const (
	TaskCommandcdnCreate string = "cdn_create"
	TaskCommandcdnRemove string = "cdn_remove"
	TaskCommandcdnUpdate string = "cdn_update"
)

const (
	// 对应DomainInfo.Status
	ConfEnabled  string = "enable"  // CDN配置启用状态
	ConfDisabled string = "disable" // CDN配置禁用状态
)

type DomainInfo struct {
	CdnProvider string                 `json:"cdnProvider"`
	Domain      string                 `json:"domain"`
	DomainTag   string                 `json:"domainTag"`
	Cname       string                 `json:"cname"`
	Status      string                 `json:"status"`
	ExtroInfo   map[string]interface{} `json:"extroInfo"`
	Conf        QiniuConfiguration
}

const (
	CDNSTATENOTFOUND   = 0
	CDNSTATEEXAMING    = 1
	CDNSTATEREJECTED   = 2
	CDNSTATEAPPROVED   = 3
	CDNSTATEPROCESSING = 4
	CDNSTATEONLINE     = 5
	CDNSTATEOFFLINE    = 6
	CDNSTATEDELETED    = 7
)

const (
	STATEDONE   = "done"
	STATEFAILED = "fail"
)

const (
	CodeServerErr     = 500
	CodeServerOK      = 200
	CodeInvalidParams = 400
)

const (
	CodeOK                             = 200
	CodeDeleted                        = 200
	CodeError                          = 400000
	CodeAuthErr                        = 400001
	CodeInputErr                       = 400002
	RegisterNoGetErr                   = 400003
	CodeInvalidParam                   = 400004
	CodeServerOpFailErr                = 500099
	CodeResourceNotExist               = 600012
	CodeResourceExisted                = 600014
	CodeServerInternalErr              = 500000
	CodeAddDomainAndConfigNotRefFailed = 500008
	CodeRepeatBindErr                  = 500009
	CodeInputScheduleIdErr             = 500010
	CodeExistUnfinishedOperate         = 500011
	CodeUidInvalid                     = 500012
	TC_DECODE_FAIL                     = 500013
	CodeLetvCreateDomainErr            = 500015
	CodeTencentDomainConflict          = 400016
	CodeTencentCreateDomainErr         = 500024
	CodeWangsuFAILDOMAINEXIST          = 400017
	CodeWangsuFAILDOMAINATTRIBUTEERROR = 500018
	CodeWangsuFailed                   = 500019
	CodeCdnpubDBFalied                 = 500020
	CodeGetTasksDbFailed               = 500021
	CodeDilianCreateDomainErr          = 500022
	CodeDilianSucc                     = 500023
	CodeCdnCreateFailed                = 500040
	CodeUnaAccomplishedTask            = 400032 // 重试了未完成的人工任务
	CodeReAccessCheckFail              = 400033
)

const (
	Tc_SUCC                        = 200
	Tc_FAIL                        = 500001
	Tc_FAIL_DOMAIN_ATTRIBUTE_ERROR = 500002
	Tc_FAIL_DECODE                 = 500003
	TxMarshalFail                  = 500004
	TcDomainNameErr                = 400028
	// "(20004)未备案 cdn audit no icp[cdn audit no icp]"，实际检查中配合这个code检查了返回的message中是否包含cdn audit no icp
	TcNoICPErr           = 400077
	TcOtherAccountDomain = 400076
)

const (
	WsDomainNameErr                = 400021
	WsMarshalConfFail              = 400018
	WsNewRequestFail               = 400019
	WsOtherAccountDomain           = 400020
	WsCannotChangePlatForm         = 400081
	WsClientDoFail                 = 500405
	WS_FAIL                        = 500400
	WS_FAIL_DOMAIN_ATTRIBUTE_ERROR = 500401
	WS_FAIL_DECODE                 = 500402
	WS_FAIL_DOMAIN_EXIST           = 500403
	WS_SYSTEM_ERROR                = 500404
	WsBatchQueryDetailFail         = 500405
	WS_SUCC                        = 200
)

const (
	LtSuccCode           = 200
	LtFailedCode         = 500000
	LtMarshalConfFail    = 400020
	LtDomainConflict     = 400024
	LtOtherAccountDomain = 400025
)

const (
	DlFail               = 500025
	DlDecodeFail         = 500026
	DlSucc               = 200
	DlDomainConflict     = 400027
	DlDomainNameErr      = 400029
	DlXMLError           = 400031 // TODO 对应InappropriateXML，之后需要将DlDomainNameErr从里面区别出来
	DlOtherAccountDomain = 400030
)

const (
	KwAssignConfErr      = 500028
	KwMarshalErr         = 500029
	KwNewRequestErr      = 500030
	KwDefaultClientErr   = 500031
	KwDecodeResposeErr   = 500032
	KwCreateSuccess      = 200
	KwCreateFail         = 500033
	KwDomainConflict     = 400033
	KwDomainNameErr      = 400034
	KwOtherAccountDomain = 400035
)

const (
	BsyHTTPReqErr         = 500050 // HTTP请求构造室出错或者请求时返回err
	BsyHTTPReqFailed      = 500051 // HTTP请求成功，但返回的StatusCode非200
	BsyFail               = 500052 // 接口调用判断为失败
	BsyResultDecodeErr    = 500053 // CDN接口返回数据无法json decode
	BsyConflict           = 400051 // 域名冲突
	BsyOtherAccountDomain = 400052 // 域名属于其他账户
	BsyOriginErr          = 400050 // 请求参数中SourceType / SourceIPs / SourceDomain不正确
)

const (
	StateProcessing = "processing"
	StateSuccess    = "success"
	StateFailed     = "failed"
	StateRetried    = "retried"
	StateAbort      = "aborted"
)

const DESKEY = "11111111"
