package main

import (
	_ "github.com/qiniu/version"
	_ "qbox.us/profile"
	"qiniu.com/apigate.v1/gateapp"
)

func main() {
	gateapp.Main(gateapp.Apigate_Default)
}
