package urlrewrite

import (
	"github.com/qiniu/http/rewrite"
	"net/http"
	"net/url"
	"syscall"
)

// ---------------------------------------------------------------------------

type RuleFinder interface {
	RuleFind(host string) (router *rewrite.Router, err error)
}

// ---------------------------------------------------------------------------

type Table map[string][]rewrite.RouteItem
type Rules map[string]*rewrite.Router

func New(args Table) (RuleFinder, error) {

	p := make(Rules, len(args))
	for host, items := range args {
		rule, err2 := rewrite.Compile(items)
		if err2 != nil {
			return nil, err2
		}
		p[host] = rule
	}
	return p, nil
}

func (p Rules) RuleFind(host string) (router *rewrite.Router, err error) {

	if r, ok := p[host]; ok {
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
	newuri, err := router.Rewrite(old.RequestURI())
	if err != nil {
		if err == rewrite.ErrUnmatched {
			err = syscall.ENOENT
		}
		return
	}

	new, err := url.ParseRequestURI(newuri)
	if err != nil {
		return
	}
	new.Scheme, new.Host = old.Scheme, old.Host
	req.URL = new
	req.Header.Set("X-Origin-Uri", old.RequestURI())
	return
}

// ---------------------------------------------------------------------------
