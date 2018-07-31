package model

import (
	"github.com/qiniu/ctype"
	"github.com/qiniu/http/rewrite"
)

type Global struct {
	// 全局默认缓存时间。
	GlobalCacheTime int `json:"globalCacheTime" bson:"globalCacheTime"`
	// 分发配置时允许的最大域名变化个数，用于防错。
	MaxDomainDiffNum int `json:"maxDomainDiffNum" bson:"maxDomainDiffNum"`
}

// 中间源节点。
type Node struct {
	// 名字。
	Name string `json:"name"`
	// 配置推送地址。
	Hosts []string `json:"hosts"`

	// 中间源可能和七牛源站部署在一个机房内网。
	// 如果发现回源域名对应到本机房的七牛源站，那么将域名替换成内网 IP。
	PriorSourceResolve Resolve `json:"prior_source_resolve"`
}

type Resolve struct {
	Suffixs []string `json:"suffixs"`
	IPs     []string `json:"ips"`
}

type Domain struct {
	// 用户域名。
	Name string `json:"name" bson:"name"`
	// 缓存策略。
	Cache `bson:",inline"`
	// 回源策略。
	Proxy `bson:",inline"`
}

type Cache struct {
	// 全局缓存时间，单位秒。值 0 表示使用七牛的默认全局缓存时间。
	GlobalCacheTime int `json:"globalCacheTime" bson:"globalCacheTime"`
}

type Proxy struct {
	// 回源点。
	Sources []Source `json:"sources" bson:"sources"`
	// 回源地址，为 Sources 中的 Addr。
	SourceAddrs []string `json:"sourceAddrs" bson:"sourceAddrs"`
	// 回源请求 Host。
	SourceHost string `json:"sourceHost" bson:"sourceHost"`
	// URL 重写规则。
	URLRewrites []Rewrite `json:"urlRewrites" bson:"urlRewrites,omitempty"`
	// 是否跟随跳转。
	// GET 请求跟随的跳转 Code 有 301，302，303，307。
	FollowRedirect bool `json:"followRedirect" bson:"followRedirect"`
	// 遇到 Range 请求是否在后台主动向 ATS 发起非 Range 请求从而让 ATS 缓存住对象。
	// 需要这么做的原因是 ATS 不能缓存 Range 请求，但已缓存对象的 Range 请求可以不用回源直接返回 Range 响应。
	DisableBackgroundFetch bool `json:"disableBackgroundFetch" bson:"disableBackgroundFetch"`
}

type Source struct {
	Addr   string `json:"addr" bson:"addr"`
	Weight int    `json:"weight" bson:"weight"`
	Backup bool   `json:"backup" bson:"backup"`
}

type Rewrite struct {
	// 匹配正则表达式。
	Pattern string `json:"pattern" bson:"pattern"`
	// 重写替换。
	Repl string `json:"repl" bson:"repl"`
}

func CompileRewrites(rs []Rewrite) (*rewrite.Router, error) {
	items := make([]rewrite.RouteItem, 0, len(rs))
	for _, r := range rs {
		item := rewrite.RouteItem{
			Pattern:     r.Pattern,
			Replacement: r.Repl,
		}
		items = append(items, item)
	}
	return rewrite.Compile(items)
}

// [a-zA-Z0-9] || [-.]
func IsDomainName(s string) bool {
	return ctype.IsType(ctype.DOMAIN_CHAR|ctype.COLON, s)
}
