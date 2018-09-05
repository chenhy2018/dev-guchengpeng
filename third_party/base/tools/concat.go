package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
)

func ConcatFile(file string) (idx int, err error) {

	w, err := os.OpenFile(file, os.O_RDWR | os.O_CREATE | os.O_EXCL, 0666)
	if err != nil {
		return
	}
	defer w.Close()

	unitFilePrefix := file + "."
	for {
		unitFile := unitFilePrefix + strconv.Itoa(idx+1)
		err = concatUnit(w, unitFile)
		if err != nil {
			if err == syscall.ENOENT {
				err = nil
				return
			}
			return
		}
		idx++
	}
}

func concatUnit(w *os.File, unitFile string) (err error) {

	f, err := os.Open(unitFile)
	if err != nil {
		if e, ok := err.(*os.PathError); ok {
			return e.Err
		}
		return
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return
}

// Usage: concat <File>
//
func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: concat <File>")
		return
	}

	file := os.Args[1]
	n, err := ConcatFile(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Concat failed:", err)
		os.Exit(1)
	}

	if n == 0 {
		os.Remove(file)
		fmt.Fprintf(os.Stderr, "Concat `%s` failed: no part files found\n", file)
		os.Exit(2)
	}

	fmt.Printf("Concat `%s` ok, total %d files\n", file, n)
}
