package main

import (
	"flag"
	"net"
	"log"
	"stun"
	"conf"
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
	listenUDP()

	return
}

func listenUDP() {

	udp, err := net.ResolveUDPAddr("udp", *conf.Args.IP + ":" + *conf.Args.Port)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	conn, err := net.ListenUDP("udp", udp)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		nr, rm, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error:", err)
			return
		}

		go func(req []byte, r *net.UDPAddr) {

			msg := &stun.Message{}
			var err error

			// dbg.PrintMem(req, 8)
			
			msg, err = stun.NewMessage(req)
			if err != nil {
				log.Println("Error: drop packet:", err)
				return
			}

			msg.Print("request") // request

			msg, err = msg.ProcessUDP(r)
			if err != nil {
				log.Println("Error: proc failure:", err)
				return
			}

			if msg == nil {
				return // no response
			}

			msg.Print("response") // response

			resp := msg.Buffer()
			_, err = conn.WriteToUDP(resp, r)
			if err != nil {
				log.Println("Error: write failure:", err)
			}
		}(buf[:nr], rm)
	}
}
