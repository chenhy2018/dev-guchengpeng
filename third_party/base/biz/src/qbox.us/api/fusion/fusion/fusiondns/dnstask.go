package fusiondns

import (
	"code.google.com/p/go.net/context"
	"github.com/qiniu/xlog.v1"
	"qbox.us/api/fusion/fusion"
)

type DNSTask struct {
	Id          string   `json:"id" bson:"_id"`
	Domain      string   `json:"domain" bson:"domain"`
	DnsProvider string   `json:"dnsProvider" bson:"dnsProvider"`
	RecordType  string   `json:"recordType" bson:"recordType"`
	CallbackUrl string   `json:"callbackURL" bson:"callbackURL"`
	RecordValue string   `json:"recordValue" bson:"recordValue"`
	TaskCommand string   `json:"taskCommand" bson:"taskCommand"`
	SubDomain   string   `json:"subDomain" bson:"subDomain"`
	Position    Position `json:"position" bson:"position"`
	Status      string   `json:"status" bson:"status"`
	Retry       int      `json:"retry" bson:"retry"`
	Line        string   `bson:"line" json:"line"`
	Ttl         int64    `bson:"ttl" json:"ttl"`
	Mx          int64    `bson:"mx" json:"mx"`
	Enabled     string   `bson:"enabled" json:"enabled"`
	Remark      string   `bson:"remark" json:"remark"`
	Weight      int      `json:"weight" bson:"weight"`
	RequestId   string   `bson:"requestId"`
	UpdateTime  int64    `json:"updateTime" bson:"updateTime"`
}

const (
	ErrOK = 200
)

type Position struct {
	Country  string `json:"country" bson:"country"`
	Province string `json:"province" bson:"province"`
	City     string `json:"city" bson:"city"`
	ISP      string `json:"isp" bson:"isp"`
}

const (
	TaskStatusReady string = "ready"
	TaskStatusDoing string = "processing"
	TaskStatusDone  string = "success"
	TaskStatusFail  string = "failed"
)

const (
	TaskCommandRemoveRecord string = "dnspod_remove"
	TaskCommandQueryRecord  string = "dnspod_query"
	TaskCommandCreate       string = "dnspod_create"
	TaskCommandSet          string = "dnspod_set"
)

const (
	Default  string = "默认"
	Dianxin  string = "电信"
	Liantong string = "联通"
	Jiaoyu   string = "教育网"
	Yidong   string = "移动"
	Tietong  string = "铁通"
	China    string = "国内"
	Foreign  string = "国外"
)

func Position2Line(position Position, lines []string) (line string) {
	if len(lines) == 0 {
		return Default
	}

	if position.Province == "" && position.ISP == "" {
		return Default
	}

	linesParam := position.Province + position.ISP

	for _, line := range lines {
		if line == linesParam {
			return line
		}
	}

	return
}

func Position2LineEx(xl *xlog.Logger, position Position, lines []string) (line string) {
	xl.Info("position", position)
	if len(lines) == 0 {
		return Default
	}

	if position.Province == "" && position.ISP == "" {
		return Default
	}

	linesParam := position.Province + position.ISP

	xl.Info("linesParam", linesParam)

	xl.Info("lines", lines)

	for _, line := range lines {
		if line == linesParam {
			return line
		}
	}

	return
}

type TaskResult struct {
	Id string `json:"taskId"`
}

type SetRecordArgs struct {
	Domain      string             `json:"-"`
	SubDomain   string             `json:"-"`
	Position    Position           `json:"position"`
	DNSProvider fusion.DNSProvider `json:"cdnProvider"`
	RecordType  string             `json:"recordType"`
	RecordValue string             `json:"recordValue"`
	SyncOp      bool               `json:"syncOp"`
	CallbackURL string             `json:"callbackURL"`
	Weight      int                `json:"weight"`
}

type Service interface {
	SetRecord(ctx context.Context, args *SetRecordArgs) (task TaskResult, err error)
	QueryRecord(ctx context.Context, args *QueryRecordArgs) (ret *QueryRecordRes, err error)
	DeleteRecord(ctx context.Context, args *DeleteRecordArgs) (task TaskResult, err error)
	DnsGrey(ctx context.Context, args *DnsGreyArgs) (err error)
}

type QueryRecordArgs struct {
	Position    Position           `json:"position"`
	DNSProvider fusion.DNSProvider `json:"cdnProvider"`
	SubDomain   string             `json:"-"`
}

type DnsGreyArgs struct {
	CnameSource     string `json:"cnameSource"`
	OldCnameAddress string `json:"oldCnameAddress"`
	NewCnameAddress string `json:"newCnameAddress"`
	ToWeightLevel   int    `json:"toWeightLevel"`
}

type DeleteRecordArgs struct {
	Position    Position           `json:"position"`
	DNSProvider fusion.DNSProvider `json:"cdnProvider"`
	SubDomain   string             `json:"-"`
	SyncOp      bool               `json:"syncOp"`
}

type QueryRecordItem struct {
	Position    Position           `json:"position"`
	DNSProvider fusion.DNSProvider `json:"cdnProvider"`
	RecordType  string             `json:"recordType"`
	RecordValue string             `json:"recordValue"`
	SyncOp      bool               `json:"syncOp"`
	Weight      int                `json:"weight"`
}

type SubDomainArgs struct {
	Records   []PosRecord `json:"records"`
	SubDomain string      `json:"-"`
}

type PosRecord struct {
	Position    Position `json:"position"`
	RecordValue string   `json:"recordValue"`
}

type QueryRecordRes struct {
	Items []QueryRecordItem `json:"items"`
}

type GetDomainArgs struct {
	CmdArgs  []string
	Position Position `json:"position"`
}
