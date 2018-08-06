// +build go1.8

package pprof

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"
)

func Mutex(w http.ResponseWriter, r *http.Request) {
	rate, err := strconv.ParseInt(r.FormValue("rate"), 10, 64)
	if err != nil {
		rate = 1
	}
	debug, err := strconv.ParseInt(r.FormValue("debug"), 10, 64)
	if err != nil {
		debug = 1
	}
	sec, err := strconv.ParseFloat(r.FormValue("seconds"), 64)
	if sec <= 0 || err != nil {
		sec = 30
	}
	runtime.SetMutexProfileFraction(int(rate))
	defer runtime.SetMutexProfileFraction(0)
	sleep(w, time.Duration(sec)*time.Second)

	// Set Content Type assuming block.Start will work,
	// because if it does it starts writing.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if b := pprof.Lookup("mutex"); b != nil {
		err = b.WriteTo(w, int(debug))
		if err != nil {
			fmt.Fprintf(w, "Could not dump block: %v\n", err)
		}
	}
}
