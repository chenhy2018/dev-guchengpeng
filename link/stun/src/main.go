package main

import (
	"flag"
	"conf"
	"stun"
	"sync"
)

var (
	help = flag.Bool("h", false, "print usage")
)

func init() {
	conf.Args.IP = flag.String("ip", "127.0.0.1", "udp server binding IP address")
	conf.Args.Port = flag.String("port", "3478", "specific port to bind")
	conf.Args.Realm = flag.String("realm", "link.org", "used for long-term cred for TURN")
	flag.Var(&conf.Args.Users, "u", "add one user to TURN server")

	flag.Parse()
}

func main() {

	// print message
	if *help {
		flag.Usage()
		return
	}

	wg := &sync.WaitGroup{}

	// start listening
	wg.Add(2)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			stun.ListenUDP(*conf.Args.IP, *conf.Args.Port)
		}
	}(wg)

	go func (wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			stun.ListenTCP(*conf.Args.IP, *conf.Args.Port)
		}
	}(wg)

	wg.Wait()

	return
}

