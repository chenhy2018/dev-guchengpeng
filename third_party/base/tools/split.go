package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
)

var (
	unit uint64
	unitString = flag.String("u", "1G", "split file unit")
)

const (
	unitK = 1024
	unitM = 1024 * unitK
	unitG = 1024 * unitM
)

func ParseUnit(u string) (unit uint64, err error) {

	if u == "" {
		return 0, syscall.EINVAL
	}

	n := len(u)-1
	switch u[n] {
	case 'G':
		unit, u = unitG, u[:n]
	case 'M':
		unit, u = unitM, u[:n]
	case 'K':
		unit, u = unitK, u[:n]
	default:
		unit = 1
	}

	v, err := strconv.ParseUint(u, 10, 0)
	if err != nil {
		return
	}
	return v * unit, nil
}

func SplitFile(file string, unit int64) (err error) {

	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return
	}
	fsize := fi.Size()

	if fsize <= unit { // nothing to do
		return
	}

	idx := 0
	splitFilePrefix := file + "."
	for fsize > 0 {
		idx++
		unitFile := splitFilePrefix + strconv.Itoa(idx)
		unitLen := unit
		if fsize < unitLen {
			unitLen = fsize
		}
		err = copyUnit(unitFile, f, unitLen)
		if err != nil {
			return
		}
		fsize -= unitLen
	}
	return
}

func copyUnit(unitFile string, f *os.File, unitLen int64) (err error) {

	w, err := os.OpenFile(unitFile, os.O_RDWR | os.O_CREATE | os.O_EXCL, 0666)
	if err != nil {
		return
	}
	defer w.Close()

	_, err = io.CopyN(w, f, unitLen)
	return
}

// split -u 1G <File1> <File2> ...
//
func main() {

	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Usage: split [-u <Unit>] <File1> <File2> ...")
		flag.PrintDefaults()
		return
	}

	var err error
	unit, err = ParseUnit(*unitString)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ParseUnit failed:", err)
		os.Exit(1)
	}

	for _, file := range flag.Args() {
		err = SplitFile(file, int64(unit))
		if err != nil {
			fmt.Fprintf(os.Stderr, "SpiltFile `%s` failed: %v\n", file, err)
			os.Exit(2)
		}
	}
}
