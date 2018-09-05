package largefile

import (
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"qbox.us/cc"
	"qbox.us/errors"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/ts"
)

func init() {
	log.SetOutputLevel(0)
	runtime.GOMAXPROCS(8)
}

func TestBasic(t *testing.T) {

	home := os.Getenv("HOME")
	name := home + "/largefileTest"
	os.RemoveAll(name)

	f, err := Open(name, 3) // 8 bytes
	if err != nil {
		ts.Fatal(t, "Open failed:", err)
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	text1 := "Hello, world! Hello,"
	text2 := "Golang!!!"
	text := text1 + text2

	w := &cc.Writer{WriterAt: f}
	io.WriteString(w, text1)
	io.WriteString(w, text2)

	r := &cc.Reader{ReaderAt: f}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		ts.Fatal(t, "ReadAll failed:", errors.Detail(err))
	}
	if string(b) != text {
		ts.Fatal(t, "Read failed:", string(b))
	}

	err = f.Truncate(5)
	if err != nil {
		ts.Fatal(t, "Truncate failed:", err)
	}

	fsize, err := f.Size()
	if err != nil {
		ts.Fatal(t, "f.Size failed:", err)
	}
	if fsize != 5 {
		ts.Fatal(t, "f.Size != 5", fsize)
	}

	{
		r := &cc.Reader{ReaderAt: f}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			ts.Fatal(t, "ReadAll failed:", errors.Detail(err))
		}
		if string(b) != "Hello" {
			ts.Fatal(t, "Read failed:", string(b))
		}
	}
}
