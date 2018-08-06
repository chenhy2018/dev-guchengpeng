package fusion

import (
	"net/http"
	"time"

	"qbox.us/api/cdn.v1/model"
	"qbox.us/api/tblmgr"
)

type ISetCode interface {
	setCode(code int)
}

type CreateDomainReq struct {
	Uid               uint32                  `json:"uid"`
	Type              Type                    `json:"type"`
	SourceHost        string                  `json:"sourceHost"`
	SourceType        SourceType              `json:"sourceType"`
	SourceIPs         []string                `json:"sourceIPs"`
	SourceDomain      string                  `json:"sourceDomain"`
	SourceQiniuBucket string                  `json:"sourceQiniuBucket"`
	AdvancedSources   []AdvancedSource        `json:"advancedSources" bson:"advancedSources"`
	TestURLPath       string                  `json:"testURLPath"`
	Platform          Platform                `json:"platform"`
	GeoCover          GeoCover                `json:"geoCover"`
	Protocol          Protocol                `json:"protocol"`
	QiniuPrivate      bool                    `json:"qiniuPrivate"`
	LineId            string                  `json:"lineId"`
	PlatformLevel     PlatformLevel           `json:"platformLevel"`
	RefererType       RefererType             `json:"refererType"`
	RefererValue      string                  `json:"refererValue"`
	SkipCheckSource   bool                    `json:"skipCheckSource"`
	NullReferer       *bool                   `json:"nullReferer"`
	IgnoreQueryStr    *bool                   `json:"ignoreParam"`
	CacheControls     []CacheControl          `json:"cacheControls"`
	TimeACL           *bool                   `json:"timeACL"`
	TimeACLKeys       []string                `json:"timeACLKeys"`
	Notes             map[NoteType]string     `json:"notes"`
	ServerKey         string                  `json:"serverKey" bson:"serverKey"`
	ServerCrt         string                  `json:"serverCrt" bson:"serverCrt"`
	Abilities         map[AbilityType]Ability `json:"abilities"`
	DisableIcpCheck   bool                    `json:"disableIcpCheck"`
	RegisterNo        string                  `json:"registerNo"`
}

type CommonRes struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func (r *CommonRes) setCode(code int) {
	r.Code = code
	r.Error = ErrorText[code]
}

type DnsCallbackReq struct {
	TaskId     string         `json:"taskId"`
	Result     OperatingState `json:"result"`
	ResultDesc string         `json:"resultDesc"`
}

type CdnCallbackReq struct {
	TaskId     string         `json:"taskId"`
	Cname      string         `json:"cname"`
	Result     OperatingState `json:"result"`
	ResultDesc string         `json:"resultDesc"`
}

type ModifyDomainSourceReq struct {
	SourceHost           string           `json:"sourceHost"`
	SourceType           SourceType       `json:"sourceType"`
	SourceIPs            []string         `json:"sourceIps"`
	SourceDomain         string           `json:"sourceDomain"`
	SourceQiniuBucket    string           `json:"sourceQiniuBucket"`
	AdvancedSources      []AdvancedSource `json:"advancedSources"`
	TestURLPath          string           `json:"testURLPath"`
	Uid                  uint32           `json:"uid"`
	SkipSourceHostCheck  bool             `json:"skipSourceHostCheck"`
	URLRewrites          []model.Rewrite  `json:"urlRewrites"`
	AddRespHeader        http.Header      `json:"addRespHeader"`
	SourceRetryCodes     []int            `json:"sourceRetryCodes"`
	MaxSourceRate        int              `json:"maxSourceRate"`
	MaxSourceConcurrency int              `json:"maxSourceConcurrency"`
	SourceURLScheme      string           `json:"sourceURLScheme"`
}

//	更新缓存api结构
type UpdateCacheReq struct {
	CmdArgs        []string
	Uid            uint32         `json:"uid"`
	CacheControls  []CacheControl `json:"cacheControls"`
	IgnoreQueryStr *bool          `json:"ignoreParam"`
}

