package profile

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	rpprof "runtime/pprof"
	"strconv"
	"sync"
	"time"

	"qbox.us/profile/expvar"
	"qbox.us/profile/pprof"
)

var (
	pushGateWay  = os.Getenv("service_metrics_push_gateway_addr")
	pushDuration = os.Getenv("service_metrics_push_duration_seconds")
)

var profileAddr string
var initDone = make(chan bool)

func GetProfileAddr() string {
	<-initDone
	return profileAddr
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "usage:\n\nvars:\n%s/debug/vars\n\nprofile:\n%s/debug/pprof/\n\npromethus:\n%s/metrics\n",
		GetProfileAddr(), GetProfileAddr(), GetProfileAddr(),
	)
}

func genDumpCommand(listenAddr string) {
	fn := os.Args[0] + "_profile_dump.sh"
	profileAddr = "http://" + listenAddr
	close(initDone)
	log.Println("profile listen at:", profileAddr)
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	prefix := fmt.Sprintf("%s_%s_%s_",
		filepath.Base(os.Args[0]),
		fmt.Sprint(os.Getpid()),
		time.Now().Format("20060102_150405"),
	)
	fmt.Fprintln(f, "#!/bin/sh")
	fmt.Fprintln(f, "echo `date` dumping runtime status")
	fmt.Fprintln(f, "set -x")
	fmt.Fprintf(f, "curl -sS '%s/debug/vars?seconds=5' -o %svars\n", profileAddr, prefix)
	fmt.Fprintf(f, "curl -sS '%s/metrics' -o %smetrics\n", profileAddr, prefix)
	for _, p := range rpprof.Profiles() {
		fmt.Fprintf(f, "curl -sS '%s/debug/pprof/%s?seconds=5' -o %s%s\n",
			profileAddr, p.Name(), prefix, p.Name())
		if p.Name() == "goroutine" {
			fmt.Fprintf(f, "curl -sS '%s/debug/pprof/%s?debug=1' -o %s%s_debug_1\n",
				profileAddr, p.Name(), prefix, p.Name())
			fmt.Fprintf(f, "curl -sS '%s/debug/pprof/%s?debug=2' -o %s%s_debug_2\n",
				profileAddr, p.Name(), prefix, p.Name())
		}
	}
	fmt.Fprintf(f, "curl -sS '%s/debug/pprof/profile?seconds=5' -o %sprofile\n", profileAddr, prefix)
	fmt.Fprintf(f, "curl -sS '%s/debug/pprof/trace?seconds=5' -o %strace\n", profileAddr, prefix)
}

func runPush() {
	if pushGateWay == "" {
		pushGateWay = "http://127.0.0.1:1056/"
	}
	pushDurationSecond, _ := strconv.Atoi(pushDuration)
	if pushDurationSecond == 0 {
		pushDurationSecond = 5
	}
	if pushDurationSecond < 0 {
		return
	}
	once := &sync.Once{}
	tic := time.NewTicker(time.Second * time.Duration(pushDurationSecond))
	job := fmt.Sprintf("%v:%v:%v", filepath.Base(os.Args[0]), host, os.Getpid())
	for {
		<-tic.C
		err := prometheusRegistry.PushAdd(job, "", pushGateWay)
		if err != nil {
			once.Do(func() { log.Println(err) })
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
	expvarInit()
	registerAll()
	go func() {
		var ln net.Listener
		var err error
		// 随机选一个30000以上端口，避免和内部服务的端口冲突
		randPort := 30000 + rand.Intn(20000)
		for i := randPort; i < randPort+5000; i += rand.Intn(10) {
			ln, err = net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(i))
			if err == nil {
				break
			}
			log.Println("pprof listen failed", err)
		}
		if err != nil {
			close(initDone)
			log.Println("pprof listen failed", err)
			return
		}
		mux := http.NewServeMux()
		mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		mux.Handle("/debug/pprof/block", http.HandlerFunc(pprof.Block))
		mux.Handle("/debug/pprof/mutex", http.HandlerFunc(pprof.Mutex))
		mux.HandleFunc("/debug/vars", expvar.ExpvarHandler)
		mux.HandleFunc("/debug/var/", getOneExpvar)
		mux.Handle("/metrics", prometheusRegistry)
		mux.HandleFunc("/", index)
		genDumpCommand(ln.Addr().String())
		go runPush()
		log.Println(http.Serve(ln, mux))
	}()
}
