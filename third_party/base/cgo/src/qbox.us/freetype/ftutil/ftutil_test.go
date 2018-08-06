package ftutil

import (
	"fmt"
	"os"
	"qbox.us/freetype"
	"github.com/qiniu/ts"
	"testing"
)

func TestFonts(t *testing.T) {

	root := os.Getenv("HOME") + "/fonts/"
	os.Mkdir(root, 0777)

	ft, err := freetype.New()
	if err != nil {
		ts.Fatal(t, err)
	}
	defer ft.Release()

	fonts, err := NewFonts(ft, root, true)
	if err != nil {
		ts.Fatal(t, err)
	}

	for k, v := range fonts {
		fmt.Println(k, "-", v)
	}
}
