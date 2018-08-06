package largefile

import (
	"os"
	"github.com/qiniu/ts"
	"testing"
)

func log(f *Logger, msg string) {
	f.Log([]byte(msg))
}

func Test(t *testing.T) {

	home := os.Getenv("HOME")
	name := home + "/loggerTest"
	os.RemoveAll(name)

	f, err := Open(name, 3) // 8 bytes
	if err != nil {
		ts.Fatal(t, "Open failed:", err)
	}

	text1 := "Hello, world!!!"
	text2 := "Golang!!!"

	log(f, text1)
	f.Close()

	f, err = Open(name, 3)
	if err != nil {
		ts.Fatal(t, "Open failed:", err)
	}

	log(f, text2)
	defer f.Close()

	fsize, err := f.f.Size()
	if err != nil {
		ts.Fatal(t, "get fsize failed")
	}

	if int(fsize) != len(text1)+len(text2)+2 {
		ts.Fatal(t, "Size error")
	}
}
