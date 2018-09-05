package main

import (
	"fmt"
	"qbox.us/osl"
)

func main() {
	appdir, _ := osl.GetAppDataDir("QBox")
	fmt.Println(appdir)
}
