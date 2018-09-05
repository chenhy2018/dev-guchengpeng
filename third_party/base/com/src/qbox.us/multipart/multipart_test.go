package multipart

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"github.com/qiniu/ts"
	"testing"
)

func TestWrite(t *testing.T) {

	file := "multipartTest.tmp"
	ioutil.WriteFile(file, []byte("file content"), 0777)
	defer os.Remove(file)

	w := multipart.NewWriter(os.Stdout)
	defer w.Close()
	err := Write(w, map[string][]string{
		"a": {"1"},
		"b": {"hello"},
		"c": {"@" + file},
	})
	if err != nil {
		ts.Fatal(t, "Write failed:", err)
	}
}

func TestOpen(t *testing.T) {

	file := "multipartTest.tmp"
	ioutil.WriteFile(file, []byte("file content"), 0777)
	defer os.Remove(file)

	r, ct, err := Open(map[string][]string{
		"a": {"1"},
		"b": {"hello"},
		"c": {"@" + file},
	})
	if err != nil {
		ts.Fatal(t, "Open failed:", err)
	}
	defer r.Close()

	fmt.Println("\nContent-Type:", ct)
	io.Copy(os.Stdout, r)
}
