package fusionline

import (
	"sort"
	"time"
)

type ListArgs struct {
	Type          string
	GeoCover      string
	Protocol      string
	Platform      string
	QiniuPrivate  string
	Hide          bool
	Default       string
	PlatformLevel string
	Features      string `json:"features"`
}

type Service interface {
	GetLines(lines *[]Line, args interface{}) error
}

type Cond map[string]interface{}

type GlobalLine struct {
	Id         string    `json:"id" bson:"_id"`
	Lines      []string  `json:"lines" bson:"lines"`
	CreateTime time.Time `json:"createTime" bson:"createTime"`
	UpdateTime time.Time `json:"updateTime" bson:"updateTime"`
}

type Line struct {
	Id         string    `json:"id" bson:"_id"`
	Name       string    `json:"name" bson:"name"`
	CreateTime time.Time `json:"createTime" bson:"createTime"`
	UpdateTime time.Time `json:"updateTime" bson:"updateTime"`
	Cname      string    `json:"cname" bson:"cname"`
	Comment    string    `json:"comment" bson:"comment"` // 线路的说明
	// 包含的baseLine及对应的覆盖，key为baseLine.id，value为 范围-运营商 组合
	BaseLines map[string][]string `json:"baseLines" bson:"baseLines"`
	// 默认baseLine
	DefaultBaseLine string `json:"defaultBaseLine" bson:"defaultBaseLine"`
	// 是否已经与dnspod同步
	DNSSynced     bool     `json:"dnsSynced" bson:"dnsSynced"`
	Hide          bool     `json:"hide" bson:"hide"`
	CDNProvider   string   `json:"cdnProvider" bson:"cdnProvider"`
	BaseProviders []string `json:"baseProviders" bson:"baseProviders"`
	// http | https
	Protocol string `json:"protocol" bson:"protocol"`
	// web | download | vod
	Platform string `json:"platform" bson:"platform"`
	// 网宿的平台等级 10 普通 | 20 次优
	PlatformLevel int `json:"platformLevel" bson:"platformLevel"`
	// 地理覆盖
	GeoCover     string `json:"geoCover" bson:"geoCover"`
	QiniuPrivate bool   `json:"qiniuPrivate" bson:"qiniuPrivate"`
	// 特性
	Features []string `json:"features" bson:"features"`
	// 默认线路
	Default bool `json:"default" bson:"default"`
	// 所包含的BaseLine struct（不入库，在从数据库Get时计算出来）
	BaseLineObjs map[string]Line `json:"baseLineObjs" bson:"-"`
	// 包含的 feature struct（不入库，在从数据库Get时计算出来）
	FeatureObjs map[string]Feature `json:"featureObjs" bson:"-"`
}

type Feature struct {
	Id          string `json:"id" bson:"_id"` // 不同于自带的_id
	Catagory    string `json:"catagory" bson:"catagory"`
	Description string `json:"description" bson:"description"`
}

type LineBackup struct {
	Id   string `json:"id" bson:"_id"`
	Line string `json:"line" bson:"line"` // 备份的Line.id
	// 其他字段同Line中同名字段
	BaseLines       map[string][]string `json:"baseLines" bson:"baseLines"`
	DefaultBaseLine string              `json:"defaultBaseLine" bson:"defaultBaseLine"`
	CreateTime      time.Time           `json:"createTime" bson:"createTime"`
	UpdateTime      time.Time           `json:"updateTime" bson:"updateTime"`
}

func (l *Line) validateInsert() (err error) {
	return
}

func (l *Line) validateUpdate() (err error) {
	return
}

func (l *Line) SortBaseLines() {
	for bl, _ := range l.BaseLines {
		sort.Sort(sort.StringSlice(l.BaseLines[bl]))
	}
}

func (l *Line) CopyBaseLines() map[string][]string {
	orgBaseLines := map[string][]string{}
	for baseLineId, coverage := range l.BaseLines {
		orgBaseLines[baseLineId] = coverage
	}
	return orgBaseLines
}
