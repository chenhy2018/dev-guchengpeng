package main

import (
	"flag"
	"conf"
	"stun"
)

var (
	help = flag.Bool("h", false, "print usage")
)

func init() {
	conf.Args.IP = flag.String("ip", "127.0.0.1", "udp server binding IP address")
	conf.Args.Port = flag.String("port", "3478", "specific port to bind")
	flag.Parse()
}

func main() {

	// print message
	if *help {
		flag.Usage()
		return
	}

	// start listening
	stun.ListenUDP(*conf.Args.IP, *conf.Args.Port)

	return
}

