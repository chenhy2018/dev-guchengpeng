// +build !go1.7

package pprof

import "net/http"

func durationExceedsWriteTimeout(r *http.Request, seconds float64) bool {
	return false
}
