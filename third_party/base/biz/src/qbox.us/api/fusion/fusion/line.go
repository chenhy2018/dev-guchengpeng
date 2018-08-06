package fusion

type Line struct {
	// 线路 Id，线路唯一标示。
	Id string `json:"id" bson:"_id"`
	// 线路是否对客户公开
	Hide bool `json:"hide" bson:"hide"`
	// 线路 type，标识此线路是基本线路或融合线路
	Type LineType `json:"type" bson:"type"`
	// 线路名字。
	Name string `json:"name" bson:"name" binding:"required"`
	// 线路CNAME
	Cname string `json:"cname" bson:"cname"`
	// 融合线路的基本线路id列表
	BaseLines []string `json:"baseLines" bson:"baseLines"`
	// 线路支持的平台。
	Platform Platform `json:"platform" bson:"platform" binding:"required"`
	// 线路支持的地理覆盖。
	GeoCover GeoCover `json:"geoCover" bson:"geoCover" binding:"required"`
	// 线路平台等级
	PlatformLevel PlatformLevel `json:"platformLevel" bson:"platformLevel"`
	// 线路支持的协议。
	Protocol Protocol `json:"protocol" bson:"protocol" binding:"required"`
	// 线路是否支持缓存 Qiniu 的私有资源。
	QiniuPrivate bool `json:"qiniuPrivate" bson:"qiniuPrivate"`
	// 价格。
	Price Price `json:"price" bson:"price" binding:"required"`
	// 反馈时间。
	FeedbackHour int64 `json:"feedbackHour" bson:"feedbackHour" binding:"required"`
	// 线路的 CDN 供应商。
	CDNProvider CDNProvider `json:"cdnProvider" bson:"cdnProvider" binding:"required"`
	// 线路质量 Id。
	QualityId string `json:"qualityId" bson:"qualityId" binding:"required"`
	Default   bool   `json:"default" bson:"default"`
}

type Price struct {
	// 国外流量阶梯价格。
	ForeignLadderPerTB []LadderPrice `json:"foreignLadderPerTB" bson:"foreignLadderPerTB"`
	// 国内流量阶梯价格。
	ChinaLadderPerTB []LadderPrice `json:"chinaLadderPerTB" bson:"chinaLadderPerTB"`
}

type LadderPrice struct {
	Ladder int64   `json:"ladder" bson:"ladder" binding:"required"`
	Price  float64 `json:"price" bson:"price" binding:"required"`
}

type CDNProvider string

const (
	CDNProviderWangsu     CDNProvider = "wangsu"
	CDNProviderTencent                = "tencent"
	CDNProviderLetv                   = "letv"
	CDNProviderQiniu                  = "qiniu"
	CDNProviderKuaiwang               = "kuaiwang"
	CDNProviderDilian                 = "dilian"
	CDNProviderTongxin                = "tongxin"
	CDNProviderBaishanyun             = "baishanyun"
)

var CDNProvidersStr = map[CDNProvider]string{
	CDNProviderWangsu:     string("wangsu"),
	CDNProviderTencent:    string("tencent"),
	CDNProviderLetv:       string("letv"),
	CDNProviderQiniu:      string("qiniu"),
	CDNProviderKuaiwang:   string("kuaiwang"),
	CDNProviderDilian:     string("dilian"),
	CDNProviderTongxin:    string("tongxin"),
	CDNProviderBaishanyun: string("baishanyun"),
}

var CDNProviders = map[CDNProvider]bool{
	CDNProviderWangsu:     true,
	CDNProviderTencent:    true,
	CDNProviderLetv:       true,
	CDNProviderQiniu:      false,
	CDNProviderKuaiwang:   true,
	CDNProviderDilian:     true,
	CDNProviderTongxin:    true,
	CDNProviderBaishanyun: true,
}

type LineType string

const (
	BaseLineType   LineType = "base"
	FusionLineType          = "fusion"
)
