package fusionline

import (
	"time"

	"golang.org/x/net/context"
	"qbox.us/api/fusion/fusion"
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
	Create(ctx context.Context, l *fusion.Line) (err error)
	Get(ctx context.Context, id string) (l *fusion.Line, err error)
	Update(ctx context.Context, l *fusion.Line) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, listArgs ListArgs) (ls []*fusion.Line, err error)
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

}
