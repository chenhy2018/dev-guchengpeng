// +build go1.7

package lb

import "net/http"

func isCancelReq(req *http.Request) bool {
	select {
	case <-req.Context().Done():
		return true
	case <-req.Cancel:
		return true
	default:
		return false
	}
	return false
}
