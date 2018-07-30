package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

//
// mkver XXX.ver
//
func main() {

	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: mkver XXX.ver")
		return
	}

	verfile := os.Args[1]
	if !strings.HasSuffix(verfile, ".ver") {
		fmt.Fprintln(os.Stderr, "Invalid filename, must use .ver as extension")
		return
	}

	vdata, err := ioutil.ReadFile(verfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Open verfile failed:", err)
		return
	}

	verdir := verfile[:len(verfile)-4] + "/ver"
	err = os.MkdirAll(verdir, 0777)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Mkdir failed:", err)
		return
	}

	vtime := time.Now().Local().Format("20060102")

	gofile := verdir + "/verno.go"
	vertext := fmt.Sprintf("package ver\n\nvar String = \"%s.%s\"\n\n", strings.TrimSpace(string(vdata)), vtime)
	err = ioutil.WriteFile(gofile, []byte(vertext), 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Write failed:", err)
	}
}
