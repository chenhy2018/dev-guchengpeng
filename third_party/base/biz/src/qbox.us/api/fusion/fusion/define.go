package fusion

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/rpc.v2"
	"github.com/qiniu/xlog.v1"
	"qbox.us/api/cdn.v1/model"
	account "qbox.us/http/account.v2"
	"qbox.us/net/httputil"
)

const (
	DUMP_MAX_SIZE = 5000
	DEFAULT_LIMIT = 10
	MAX_LIMIT     = 1000
)

const (
	DESKEY = "11111111"
)

const (
	QINIU_BUCKET_TESTURL  = "qiniu_do_not_delete.gif"
	MID_SOURCE_DOMAINNAME = "nb-gate-io-msrc.qiniu.com"
)

const (
	LAYOUT = "2006-01-02 15:04:05"
)

type Granularity string

const (
	Granularity5Min  = "5min"
	GranularityHour  = "hour"
	GranularityDay   = "day"
	GranularityMonth = "month"
)

func ValidGranularity(s string) (Granularity, bool) {
	granularity := Granularity(s)
	switch granularity {
	case Granularity5Min, GranularityHour, GranularityDay, GranularityMonth:
		return granularity, true
	}
	return "", false
}

var (
	Local = time.FixedZone("CST", 8*3600)
)

type RefererType string

const (
	REFERER_TYPE_WHITE RefererType = "white"
	REFERER_TYPE_BLACK RefererType = "black"
)

func (r RefererType) IsValid() bool {
	return r == REFERER_TYPE_WHITE || r == REFERER_TYPE_BLACK
}

type HttpsCrtType string

const (
	HTTPS_CRT_TYPE_SVR_KEY HttpsCrtType = "serverKey"
	HTTPS_CRT_TYPE_SVR_CRT HttpsCrtType = "serverCrt"
	HTTPS_CRT_TYPE_CA_CRT  HttpsCrtType = "caCrt"
	HTTPS_CRT_TYPE_ALL     HttpsCrtType = "all"
)

func (h HttpsCrtType) IsValid() bool {
	switch h {
	case HTTPS_CRT_TYPE_SVR_KEY, HTTPS_CRT_TYPE_SVR_CRT, HTTPS_CRT_TYPE_CA_CRT,
		HTTPS_CRT_TYPE_ALL:
		return true
	}
	return false
}

var (
	HTTPS_CRT_TYPES = []HttpsCrtType{HTTPS_CRT_TYPE_SVR_KEY, HTTPS_CRT_TYPE_SVR_CRT, HTTPS_CRT_TYPE_CA_CRT}
)

func ReplyBaseErr(l rpc.Logger, env *rpcutil.Env, status, code int, res ISetCode, v ...interface{}) {
	xl := xlog.NewWith(l)
	xl.Error(v...)
	res.setCode(code)
	httputil.Reply(env.W, status, res)
	xl.Infof("res:%#v", res)
	return
}

func ReplyBaseOK(l rpc.Logger, env *rpcutil.Env, res ISetCode, v ...interface{}) {
	xl := xlog.NewWith(l)
	xl.Info(v...)
	res.setCode(ErrOK)
	httputil.Reply(env.W, 200, res)
	xl.Infof("res:%#v", res)
	return
}

func ReplyErr(l rpc.Logger, env *account.Env, status, code int, res ISetCode, v ...interface{}) {
	xl := xlog.NewWith(l)
	xl.Error(v...)
	res.setCode(code)
	httputil.Reply(env.W, status, res)
	xl.Infof("res:%#v", res)
	return
}

func ReplyOK(l rpc.Logger, env *account.Env, res ISetCode, v ...interface{}) {
	xl := xlog.NewWith(l)
	xl.Info(v...)
	res.setCode(ErrOK)
	httputil.Reply(env.W, 200, res)
	xl.Infof("res:%#v", res)
	return
}

func ReplyAdminErr(l rpc.Logger, env *account.AdminEnv, status, code int, res ISetCode, v ...interface{}) {
	xl := xlog.NewWith(l)
	xl.Error(v...)
	res.setCode(code)
	httputil.Reply(env.W, status, res)
	xl.Infof("res:%#v", res)
	return
}

