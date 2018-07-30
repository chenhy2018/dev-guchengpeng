package main

import (
	"fmt"
	"github.com/qiniu/iconv/encodings"
	"github.com/qiniu/log.v1"
	"io/ioutil"
	"os"
	"strings"
)

// ---------------------------------------------------

func usage() {

	fmt.Println(
		`Usage: qconvmv Encoding1,Encoding2,...,EncodingN Path1 Path2 ... PathN
`)

}

func convmv(cds *encodings.Iconvs, path string, outbuf []byte) {

	fi, err := os.Stat(path)
	if err != nil {
		log.Warn("Stat failed:", err)
		return
	}

	convmvImpl(cds, path, fi, outbuf)
}

func convmvImpl(cds *encodings.Iconvs, path string, fi os.FileInfo, outbuf []byte) {

	if fi.IsDir() {
		fis, err := ioutil.ReadDir(path)
		if err != nil {
			log.Warn("ReadDir failed:", err)
			return
		}
		path1 := path + "/"
		for _, sub := range fis {
			convmvImpl(cds, path1+sub.Name(), sub, outbuf)
		}
	}

	name := fi.Name()
	if out, ok := cds.Conv([]byte(name), outbuf, 0); ok {
		path2 := path[:len(path)-len(name)] + string(out)
		err := os.Rename(path, path2)
		if err != nil {
			log.Warn("Rename failed:", err)
			return
		}
		log.Info("Rename", path, "=>", path2)
	}
}

// ---------------------------------------------------

func main() {

	if len(os.Args) < 2 {
		usage()
		os.Exit(-1)
	}

	names := append([]string{"UTF8"}, strings.Split(os.Args[1], ",")...)
	cds, err := encodings.Open(names)
	if err != nil {
		fmt.Fprintln(os.Stderr, "encodings.Open failed:", err)
		os.Exit(-2)
	}
	defer cds.Close()

	outbuf := make([]byte, 4096)
	for i := 2; i < len(os.Args); i++ {
		convmv(cds, os.Args[i], outbuf)
	}
}

// ---------------------------------------------------
