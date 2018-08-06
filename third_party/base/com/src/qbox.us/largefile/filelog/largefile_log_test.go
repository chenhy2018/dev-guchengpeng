package filelog

import (
	"io"
	"io/ioutil"
	"os"
	"qbox.us/cc"
	"qbox.us/errors"
	"github.com/qiniu/ts"
	"testing"
)

func Test(t *testing.T) {

	home := os.Getenv("HOME")
	name := home + "/largelogTest"
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

	io.WriteString(f, text1)
	io.WriteString(f, text2)

	r := &cc.Reader{ReaderAt: f.Instance}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		ts.Fatal(t, "ReadAll failed:", errors.Detail(err))
	}
	if string(b) != text {
		ts.Fatal(t, "Read failed:", string(b))
	}

	f.Close()
	f = nil

	f, err = Open(name, 3) // 8 bytes
	if err != nil {
		ts.Fatal(t, "Open failed:", err)
	}

	io.WriteString(f, text1)
	io.WriteString(f, text2)

	r = &cc.Reader{ReaderAt: f.Instance}
	b, err = ioutil.ReadAll(r)
	if err != nil {
		ts.Fatal(t, "ReadAll failed:", errors.Detail(err))
	}
	if string(b) != text+text {
		ts.Fatal(t, "Read failed:", string(b))
	}
}
