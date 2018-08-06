// +build !go1.5

package pprof

import (
	"net/http"
)

// Trace responds with the execution trace in binary form.
// Tracing lasts for duration specified in seconds GET parameter, or for 1 second if not specified.
// The package initialization registers it as /debug/pprof/trace.
func Trace(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	w.Write([]byte("not support"))
}
