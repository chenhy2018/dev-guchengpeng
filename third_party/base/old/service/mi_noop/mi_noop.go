package main

import (
	"fmt"
	"http"
	"qbox.us/api"
	"qbox.us/rpc"
)

func connect(w1 http.ResponseWriter, req *http.Request) {
	w := rpc.ResponseWriter{w1}
	w.ReplyWithCode(api.VersionTooOld)
}

func main() {

	fmt.Println("Start mi noop server ...")
	http.HandleFunc("/connect", connect)
	http.ListenAndServe(":9876", nil)
}
