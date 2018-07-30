package main

import (
	"fmt"
	"flag"
	"io/ioutil"
	"testing"

	"qiniupkg.com/x/log.v7"

	. "qiniupkg.com/httptest.v1/exec"
	_ "qiniupkg.com/qiniutest/httptest.v1/exec/plugin"
)

var (
	verbose = flag.Bool("v", false, "verbose")
)

// ---------------------------------------------------------------------------

func allMatch(pat, str string) (bool, error) {

	return true, nil
}

func allTests(t *testing.T) {

	if *verbose {
		log.SetOutputLevel(0)
	}

	if flag.NArg() < 1 {
		fmt.Println("Usage: qiniutest -v <QiniutestFile.qtf>")
		return
	}

	filePath := flag.Arg(0)
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal("ReadFile failed:", err)
		return
	}

	err = ExecCases(t, string(b))
	if err != nil {
		t.Fatal("ExecCases failed:", err)
	}
}

// Usage: qiniutest <QiniutestFile.qtf>
//
func main() {

	log.SetFlags(log.Llevel | log.LstdFlags)
	tests := []testing.InternalTest{
		{"main", allTests},
	}
	testing.Main(allMatch, tests, nil, nil)
}

// ---------------------------------------------------------------------------

