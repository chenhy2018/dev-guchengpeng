package urlrewrite_test

import (
	"net/http"
	"testing"

	"github.com/qiniu/http/rewrite"
	"github.com/stretchr/testify/assert"
	"qbox.us/urlrewrite.v2"
)

// ---------------------------------------------------------------------------

func Test(t *testing.T) {

	//pan-domain-support and change the host when router_with_host is true
	req, err := http.NewRequest("GET", "http://a.img.detuyun.cn/dir1/dir2/path.jpg", nil)
	req.URL.Host = ""
	if err != nil {
		t.Fatal("NewRequest failed:", err)
	}
	hosts := []string{".img.detuyun.cn"}
	routers := []rewrite.RouteItem{{"(.*).img.detuyun.cn/(.*)", "xxx.qiniudn.com/${1}/${2}"}}
	routerWithHost := true

	table := &urlrewrite.Table{
		{
			Hosts:          hosts,
			Routers:        routers,
			RouterWithHost: routerWithHost,
		},
	}
	rules, err := urlrewrite.New(*table)
	if err != nil {
		t.Fatal("urlrewrite.New failed:", err)
	}

	_, err = urlrewrite.Do(rules, req)
	if err != nil {
		t.Fatal("urlrewrite.Do failed:", err)
	}
	assert.Equal(t, "http://xxx.qiniudn.com/a/dir1/dir2/path.jpg", req.URL.String(), "urlrewrite failed")
	assert.Equal(t, "xxx.qiniudn.com", req.Host, "urlrewrite host failed")

	//uri
	req, err = http.NewRequest("GET", "http://img.detuyun.cn/dir1/dir2/path.jpg", nil)
	req.URL.Host = ""
	if err != nil {
		t.Fatal("NewRequest failed:", err)
	}
	hosts = []string{"img.detuyun.cn"}
	routers = []rewrite.RouteItem{{"/(.*)", "/${1}"}}
	routerWithHost = false
	table = &urlrewrite.Table{
		{
			Hosts:          hosts,
			Routers:        routers,
			RouterWithHost: routerWithHost,
		},
	}
	rules, err = urlrewrite.New(*table)
	if err != nil {
		t.Fatal("urlrewrite.New failed:", err)
	}

	_, err = urlrewrite.Do(rules, req)
	if err != nil {
		t.Fatal("urlrewrite.Do failed:", err)
	}
	assert.Equal(t, "http:///dir1/dir2/path.jpg", req.URL.String(), "urlrewrite failed")
	assert.Equal(t, "img.detuyun.cn", req.Host, "urlrewrite host failed")

	//pan-domain support and uri
	req, err = http.NewRequest("GET", "http://a.img.detuyun.cn/dir1/dir2/path.jpg", nil)
	req.URL.Host = ""
	if err != nil {
		t.Fatal("NewRequest failed:", err)
	}
	hosts = []string{".img.detuyun.cn"}
	routers = []rewrite.RouteItem{{"/(.*)", "/${1}"}}
	routerWithHost = false
	table = &urlrewrite.Table{
		{
			Hosts:          hosts,
			Routers:        routers,
			RouterWithHost: routerWithHost,
		},
	}
	rules, err = urlrewrite.New(*table)
	if err != nil {
		t.Fatal("urlrewrite.New failed:", err)
	}

	_, err = urlrewrite.Do(rules, req)
	if err != nil {
		t.Fatal("urlrewrite.Do failed:", err)
	}
	assert.Equal(t, "http:///dir1/dir2/path.jpg", req.URL.String(), "urlrewrite failed")
	assert.Equal(t, "a.img.detuyun.cn", req.Host, "urlrewrite host failed")
}

// ---------------------------------------------------------------------------