type UpdateDomainRefererReq struct {
	Uid          uint32      `json:"uid"`
	RefererType  RefererType `json:"refererType"`
	RefererValue string      `json:"refererValue"`
	NullReferer  *bool       `json:"nullReferer"`
}

type UpdateDomainTimeACLReq struct {
	Uid         uint32   `json:"uid"`
	TimeACL     bool     `json:"timeacl"`
	TimeACLKeys []string `json:"timeACLKeys"`
}

type GetDomains_Res struct {
	Code  int    `json:"code,omitempty"`
	Error string `json:"error,omitempty"`
	DomainInfo
}

func (r *GetDomains_Res) setCode(code int) {
	r.Code = code
	r.Error = ErrorText[code]
}

type GetDomainsReq struct {
	Marker            string         `json:"marker"`
	Limit             int            `json:"limit"`
	Uid               uint32         `json:"uid"`
	DomainPrefix      string         `json:"domainPrefix"`
	SourceType        SourceType     `json:"sourceType"`
	SourceQiniuBucket string         `json:"sourceQiniuBucket"`
	OperationType     Operation      `json:"operationType"`
	OperatingState    OperatingState `json:"operatingState"`
	CreateStart       time.Time      `json:"createStart"`
	CreateEnd         time.Time      `json:"createEnd"`
	RefererTypes      string         `json:"refererTypes"`
	NoQiniu           bool           `json:"noQiniu"`
	AbilityType       string         `json:"abilityType"`
}

