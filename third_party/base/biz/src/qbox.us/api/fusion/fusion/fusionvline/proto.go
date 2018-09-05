package fusionvline

import (
	"time"

	"golang.org/x/net/context"
	"qbox.us/api/fusion/fusion"
)

type Task struct {
	Id                string            `json:"taskId" bson:"_id"`
	Cmd               string            `json:"cmd" bson:"cmd"`
	Uid               uint32            `json:"Uid" bson:"uid"`
	CDNStates         []*CDNState       `json:"cdnStates" bson:"cdnStates"`
	Domain            string            `json:"domain" bson:"domain"`
	LineId            string            `json:"lineId" bson:"lineId"`
	Cname             string            `json:"cname" bson:"cname"`
	SourceCname       string            `json:"sourceCname" bson:"sourceCname"`
	SourceType        fusion.SourceType `json:"sourceType" bson:"sourceType"`
	SourceIPs         string            `json:"sourceIPs" bson:"sourceIPs"`
	SourceDomain      string            `json:"sourceDomain" bson:"sourceDomain"`
	SourceQiniuBucket string            `json:"sourceQiniuBucket"`
	CallbackURL       string            `json:"callbackURL" bson:"callbackURL"`
	TestURL           string            `json:"testURL" bson:"testURL"`
	State             string            `json:"state"`
	DnsOpReady        bool              `json:"dnsOpReady" bson:"dnsOpReady"`
	DnsFlowPercent    int               `json:"dnsFlowPercent" bson:"dnsFlowPercent"`
	OldCnameAddress   string            `json:"oldCnameAddress" bson:"oldCnameAddress"`
	Tag               string
}

type CnameCdn struct {
	Cname string `json:"cname"`
	Cdn   string `json:"cdn"`
}

type CDNState struct {
	CDNProvider fusion.CDNProvider    `json:"cdnProvider" bson:"cdnProvider"`
	TaskId      string                `json:"taskId" bson:"taskId"`
	LineId      string                `json:"lineId" bson:"lineId"`
	Cname       string                `json:"cname" bson:"cname"`
	State       fusion.OperatingState `json:"state" bson:"state"`
	StateDesc   string                `json:"stateDesc" bson:"stateDesc"`
}

type TaskResult struct {
	Id string `json:"taskId"`
}

type CreateArgs struct {
	Domain            string            `json:"-"`
	SourceType        fusion.SourceType `json:"sourceType"`
	SourceDomain      string            `json:"sourceDomain"`
	SourceIPs         string            `json:"sourceIPs"`
	SourceQiniuBucket string            `json:"sourceQiniuBucket"`
	LineId            string            `json:"lineId"`
	CallbackURL       string            `json:"callbackURL"`
	Uid               string            `json:"uid"`
	Cname             string            `json:"cname"`
}

type SyncToDomainDBArgs struct {
	Domain            string                `json:"-"`
	Uid               uint32                `json:"uid"`
	Cname             string                `json:"cname"`
	SourceCname       string                `json:"sourceCname"`
	SourceType        string                `json:"sourceType"`
	SourceIPs         string                `json:"sourceIPs"`
	SourceDomain      string                `json:"sourceDomain"`
	SourceQiniuBucket string                `json:"sourceQiniuBucket"`
	TestURLPath       string                `json:"testURLPath"`
	LineId            string                `json:"lineId"`
	LineCname         string                `json:"lineCname"`
	WeightLevel       int                   `json:"weightLevel" bson:"weightLevel"`
	CreateAt          time.Time             `json:"createAt" bson:"createAt"`
	RefererType       string                `json:"refererType" bson:"refererType"`
	RefererValues     []string              `json:"refererValues" bson:"refererValues"`
	IgnoreQueryStr    bool                  ` json:"ignoreParam" `
	CacheControls     []fusion.CacheControl `json:"cacheControls" bson:"cacheControls"`
}

type DomainExt struct {
	Domain   string `json:"domain"`
	Cname    string `json:"cname"`
	UseHTTPS bool   `json:"useHTTPS"`
	CDNState string `json:"cdnState"`
	Task     *Task  `json:"task"`
	Bucket   string `json:"bucket"`
}

type Service interface {
	Create(ctx context.Context, args *CreateArgs) (taskResult TaskResult, err error)
	GetTask(ctx context.Context, taskId string) (task Task, err error)
	DeleteTask(ctx context.Context, taskId string) (err error)
	//GetAllTask(ctx context.Context) (tasks []*Task, err error)
	SyncToDomainDB(ctx context.Context, args *SyncToDomainDBArgs) (err error)
	Buckets(ctx context.Context, uid uint32) (buckets []string, err error)
	Domains(ctx context.Context, uid uint32, bucket string) (ds []*DomainExt, err error)
}
