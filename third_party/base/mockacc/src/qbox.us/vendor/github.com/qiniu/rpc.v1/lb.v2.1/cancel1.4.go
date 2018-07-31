// +build go1.4
// +build !go1.5

package lb

import (
	"net/http"
	"github.com/qiniu/xlog.v1"
)

func isCancelled(xl *xlog.Logger, req *Request) bool {
	return false
}

func isCancelReq(req *http.Request) bool {
	return false
}
