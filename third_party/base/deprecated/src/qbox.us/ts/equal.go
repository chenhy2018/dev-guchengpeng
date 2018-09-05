package ts

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
)

// ---------------------------------------------------

func EqualStream(a io.Reader, b io.Reader) bool {
	a1, err := ioutil.ReadAll(a)
	if err != nil {
		return false
	}
	b1, err := ioutil.ReadAll(b)
	if err != nil {
		return false
	}
	return bytes.Equal(a1, b1)
}

func EqualFile(f1 string, f2 string) bool {
	file1, err := os.Open(f1)
	if err != nil {
		return false
	}
	defer file1.Close()
	file2, err := os.Open(f2)
	if err != nil {
		return false
	}
	defer file2.Close()
	return EqualStream(file1, file2)
}

// ---------------------------------------------------