type GetDomainsErrorsReq struct {
	Marker    string    `json:"marker"`
	Limit     int       `json:"limit"`
	Domain    string    `json:"domain"` // regex support
	ErrType   ErrType   `json:"errType"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type GetDomainsRes struct {
	Code        int          `json:"code"`
	Error       string       `json:"error"`
	Marker      string       `json:"marker"`
	DomainInfos []DomainInfo `json:"domainInfos"`
}

type GetDomainsLogsReq struct {
	Marker     string         `json:"marker"`
	Limit      int            `json:"limit"`
	Domain     string         `json:"domain"`
	Operation  Operation      `json:"operation"`
	StartTime  time.Time      `json:"startTime"`
	EndTime    time.Time      `json:"endTime"`
	GotDetails bool           `json:"gotDetails"`
	State      OperatingState `json:"state"`
	SortAsc    bool           `json:"sortAsc"`
}

type GetDomainsLogsRes struct {
	Marker     string          `json:"marker"`
	DomainLogs []DomainInfoLog `json:"domainLogs"`
}

type GetHttpsCrtsReq struct {
	Marker    string    `json:"marker"`
	Uid       uint32    `json:"uid"`
	Limit     int       `json:"limit"`
	Domain    string    `json:"domain"` // regex support
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type GetHttpsCrtsRes struct {
	Marker    string        `json:"marker"`
	HttpsCrts []HttpsCrtRes `json:"httpscrts"`
}

type HttpsCrtRes struct {
	Name     string    `json:"name" bson:"name"`
	Uid      uint32    `json:"uid" bson:"uid"`
	CreateAt time.Time `json:"createAt" bson:"createAt"`
	ModifyAt time.Time `json:"modifyAt" bson:"modifyAt"`
}

type GetHttpsCrtsFileReq struct {
	Type HttpsCrtType `json:"type"`
}

type GetDomainsErrRes struct {
	CommonRes
	Marker     string      `json:"marker"`
	DomainErrs []DomainErr `json:"domainErrs"`
}

func (r *GetDomainsRes) setCode(code int) {
	r.Code = code
	r.Error = ErrorText[code]
}

type CreateUserSettingReq struct {
	Uid               uint32 `json:"uid"`
	Priority          int    `json:"priority"`
	RefreshUrlStatus  string `json:"refreshUrlStatus"`
	RefreshDirStatus  string `json:"refreshDirStatus"`
	RefreshUrlLimit   int    `json:"refreshUrlLimit"`
	RefreshDirLimit   int    `json:"refreshDirLimit"`
	PrefetchUrlStatus string `json:"prefetchUrlStatus"`
	PrefetchUrlLimit  int    `json:"prefetchUrlLimit"`
	//CreateAt          time.Time `json:"createAt" bson:"createAt"`
}

type GetUserSettingRes struct {
	CommonRes
	Uid               uint32    `json:"uid"`
	Priority          int       `json:"priority"`
	RefreshUrlStatus  string    `json:"refreshUrlStatus"`
	RefreshDirStatus  string    `json:"refreshDirStatus"`
	RefreshUrlLimit   int       `json:"refreshUrlLimit"`
	RefreshDirLimit   int       `json:"refreshDirLimit"`
	PrefetchUrlStatus string    `json:"prefetchUrlStatus"`
	PrefetchUrlLimit  int       `json:"prefetchUrlLimit"`
	CreateAt          time.Time `json:"createAt" bson:"createAt"`
}

type SetUserDefaultReq struct {
	IoDomains         []string `json:"ioDomains"`
	RefreshUrlStatus  string   `json:"refreshUrlStatus"`
	RefreshDirStatus  string   `json:"refreshDirStatus"`
	PrefetchUrlStatus string   `json:"prefetchUrlStatus"`
	RefreshUrlLimit   int      `json:"refreshUrlLimit"`
	RefreshDirLimit   int      `json:"refreshDirLimit"`
	PrefetchUrlLimit  int      `json:"prefetchUrlLimit"`
	Priority          int      `json:"priority"`
}

type GetUserDefaultRes struct {
	CommonRes
	IoDomains         []string `json:"ioDomains"`
	RefreshUrlStatus  string   `json:"refreshUrlStatus"`
	RefreshDirStatus  string   `json:"refreshDirStatus"`
	PrefetchUrlStatus string   `json:"prefetchUrlStatus"`
	RefreshUrlLimit   int      `json:"refreshUrlLimit"`
	RefreshDirLimit   int      ` json:"refreshDirLimit"`
	PrefetchUrlLimit  int      `json:"prefetchUrlLimit"`
	Priority          int      `json:"priority"`
}

type SetUserStatisticsReq struct {
	Uid                     uint32 `bson:"uid" json:"uid"`
	CurrentRefreshUrlCount  int    `json:"currentRefreshUrlCount"`
	CurrentRefreshDirCount  int    `json:"currentRefreshDirCount"`
	CurrentPrefetchUrlCount int    `json:"currentPrefetchUrlCount"`
}

type GetUserStatisticsRes struct {
	CommonRes
	Uid                     uint32 `bson:"uid" json:"uid"`
	CurrentRefreshUrlCount  int    `json:"currentRefreshUrlCount"`
	CurrentRefreshDirCount  int    `json:"currentRefreshDirCount"`
	CurrentPrefetchUrlCount int    `json:"currentPrefetchUrlCount"`
	Date                    string `json:"date"`
}

type SetCdnDefaultReq struct {
	Cdn              string `json:"cdn"`
	RefreshUrlLimit  int    `json:"refreshUrlLimit"`
	RefreshDirLimit  int    `json:"refreshDirLimit"`
	PrefetchUrlLimit int    `json:"prefetchUrlLimit"`
}

type GetCdnDefaultRes struct {
	CommonRes
	Cdn              string `json:"cdn"`
	RefreshUrlLimit  int    `json:"refreshUrlLimit"`
	RefreshDirLimit  int    `json:"refreshDirLimit"`
	PrefetchUrlLimit int    `json:"prefetchUrlLimit"`
}

type SetCdnStatisticsReq struct {
	Cdn                     string `json:"cdn"`
	CurrentRefreshUrlCount  int    `json:"currentRefreshUrlCount"`
	CurrentRefreshDirCount  int    `json:"currentRefreshDirCount"`
	CurrentPrefetchUrlCount int    `json:"currentPrefetchUrlCount"`
}

type GetCdnStatisticsRes struct {
	CommonRes
	Cdn                     string `json:"cdn"`
	CurrentRefreshUrlCount  int    `json:"currentRefreshUrlCount"`
	CurrentRefreshDirCount  int    `json:"currentRefreshDirCount"`
	CurrentPrefetchUrlCount int    `json:"currentPrefetchUrlCount"`
	Date                    string `json:"date"`
}

type CdnInvokeSt struct {
	ApiName       string `json:"apiName"` //Refresh  Prefetch  PrefetchStatus
	Cdn           string `json:"cdn"`
	IntervalTime  int    `json:"intervalTime"` //mill second
	ExpireTime    int    `json:"expireTime"`   //second
	UrlNumPerTime int    `json:"urlNumPerTime"`
	DirNumPerTime int    `json:"dirNumPerTime"`
}

type SetCdnInvokeReq struct {
	CdnInvokeSt
}

type GetCdnInvokeRes struct {
	CommonRes
	CdnInvokes []CdnInvokeSt `json:"cdnInvokes"`
}

type PostRefreshReq struct {
	Urls   []string      `json:"urls"`
	Dirs   []string      `json:"dirs"`
	Caller string        `json:"caller"`
	Cdns   []CDNProvider `json:"cdns"`
}

type PostRefreshRes struct {
	CommonRes
	RequestId     string   `json:"requestId"`
	InvalidUrls   []string `json:"invalidUrls"`
	InvalidDirs   []string `json:"invalidDirs"`
	UrlQuotaDay   int      `json:"urlQuotaDay"`
	UrlSurplusDay int      `json:"urlSurplusDay"`
	DirQuotaDay   int      `json:"dirQuotaDay"`
	DirSurplusDay int      `json:"dirSurplusDay"`
}

type PostPrefetchReq struct {
	Urls        []string      `json:"urls"`
	Caller      string        `json:"caller"`
	Cdns        []CDNProvider `json:"cdns"`
	Prefetch    int           `json:"prefetch"`
	Fetchtype   int           `json:"fetchtype"`
	CallbackUrl string        `json:"callbackUrl"`
}

type PostPrefetchRes struct {
	CommonRes
	RequestId   string   `json:"requestId"`
	InvalidUrls []string `json:"invalidUrls"`
	QuotaDay    int      `json:"QuotaDay"`
	SurplusDay  int      `json:"SurplusDay"`
}

type SurplusQuotaDay struct {
	UrlQuotaDay   int
	UrlSurplusDay int
	DirQuotaDay   int
	DirSurplusDay int
	QuotaDay      int
	SurplusDay    int
}

type GetPrefetchQueryReq struct {
	Uid       uint32 `json:"uid"`
	RequestId string `json:"requestId"`
}

type UrlResult struct {
	Url    string `json:"url"`
	Result string `json:"result"`
}
type GetPrefetchQueryRes struct {
	CommonRes
	Uid       uint32      `json:"uid"`
	RequestId string      `json:"requestId"`
	Urls      []UrlResult `json:"urls"`
}

type PostPrefetchListReq struct {
	Marker    string   `json:"marker"`
	Limit     int      `json:"limit"`
	Uid       uint32   `json:"uid"`
	RequestId string   `json:"requestId"`
	Domain    string   `json:"domain"`
	State     string   `json:"state"`
	Urls      []string `json:"urls"`
	StartTime string   `json:"startTime"` //format 2015-11-01 12:05:05
	EndTime   string   `json:"endTime"`
	Caller    string   `json:"caller"`
}

type PrefetchUrlResult struct {
	Uid       uint32 `json:"uid"`
	RequestId string `json:"requestId"`
	State     string `json:"state"`
	Url       string `json:"url"`
	CreateAt  string `json:"createAt"`
}
type PostPrefetchListRes struct {
	CommonRes
	Marker string                     `json:"marker"`
	Items  []RefreshPrefetchUrlResult `json:"items"`
}

type PostRefreshListReq struct {
	Marker    string   `json:"marker"`
	Limit     int      `json:"limit"`
	Uid       uint32   `json:"uid"`
	RequestId string   `json:"requestId"`
	Domain    string   `json:"domain"`
	IsDir     string   `json:"isDir"`
	State     string   `json:"state"`
	Urls      []string `json:"urls"`
	StartTime string   `json:"startTime"` //format 2015-11-01 12:05:05
	EndTime   string   `json:"endTime"`
	Caller    string   `json:"caller"`
}

type RefreshPrefetchUrlResult struct {
	Uid         uint32 `json:"uid"`
	RequestId   string `json:"requestId"`
	Url         string `json:"url"`
	State       string `json:"state"`
	StateDetail string `json:"stateDetail"`
	CreateAt    string `json:"createAt"`
	BeginAt     string `json:"beginAt"`
	EndAt       string `json:"endAt"`
}

type RefreshUrlResult struct {
	RefreshPrefetchUrlResult
	MidState string `json:"midState"`
}
type PostRefreshListRes struct {
	CommonRes
	Marker string             `json:"marker"`
	Items  []RefreshUrlResult `json:"items"`
}

type RefreshDetail struct {
	Id          string `json:"id"`
	Uid         uint32 `json:"uid"`
	RequestId   string `json:"requestId"`
	Url         string `json:"url"`
	IsDir       string `json:"isDir"`
	Cdn         string `json:"cdn"`
	State       string `json:"state"`
	StateDetail string `json:"stateDetail"`
	AlarmId     string `json:"alarmId"`
	CreateAt    string `json:"createAt"`
	BeginAt     string `json:"beginAt"`
	EndAt       string `json:"endAt"`
	ExpireAt    string `json:"expireAt"`
}

type GetRefreshDetailQueryRes struct {
	CommonRes
	Items []RefreshDetail `json:"items"`
}

type PrefetchDetail struct {
	Id        string `json:"id"`
	Uid       uint32 `json:"uid"`
	RequestId string `json:"requestId"`
	Url       string `json:"url"`
	//IsDir       string `json:"isDir"`
	Cdn         string `json:"cdn"`
	State       string `json:"state"`
	StateDetail string `json:"stateDetail"`
	AlarmId     string `json:"alarmId"`
	CreateAt    string `json:"createAt"`
	BeginAt     string `json:"beginAt"`
	EndAt       string `json:"endAt"`
	ExpireAt    string `json:"expireAt"`
}

type GetPrefetchDetailQueryRes struct {
	CommonRes
	Items []PrefetchDetail `json:"items"`
}

// func (r *PostRefreshListRes) setCode(code int) {
// 	r.Code = code
// 	r.Error = ErrorText[code]
// }

type GetRandomUuidReq struct {
	Uid uint32 `json:"uid"`
}

type GetRandomUuidRes struct {
	CommonRes
	Id string `json:"id"`
}

type UpdateDomainNotesReq struct {
	Notes map[NoteType]string `json:"notes"`
	Uid   uint32              `json:"uid"`
}

type UpdateDomainAbilitiesReq struct {
	Abilities map[AbilityType]Ability `json:"abilities"`
	Uid       uint32                  `json:"uid"`
}

type GetDomainTrafficBandwidthReq struct {
	Uid       uint32 `json:"uid"`
	Domain    string `json:"domain"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type GetDomainBandwidthReq struct {
	Uid       uint32 `json:"uid"`
	Domain    string `json:"domain"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type TimeBandwidth struct {
	Time      int64 `json:"time"`
	Bandwidth int64 `json:"bandwidth"`
}

type GetDomainBandwidthRes struct {
	CommonRes
	Data []TimeBandwidth `json:"data"`
}

type TimeTraffic struct {
	Time int64 `json:"time"`
	Flow int64 `json:"flow"`
}

// type TimeFlow struct {
// 	Time int64 `json:"time"`
// 	Flow int64 `json:"flow"`
// }

type GetDomainTrafficRes struct {
	CommonRes
	Data []TimeTraffic `json:"data"`
}

type GetCdnTrafficStatReq struct {
	Domain    string   `json:"domain"`
	Protocol  Protocol `json:"protocol"`
	GeoCover  GeoCover `json:"geoCover"`
	StartTime string   `json:"startTime"`
	EndTime   string   `json:"endTime"`
}

type TimeValue struct {
	Time  []int64 `json:"time"`
	Value []int64 `json:"value"`
}

type GetCdnTrafficStatRes struct {
	CommonRes
	Data TimeValue `json:"data"`
	//Data []TimeTraffic `json:"data"`
}

type GetDomainsBandwidthsRes struct {
	CommonRes
	Data map[string]TimeValue `json:"data"`
}

type GetDomainsBandwidthsReq struct {
	Domains    string `json:"domains"`
	Uid        uint32 `json:"uid"`
	Start      int64  `json:"start"`
	End        int64  `json:"end"`
	Protocol   string `json:"protocol"`
	GeoCover   string `json:"geoCover"`
	PointCount int    `json:"pointCount"`
	Format     string `json:"format"`
}

type GetDomainsTrafficsRes struct {
	CommonRes
	Data map[string]TimeValue `json:"data"`
}

type GetDomainsTrafficsReq struct {
	Domains     string `json:"domains"`
	Uid         uint32 `json:"uid"`
	Start       int64  `json:"start"`
	End         int64  `json:"end"`
	Protocol    string `json:"protocol"`
	GeoCover    string `json:"geoCover"`
	Granularity int    `json:"granularity"`
	Format      string `json:"format"`
}

type GetDomainsBackTrafficsReq struct {
	CmdArgs     []string
	Uid         uint32 `json:"uid"`
	Start       int64  `json:"start"`
	End         int64  `json:"end"`
	Granularity int    `json:"granularity"`
}

type GetDomainsBackTrafficsRes struct {
	CommonRes
	Data TimeValue `json:"data"`
}

type GetDomainsBackBandwidthReq struct {
	CmdArgs    []string
	Uid        uint32 `json:"uid"`
	Start      int64  `json:"start"`
	End        int64  `json:"end"`
	PointCount int    `json:"point_count"`
}

type GetDomainsBackBandwidthRes struct {
	CommonRes
	Data TimeValue `json:"data"`
}

type PostDomainsSourceCheckReq struct {
	SourceType        SourceType       `json:"sourceType"`
	SourceIPs         []string         `json:"sourceIPs"`
	SourceDomain      string           `json:"sourceDomain"`
	SourceQiniuBucket string           `json:"sourceQiniuBucket"`
	AdvancedSources   []AdvancedSource `json:"advancedSources"`
	SourceHost        string           `json:"sourceHost"`
	TestURLPath       string           `json:"testURLPath"`
}

type GetDomainsBucketRerererReq struct {
	Uid    uint32 `json:"uid"`
	Bucket string `json:"bucket"`
}

type GetDomainsBucketRerererRes struct {
	RefererType   RefererType `json:"refererType"`
	RefererValues []string    `json:"refererValues"`
	NullReferer   bool        `json:"nullReferer"`
}

type GetDomainsBsrefererReq struct {
	CmdArgs []string
	Uid     uint32 `json:"uid"`
}

type GetDomainsBsrefererRes struct {
	RefererType   RefererType `json:"refererType"`
	RefererValues []string    `json:"refererValues"`
	NullReferer   bool        `json:"nullReferer"`
}

type PostPandomainReq struct {
	Bucket tblmgr.BucketEntry `json:"bucket"`
}

type PostUserPandomainReq struct {
	Uid        uint32 `json:"uid"`
	Bucket     string `json:"bucket"`
	PareDomain string `json:"pareDomain"`
}

type PostQnsslDomainReq struct {
	Uid    uint32 `json:"uid"`
	Bucket string `json:"bucket"`
}

type PostQnsslDomainRes struct {
	Domain string `json:"domain"`
}

type DelQnsslDomainReq struct {
	CmdArgs []string
	Uid     uint32 `json:"uid"`
}

type DelPanDomainReq struct {
	CmdArgs []string
	Uid     uint32 `json:"uid"`
}

type SyncCDNReq struct {
	CmdArgs     []string
	CDNProvider CDNProvider `json:"cdnProvider"`
}

type PostPubDomainReq struct {
	Uid    uint32 `json:"uid"`
	Domain string `json:"domain"`
	Bucket string `json:"bucket"`
}

type PostUnPubDomainReq struct {
	Uid    uint32 `json:"uid"`
	Domain string `json:"domain"`
}

type GetDomainsUidRes struct {
	Uid uint32 `json:"uid"`
}

type GetUserWildcarddomainReq struct {
	Uid    uint32 `json:"uid"`
	Marker string `json:"marker"`
	Limit  int    `json:"limit"`
}

type GetUserPandomainReq struct {
	Uid        uint32 `json:"uid"`
	PareDomain string `json:"pareDomain"`
	Marker     string `json:"marker"`
	Limit      int    `json:"limit"`
}

type PostDomainTaskReq struct {
	Domain      string         `json:"domain" bson:"domain"`
	CallbackUrl string         `json:"callbackURL" bson:"callbackURL"`
	Type        DomainTaskType `json:"type" bson:"type"`
	Uid         uint32         `json:"uid" bson:"uid"`
	RequestId   string         `json:"requestId" bson:"requestId"`
	Message     string         `json:"message" bson:"message"`
}

type IdRet struct {
	TaskId string `json:"taskId" bson:"taskId"`
}

type DomainTaskCallbackReq struct {
	TaskId     string         `json:"taskId"`
	Result     OperatingState `json:"result"`
	ResultDesc string         `json:"resultDesc"`
}

type TaskDomain struct {
	TaskId     string         `json:"taskId" bson:"taskId"`
	Domain     string         `json:"domain" bson:"domain"`
	Type       DomainTaskType `json:"type" bson:"type"`
	Uid        uint32         `json:"uid" bson:"uid"`
	Message    string         `json:"message" bson:"message"`
	Result     OperatingState `json:"result" bson:"result"`
	ResultDesc string         `json:"resultDesc" bson:"resultDesc"`
}

type ReqGetTaskDomain struct {
	TaskId string         `json:"taskId" bson:"taskId"`
	Domain string         `json:"domain" bson:"domain"`
	Type   DomainTaskType `json:"type" bson:"type"`
	Uid    uint32         `json:"uid" bson:"uid"`
	Result OperatingState `json:"result" bson:"result"`
	Marker string         `json:"marker"`
	Limit  int            `json:"limit"`
}

type GetTaskDomainsRes struct {
	Tasks  []TaskDomain `json:"tasks"`
	Marker string       `json:"marker"`
}

type ReqDomainsTransfer struct {
	CmdArgs []string
	Uid     uint32 `json:"Uid"`
	NewUid  uint32 `json:"NewUid"`
	Bucket  string `json:"Bucket"`
	Domain  string `json:"Domain"`
}

type CrtOverdue struct {
	CrtOverdueDays      int `json:"crt_overdue_days" bson:"crt_overdue_days"`
	AdminCrtOverdueDays int `json:"admin_crt_overdue_days" bson:"admin_crt_overdue_days"`
}

type FreezeUserDomainsArgs struct {
	Uid     uint32 `json:"Uid"`
	Message string `json:"Message"`
}