func ReplyAdminOK(l rpc.Logger, env *account.AdminEnv, res ISetCode, v ...interface{}) {
	xl := xlog.NewWith(l)
	xl.Info(v...)
	res.setCode(ErrOK)
	httputil.Reply(env.W, 200, res)
	xl.Infof("res:%#v", res)
	return
}

func ReplyAdminBinaryOK(l rpc.Logger, env *account.AdminEnv, name string, body io.Reader, bytes int64, v ...interface{}) {
	xl := xlog.NewWith(l)
	xl.Info(v...)
	env.W.Header().Add("Content-Disposition", "attachment; filename=\""+name+"\"")
	httputil.ReplyBinary(env.W, body, bytes)
	xl.Infof("bytes:%d", bytes)
	return
}

func ReplyErrV2(l rpc.Logger, env *account.Env, status, code int, res ISetCode, v ...interface{}) {
	xl := xlog.NewWith(l)
	res.setCode(code)
	httputil.Reply(env.W, status, res)
	xl.Error(v...)
	return
}

func ReplyOKV2(l rpc.Logger, env *account.Env, res ISetCode, v ...interface{}) {
	xl := xlog.NewWith(l)
	res.setCode(ErrOK)
	httputil.Reply(env.W, 200, res)
	xl.Info(v...)
	return
}

func ReplyAdminErrV2(l rpc.Logger, env *account.AdminEnv, status, code int, res ISetCode, v ...interface{}) {
	xl := xlog.NewWith(l)
	res.setCode(code)
	httputil.Reply(env.W, status, res)
	xl.Error(v...)
	return
}

func ReplyAdminOKV2(l rpc.Logger, env *account.AdminEnv, res ISetCode, v ...interface{}) {
	xl := xlog.NewWith(l)
	res.setCode(ErrOK)
	httputil.Reply(env.W, 200, res)
	xl.Info(v...)
	return
}

