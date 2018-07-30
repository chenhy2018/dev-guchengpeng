package ftutil

import (
	"io/ioutil"
	"qbox.us/errors"
	"qbox.us/freetype"
	"github.com/qiniu/log.v1"
	"strconv"
	"strings"
)

// --------------------------------------------------------------------
// type Fonts

type Fonts map[string]string

func NewFonts(ft freetype.Library, root string, fAlias bool) (r Fonts, err error) {

	fis, err := ioutil.ReadDir(root)

	if err != nil {
		err = errors.Info(err, "ioutil.ReadDir").Detail(err)
		return
	}

	root += "/"

	fonts := make(map[string]string)
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		fontFile := fi.Name()
		face, err2 := ft.NewFace(root+fontFile, 0)
		if err2 != nil {
			log.Warn(fontFile, "-", err2)
			continue
		}
		rec := face.Detail()
		faceName := strings.ToLower(rec.FamilyName) + ":" + strconv.Itoa(rec.StyleFlags)
		fonts[faceName] = fontFile
		face.Release()
	}
	r = Fonts(fonts)
	if fAlias {
		r.DefaultAlias()
	}
	return
}

type aliasItem struct {
	aliasName, faceName string
	styleFlags          int
}

func (fonts Fonts) DefaultAlias() {

	items := []aliasItem{
		{"宋体", "SimSun", 0},
		{"黑体", "SimHei", 0},
		{"楷体", "KaiTi", 0},
		{"仿宋", "FangSong", 0},
		{"微软雅黑", "Microsoft YaHei", 0},
	}
	for _, item := range items {
		fonts.Alias(item.aliasName, item.faceName, item.styleFlags)
	}
}

func (fonts Fonts) Alias(aliasName, faceName string, styleFlags int) (ok bool) {

	styleFlagsName := ":" + strconv.Itoa(styleFlags)
	faceName = strings.ToLower(faceName) + styleFlagsName
	fontFile, ok := fonts[faceName]
	if ok {
		aliasFaceName := strings.ToLower(aliasName) + styleFlagsName
		fonts[aliasFaceName] = fontFile
	}
	return
}

func (fonts Fonts) Get(faceName string, styleFlags int) (fontFile string, ok bool) {

	faceName = strings.ToLower(faceName) + ":" + strconv.Itoa(styleFlags)
	fontFile, ok = fonts[faceName]
	return
}

// --------------------------------------------------------------------
