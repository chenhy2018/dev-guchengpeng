package apigate

import (
	"fmt"
	"net/http"
	"strings"
)

// --------------------------------------------------------------------

type AuthInfo struct {
	Utype uint64
	Su    bool
}

type AuthStuber interface {
	AuthStub(req *http.Request) (ai AuthInfo, ok bool, err error)
}

// --------------------------------------------------------------------

var g_auths = make(map[string]AuthStuber)
var g_auths_admin = make(map[string]AuthStuber)

func RegisterAuthStuber(name string, stuber, stuber2 AuthStuber) {

	g_auths[name] = stuber
	g_auths_admin[name] = stuber2
}

func GetAuthStuber(name string, allowFrozenAdmin bool) (stuber AuthStuber, ok bool) {

	if allowFrozenAdmin {
		stuber, ok = g_auths_admin[name]
	} else {
		stuber, ok = g_auths[name]
	}
	return
}

func GetAuthStubers(names []string, allowFrozenAdmin bool) (stubers []AuthStuber, err error) {

	if len(names) == 0 {
		return
	}

	var gauths = g_auths
	if allowFrozenAdmin {
		gauths = g_auths_admin
	}

	stubers = make([]AuthStuber, len(names))
	for i, name := range names {
		stuber, ok := gauths[name]
		if !ok {
			err = fmt.Errorf("auth name`%s` not found", name)
			return
		}
		stubers[i] = stuber
	}
	return
}

func AuthStub(req *http.Request, stubers []AuthStuber) (ai AuthInfo, ok bool, err error) {

	for _, stuber := range stubers {
		if ai, ok, err = stuber.AuthStub(req); ok {
			return
		}
	}
	return
}

// --------------------------------------------------------------------

const (
	Access_Public  = -1 // 公开（public/匿名可访问)
	Access_Invalid = 0  // 无效授权
)

type AccessInfo struct {
	Allow    int64
	NotAllow uint64
	SuOnly   bool
}

func (p AccessInfo) Can(ai AuthInfo) bool {

	if p.Allow <= 0 {
		panic("don't call AccessInfo.Can")
	}

	if p.SuOnly && !ai.Su {
		return false
	}

	if (uint64(p.Allow) & ai.Utype) == 0 {
		return false
	}

	return (p.NotAllow & ai.Utype) == 0
}

// --------------------------------------------------------------------

var g_access = map[string]int64{
	"public": Access_Public,
	"pass":   Access_Public,
	"":       Access_Invalid,
}

func RegisterAccessInfo(name string, ai int64) {

	g_access[name] = ai
}

func ParseAccessInfo(allow, notAllow string, suOnly bool) (ai AccessInfo, err error) {

	allow1, err := parseAi(allow)
	if err != nil {
		return
	}

	notAllow1, err := parseAi(notAllow)
	if err != nil {
		return
	}

	return AccessInfo{allow1, uint64(notAllow1), suOnly}, nil
}

func parseAi(access string) (ai int64, err error) {

	if access == "" {
		return 0, nil
	}

	parts := strings.Split(access, "|")
	for _, name1 := range parts {
		name := strings.Trim(name1, " \t")
		ai1, ok1 := g_access[name]
		if !ok1 {
			err = ErrInvalidAccess
			return
		}
		ai |= ai1
	}
	return
}

// --------------------------------------------------------------------