type DomainInfo struct {
	Name                 string           `json:"name" bson:"name"`
	Type                 Type             `json:"type" bson:"type"`
	Cname                string           `json:"cname" bson:"cname"`
	Cnamed               bool             `json:"cnamed,omitempty" bson:"cnamed"`
	LineId               string           `json:"lineId" bson:"lineId"`
	LineName             string           `json:"lineName,omitempty" bson:"lineName"`
	Uid                  uint32           `json:"uid,omitempty" bson:"uid"`
	SourceCname          string           `json:"sourceCname" bson:"sourceCname"`
	SourceType           SourceType       `json:"sourceType" bson:"sourceType"`
	SourceHost           string           `json:"sourceHost" bson:"sourceHost"`
	SourceIPs            []string         `json:"sourceIPs" bson:"sourceIPs"`
	SourceDomain         string           `json:"sourceDomain" bson:"sourceDomain"`
	SourceQiniuBucket    string           `json:"sourceQiniuBucket" bson:"sourceQiniuBucket"`
	AdvancedSources      []AdvancedSource `json:"advancedSources" bson:"advancedSources"`
	TestURLPath          string           `json:"testURLPath" bson:"testURLPath"`
	URLRewrites          []model.Rewrite  `json:"urlRewrites"`
	AddRespHeader        http.Header      `json:"addRespHeader"`
	SourceRetryCodes     []int            `json:"sourceRetryCodes"`
	MaxSourceRate        int              `json:"maxSourceRate"`
	MaxSourceConcurrency int              `json:"maxSourceConcurrency"`
	SourceURLScheme      string           `json:"sourceURLScheme"`

	MidSource MidSource `json:"midSource,omitempty" bson:"midSource"`

	Platform     Platform `json:"platform" bson:"platform"`
	GeoCover     GeoCover `json:"geoCover" bson:"geoCover"`
	Protocol     Protocol `json:"protocol" bson:"protocol"`
	QiniuPrivate bool     `json:"qiniuPrivate" bson:"qiniuPrivate"`

	PlatformLevel      PlatformLevel  `json:"platformLevel,omitempty" bson:"platformLevel"`
	CreateAt           time.Time      `json:"createAt" bson:"createAt"`
	ModifyAt           time.Time      `json:"modifyAt" bson:"modifyAt"`
	OperationType      Operation      `json:"operationType" bson:"operationType"`
	OperatingState     OperatingState `json:"operatingState" bson:"operatingState"`
	OperatingStateDesc string         `json:"operatingStateDesc" bson:"operatingStateDesc"`

	CDNStates            []CDNState     `json:"cdnStates,omitempty" bson:"cdnStates"`
	SyncCDNStates        []CDNState     `json:"synccdnStates,omitempty" bson:"synccdnStates"`
	CnameDNSStates       []DNSState     `json:"cnameDNSState,omitempty" bson:"cnameDNSState"`
	SourceCnameDNSStates []DNSState     `json:"sourceCnameDNSState,omitempty" bson:"sourceCnameDNSState"`
	MidSourceState       MidSourceState `json:"midSourceState,omitempty" bson:"midSourceState"`
	WeightLevel          int            `json:"weightLevel,omitempty" bson:"weightLevel"`

	RefererType   RefererType `json:"refererType" bson:"refererType"`
	RefererValues []string    `json:"refererValues" bson:"refererValues"`

	IgnoreQueryStr bool           `json:"ignoreParam" bson:"ignoreParam"`
	CacheControls  []CacheControl `json:"cacheControls" bson:"cacheControls"`

	GotWildcard bool   `json:"gotWildcard" bson:"gotWildcard"`
	RegisterNo  string `json:"registerNo" bson:"registerNo"`

	IsOld bool `json:"isOld,omitempty" bson:"isOld"` // 内部标识是不是老的自定义域名

	Notes       map[NoteType]string `json:"notes,omitempty" bson:"notes"`
	NullReferer bool                `json:"nullReferer"`

	// 时间戳防盗链开关
	TimeACL bool `json:"timeACL" bson:"timeACL"`
	// 时间戳防盗链keys
	TimeACLKeys []string `json:"timeACLKeys" bson:"timeACLKeys"`
	MutiDomains bool     `json:"mutiDomains,omitempty"`
	DnsGrey     bool     `json:"dnsGrey,omitempty"`
	SwitchState string   `json:"switchState,omitempty"`

	SyncCDNState bool                    `json:"sync_cdn_state,omitempty" bson:"sync_cdn_state"`
	Abilities    map[AbilityType]Ability `json:"abilities,omitempty" bson:"abilities"`

	// 父域名(只对泛子域名有效)
	PareDomain string `json:"pareDomain" bson:"pareDomain"`

	// 是否可自助修改
	CouldOperateBySelf bool `json:"couldOperateBySelf,omitempty" bson:"couldOperateBySelf"`
}

type DIS []DomainInfo

func (d DIS) Len() int {
	return len(d)
}

func (d DIS) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d DIS) Less(i, j int) bool {
	return d[i].Name < d[j].Name
}

func (d DIS) LessItem(index int, item interface{}) bool {
	if value, ok := item.(string); ok {
		if d[index].Name < value {
			return true
		}
	}
	return false
}

func (d DIS) Equal(index int, item interface{}) bool {
	if value, ok := item.(string); ok {
		//if d[index].Name == value.Name {
		if d[index].Name == value {
			return true
		}
	}
	return false
}

func BinSearch(d DIS, item interface{}) int {
	startFlag := 0
	stopFlag := d.Len() - 1
	middleFlag := (startFlag + stopFlag) / 2
	//fmt.Println("begin:", startFlag, middleFlag, stopFlag)
	for (startFlag < stopFlag) && !d.Equal(middleFlag, item) {
		if d.LessItem(middleFlag, item) {
			startFlag = middleFlag + 1
		} else {
			stopFlag = middleFlag - 1
		}
		middleFlag = (startFlag + stopFlag) / 2
		//fmt.Println(startFlag, middleFlag, stopFlag)
	}

	//fmt.Println("end:", startFlag, middleFlag, stopFlag)
	if d.Equal(middleFlag, item) {
		return middleFlag
	} else {
		return -1
	}
}

