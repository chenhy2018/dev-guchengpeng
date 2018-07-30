package freetype

import (
	"fmt"
	"io/ioutil"
	"os"
	"github.com/qiniu/ts"
	"testing"
)

func _TestFreetype(t *testing.T) {

	fonts := os.Getenv("HOME") + "/fonts/"
	os.Mkdir(fonts, 0777)

	ft, err := New()
	if err != nil {
		ts.Fatal(t, err)
	}
	defer ft.Release()

	face, err := ft.NewFace(fonts+"simhei.ttf", 0)
	if err != nil {
		ts.Fatal(t, err)
	}
	defer face.Release()

	fmt.Println(face.Detail())
}

func TestFreetypes(t *testing.T) {

	fonts := os.Getenv("HOME") + "/fonts/"
	os.Mkdir(fonts, 0777)

	ft, err := New()
	if err != nil {
		ts.Fatal(t, err)
	}
	defer ft.Release()

	fis, err := ioutil.ReadDir(fonts)
	if err != nil {
		ts.Fatal(t, "ioutil.ReadDir failed:", err)
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		var index int
		for {
			face, err := ft.NewFace(fonts+fi.Name(), index)
			if err != nil {
				ts.Log(t, fonts+fi.Name(), "-", err)
				break
			}
			rec := face.Detail()
			fmt.Println(rec.NumFaces, index, "[Face]:", rec.FamilyName, "[Style]:", rec.StyleName, "[StyleFlags]:", rec.StyleFlags)
			face.Release()
			break
			index++
			if index >= rec.NumFaces {
				break
			}
		}
	}
}
