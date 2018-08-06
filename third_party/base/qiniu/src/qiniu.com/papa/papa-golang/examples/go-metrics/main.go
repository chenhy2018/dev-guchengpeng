package main

import (
	"flag"

	"qiniu.com/papa/papa-golang/lib"
)

var addr = flag.String("listen-address", "127.0.0.1:8088", "The address to listen on")

func main() {
	flag.Parse()
	c := make(chan int)
	lib.ServeGoMetrics(*addr)
	<-c
}
