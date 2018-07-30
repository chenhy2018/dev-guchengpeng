package main

import (
	"fmt"
	"net"
)

func main() {
	itfs, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, itf := range itfs {
		fmt.Println(itf)
	}
}