type DISTIME []DomainInfo

func (d DISTIME) Len() int {
	return len(d)
}

func (d DISTIME) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d DISTIME) Less(i, j int) bool {
	return d[i].CreateAt.After(d[j].CreateAt)
}

func (d DISTIME) LessItem(index int, item interface{}) bool {
	if value, ok := item.(time.Time); ok {
		if d[index].CreateAt.After(value) {
			return true
		}
	}
	return false
}

func (d DISTIME) Equal(index int, item interface{}) bool {
	if value, ok := item.(time.Time); ok {
		//if d[index].Name == value.Name {
		if d[index].CreateAt.Equal(value) {
			return true
		}
	}
	return false
}

func BinSearchTime(d DISTIME, item interface{}) int {
	startFlag := 0
	stopFlag := d.Len() - 1
	middleFlag := (startFlag + stopFlag) / 2
	//fmt.Println("begin:", startFlag, middleFlag, stopFlag)
	for (startFlag < stopFlag) && !d.Equal(middleFlag, item) {
		if d.LessItem(middleFlag, item) {
			startFlag = middleFlag + 1
		} else {
			stopFlag = middleFlag - 1
		}
		middleFlag = (startFlag + stopFlag) / 2
		//fmt.Println(startFlag, middleFlag, stopFlag)
	}

	//fmt.Println("end:", startFlag, middleFlag, stopFlag)
	if d.Equal(middleFlag, item) {
		return middleFlag
	} else {
		return -1
	}
}

type DomainInfoLog struct {
	Id        string         `json:"id" bson:"_id"`
	Name      string         `json:"name" bson:"name"`
	CreateAt  time.Time      `json:"createAt" bson:"createAt"`
	Operation Operation      `json:"operation" bson:"operation"`
	Snapshot  DomainInfo     `json:"snapshot" bson:"snapshot"`
	After     DomainInfo     `json:"after" bson:"after"`
	State     OperatingState `json:"state" bson:"state"`
	Message   string         `json:"message" bson:"message"`
}

func DumpInfo(xl *xlog.Logger, prefix, info string, maxSize int) {
	if len(info) > maxSize {
		xl.Info(prefix, info[:maxSize])
	} else {
		xl.Info(prefix, info)
	}
}

func RmDuplicate(list []string) []string {
	var x []string = []string{}
	for _, i := range list {
		if len(x) == 0 {
			x = append(x, i)
		} else {
			for k, v := range x {
				if i == v {
					break
				}
				if k == len(x)-1 {
					x = append(x, i)
				}
			}
		}
	}
	return x
}

func ReverseAnsiString(ss string) (rs string) {
	s := []byte(ss)
	length := len(s)
	for i := 0; i < length/2; i++ {
		s[i], s[length-1-i] = s[length-1-i], s[i]
	}
	return string(s)
}

type DomainTaskType string

const (
	DOMAINTASK_TYPE_FREEZE   DomainTaskType = "freeze"
	DOMAINTASK_TYPE_UNFREEZE DomainTaskType = "unfreeze"
)

func (t DomainTaskType) Valid() bool {
	if ALL_DOMAINTASK_TYPES[t] {
		return true
	}
	return false
}

var (
	ALL_DOMAINTASK_TYPES = map[DomainTaskType]bool{
		DOMAINTASK_TYPE_FREEZE:   true,
		DOMAINTASK_TYPE_UNFREEZE: true,
	}
)

const (
	TaskStatusReady string = "ready"
	TaskStatusDoing string = "processing"
	TaskStatusDone  string = "success"
	TaskStatusFail  string = "failed"
)

type TaskError struct {
	Details map[string]error `json:"details" bson:"details"`
}

func (t *TaskError) Error() string {
	ts := make(map[string]string, len(t.Details))
	for k, v := range t.Details {
		ts[k] = v.Error()
	}
	bs, _ := json.Marshal(ts)
	return string(bs)
}
