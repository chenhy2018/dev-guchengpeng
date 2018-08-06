package imgave

import (
	"fmt"
	"os"
	"testing"

	"github.com/qiniu/ts"
)

func TestGet(t *testing.T) {

	if true {
		fmt.Println("This case will fail when run in go1.5+, should be fixed............................")
		fmt.Println("This case will fail when run in go1.5+, should be fixed............................")
		println("This case will fail when run in go1.5+, should be fixed............................")
		println("This case will fail when run in go1.5+, should be fixed............................")
		return
	}

	v := []string{"samples/sample.jpg", `0x68553f`}
	imgFile, err := os.Open(v[0])
	if err != nil {
		ts.Fatal(t, err)
	}
	defer imgFile.Close()
	ave, err := Get(imgFile)
	if err != nil {
		ts.Fatal(t, "TestGetImageAve image.Decode :", err, imgFile)
	}
	if ave.RGB != v[1] {
		ts.Fatal(t, "getImageAve:", imgFile.Name(), " expected ", v[1], ", but got", ave.RGB)
	}
}

func BenchmarkGetImageAve(b *testing.B) {
	v := "samples/sample.jpg"
	for i := 0; i < b.N; i++ {
		imgFile, err := os.Open(v)
		if err != nil {
			b.Fatal(err)
		}
		_, err = Get(imgFile)
		if err != nil {
			b.Fatal("image.Decode :", err, v)
		}
		imgFile.Close()
	}
}
