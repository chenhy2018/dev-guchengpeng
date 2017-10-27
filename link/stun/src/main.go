package main

import (
	"flag"
	"net"
	"log"
	"stun"
)

var (
	listenAddr = flag.String("addr", ":3478", "udp server binding address")
	help = flag.Bool("h", false, "print usage")
)

func init() {

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

	udp, err := net.ResolveUDPAddr("udp", *listenAddr)
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	conn, err := net.ListenUDP("udp", udp)
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		nr, rm, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error: ", err)
			return
		}

		go func(req []byte, r *net.UDPAddr) {

			msg := &stun.Message{}
			var err error

			// dbg.PrintMem(req, 8)
			
			msg, err = stun.NewMessage(req)
			if err != nil {
				log.Println("Error: drop packet: %s", err)
				return
			}
				
			msg.Print("request") // request
			
			msg, err = msg.ProcessUDP(r)
			if err != nil {
				log.Println("Error: proc failure: %s", err)
				return
			}

			msg.Print("response") // response

			resp := msg.Buffer()
			_, err = conn.WriteToUDP(resp, r)
			if err != nil {
				log.Println("Error: write failure: %s", err)
			}
		}(buf[:nr], rm)
	}
}
