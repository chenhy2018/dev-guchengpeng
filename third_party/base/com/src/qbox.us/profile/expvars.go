package profile

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"qbox.us/profile/expvar"
)

var startTime = time.Now()
var binaryMd5 string

func expvarInit() {
	expvar.NewString("GoVersion").Set(runtime.Version())
	expvar.NewInt("NumCPU").Set(int64(runtime.NumCPU()))
	expvar.NewInt("Pid").Set(int64(os.Getpid()))
	expvar.NewInt("StartTime").Set(startTime.Unix())

	expvar.Publish("Uptime", expvar.Func(func() interface{} { return time.Since(startTime) / time.Second }))
	expvar.Publish("NumGoroutine", expvar.Func(func() interface{} { return runtime.NumGoroutine() }))

	data, err := ioutil.ReadFile(os.Args[0])
	if err != nil {
		log.Println(err)
		return
	}
	binaryMd5 = fmt.Sprintf("%x", md5.Sum(data))
	expvar.NewString("BinaryMd5").Set(binaryMd5)
	constLabels["md5"] = binaryMd5
}

// GET /debug/var/<var>
func getOneExpvar(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/debug/var/"):]
	val := expvar.Get(key)
	if val == nil {
		w.WriteHeader(612)
		return
	}
	w.Write([]byte(val.String()))
}
