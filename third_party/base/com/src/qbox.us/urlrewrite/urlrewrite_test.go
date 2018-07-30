package urlrewrite_test

import (
	"fmt"
	"github.com/qiniu/http/rewrite"
	"net/http"
	"qbox.us/urlrewrite"
	"testing"
)

// ---------------------------------------------------------------------------

func Test(t *testing.T) {

	req, err := http.NewRequest("GET", "http://starusr.vancl.com.cn/148/218/star/suitsource//417/222/28417222/suit/11351017/bc2c1de70751434398f2e1b8c9cec797_source.jpg", nil)
	if err != nil {
		t.Fatal("NewRequest failed:", err)
	}
	fmt.Println(req.URL.RequestURI())

	rules, err := urlrewrite.New(urlrewrite.Table{
		"starusr.vancl.com.cn": []rewrite.RouteItem{
			{"/(\\d+)/(\\d+)/(.*)", "/${3}-${1}x${2}"},
		},
	})
	if err != nil {
		t.Fatal("urlrewrite.New failed:", err)
	}

	_, err = urlrewrite.Do(rules, req)
	if err != nil {
		t.Fatal("urlrewrite.Do failed:", err)
	}
	fmt.Println(req.URL.String())
}

// ---------------------------------------------------------------------------
