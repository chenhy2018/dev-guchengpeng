package main

import (
	"fmt"
	"net/url"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: urlescape <Path>")
		return
	}

	uri := os.Args[1]
	uri2 := url.QueryEscape(uri)
	fmt.Println(uri2)
}
