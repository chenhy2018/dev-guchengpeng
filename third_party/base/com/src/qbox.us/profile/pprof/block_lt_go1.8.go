// +build !go1.8

package pprof

import (
	"net/http"
)

func Block(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	w.Write([]byte("not support"))
}
