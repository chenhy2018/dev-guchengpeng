package rewrite

import (
	"bytes"
	"regexp"

	"github.com/qiniu/errors"
	"qbox.us/qconf/qconfapi"
)

var ErrUnmatched = errors.New("unmatched")
var ErrNamedSubexpUnsupported = errors.New("named subexp is unsupported")
var QconfCli *qconfapi.Client = nil

// ---------------------------------------------------------------------------

type RouteItem struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"repl"`
}

// ---------------------------------------------------------------------------

type Router struct {
	re             *regexp.Regexp
	items          []RouteItem
	base           []int
	routerWithHost bool
	nativeHandler  func(string) string
}

func (p *Router) WithHost() (routerWithHost bool) {
	return p.routerWithHost
}

func Compile(items []RouteItem) (p *Router, err error) {

	return Compile2(items, false)
}

func Compile3(items []RouteItem, routerWithHost bool, nativeHandlerS string) (p *Router, err error) {
	if nativeHandlerS != "" {
		if handler, ok := NativeHandler[nativeHandlerS]; ok {
			if handler.RouterWithHost != routerWithHost {
				return nil, errors.New(ErrRouterWithHost.Error() + ":" + nativeHandlerS)
			}
			return &Router{nativeHandler: handler.Handle, routerWithHost: routerWithHost}, nil
		}
	}
	return Compile2(items, routerWithHost)
}

func Compile2(items []RouteItem, routerWithHost bool) (p *Router, err error) {

	exp := bytes.NewBuffer(nil)

	base := make([]int, len(items)+1)
	base[0] = 1
	for i, item := range items {
		re2, err2 := regexp.Compile(item.Pattern)
		if err2 != nil {
			err = errors.Info(err2, "rewrite.Compile failed:", item.Pattern, item.Replacement).Detail(err2)
			return
		}
		subexpNames := re2.SubexpNames()
		for j := 1; j < len(subexpNames); j++ {
			if subexpNames[j] != "" {
				return nil, ErrNamedSubexpUnsupported
			}
		}
		base[i+1] = base[i] + len(subexpNames)
		if i > 0 {
			exp.WriteByte('|')
		}
		exp.WriteString("(^")
		exp.WriteString(item.Pattern)
		exp.WriteString("$)")
	}

	exp1 := string(exp.Bytes())
	re, err := regexp.Compile(exp1)
	if err != nil {
		err = errors.Info(err, "regexp.Compile failed:", exp1, err)
		return
	}
	return &Router{re, items, base, routerWithHost, nil}, nil
}

func MustCompile(items []RouteItem) (p *Router) {

	p, err := Compile(items)
	if err != nil {
		panic(err)
	}
	return
}

func (p *Router) Rewrite(src string) (dest string, err error) {
	if p.nativeHandler != nil {
		return p.nativeHandler(src), nil
	}
	match := p.re.FindStringSubmatchIndex(src)
	if match == nil {
		return src, ErrUnmatched
	}
	for i, item := range p.items {
		idx := p.base[i] << 1
		if match[idx] != match[idx+1] { // matched
			idx2 := p.base[i+1] << 1
			b := make([]byte, 0, (len(item.Replacement)+7)&^7)
			return string(p.re.ExpandString(b, item.Replacement, src, match[idx:idx2])), nil
		}
	}
	return src, ErrUnmatched
}

// ---------------------------------------------------------------------------
