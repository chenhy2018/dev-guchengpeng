// 相比于之前版本，仅仅只是Table的格式发生了变化
package urlrewrite

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"syscall"

	"github.com/qiniu/http/rewrite"
	"qbox.us/qconf/qconfapi"
)

// ---------------------------------------------------------------------------

type RuleFinder interface {
	RuleFind(host string) (router *rewrite.Router, err error)
}

// ---------------------------------------------------------------------------

type Table []struct {
	Hosts          []string            `json:"hosts"`
	Routers        []rewrite.RouteItem `json:"routers"`
	RouterWithHost bool                `json:"router_with_host"`
	NativeHandler  string              `json:"native_handler"` //如果存在此项则routers中正则表达式规则将不生效，编写参考native_handler.go中已有规则
}

type Rules map[string]*rewrite.Router // host -> Router

var qConfInit sync.Once

func New(args Table) (RuleFinder, error) {
	return NewWithQconf(args, nil)
}

func NewWithQconf(args Table, qConfCli *qconfapi.Client) (RuleFinder, error) {
	if qConfCli != nil {
		qConfInit.Do(func() { rewrite.QconfCli = qConfCli })
	}
	p := make(Rules)
	for _, item := range args {
		rule, err2 := rewrite.Compile3(item.Routers, item.RouterWithHost, item.NativeHandler)
		if err2 != nil {
			return nil, err2
		}
		for _, host := range item.Hosts {
			p[host] = rule
		}
	}
	return p, nil
}

func (p Rules) RuleFind(host string) (router *rewrite.Router, err error) {

	if r, ok := p[host]; ok {
		return r, nil
	}

	//pan domain support
	var dotIndex int
	dotIndex = strings.Index(host, ".")
	if dotIndex <= 0 {
		return nil, syscall.ENOENT
	}
	if r, ok := p[host[dotIndex:]]; ok {
		return r, nil
	}
	return nil, syscall.ENOENT
}

// ---------------------------------------------------------------------------

func Do(p RuleFinder, req *http.Request) (old *url.URL, err error) {

	router, err := p.RuleFind(req.Host)
	if err != nil {
		return
	}

	old = req.URL
	var new *url.URL
	//routers_with_host 开关为true，在pattern和repl中都会将 host 部分代入匹配和替换规则
	if router.WithHost() {
		newuri, err1 := router.Rewrite(req.Host + old.RequestURI())
		if err1 != nil {
			if err1 == rewrite.ErrUnmatched {
				err1 = syscall.ENOENT
			}
			return nil, err1
		}
		//change host and uri according to router
		if !strings.HasPrefix(newuri, "http://") {
			newuri = "http://" + newuri
		}
		new, err1 = url.Parse(newuri)
		if err1 != nil {
			return nil, err1
		}
		req.Host = new.Host
	} else {
		newuri, err2 := router.Rewrite(old.RequestURI())
		if err2 != nil {
			if err2 == rewrite.ErrUnmatched {
				err2 = syscall.ENOENT
			}
			return nil, err2
		}
		new, err2 = url.ParseRequestURI(newuri)
		if err2 != nil {
			return nil, err2
		}
		new.Host = old.Host
	}
	new.Scheme = old.Scheme
	req.URL = new
	req.RequestURI = req.URL.RequestURI()
	req.Header.Set("X-Origin-Uri", old.RequestURI())
	return
}

// ---------------------------------------------------------------------------
