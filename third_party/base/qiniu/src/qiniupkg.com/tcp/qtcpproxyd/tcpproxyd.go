package main

import (
	"flag"
	"fmt"
	"os"

	"qiniupkg.com/tcp/tcpproxyd.v1"
	"qiniupkg.com/x/log.v7"
)

var (
	host = flag.String("h", "", "qtcpproxyd's listen address")
)

func main() {

	flag.Parse()
	if *host == "" {
		fmt.Fprintf(os.Stderr, "Usage: qtcpproxy -h <ListenHost>\n\n")
		os.Exit(1)
	}

	log.Fatal(tcpproxyd.ListenAndServe(*host, nil))
}
