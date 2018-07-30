package flag

import (
	"fmt"
	"qbox.us/errors"
	"github.com/qiniu/ts"
	"testing"
)

func Test(t *testing.T) {

	var mode int
	var w, h int
	var format string
	var hasw, hash bool

	err := Parse(
		"imageView/2/w/200/h/300/format/jpg",
		SW{Val: &mode},
		SW{Name: "w", Val: &w, Has: &hasw},
		SW{Name: "h", Val: &h, Has: &hash},
		SW{Name: "format", Val: &format},
	)
	if err != nil {
		ts.Fatal(t, "Parse failed:", err)
	}
	if mode != 2 || w != 200 || h != 300 || format != "jpg" || !hasw || !hash {
		ts.Fatal(t, "Parse failed:", mode, w, h, format, hasw, hash)
	}

	err = Parse(
		"imageView/2/w/200/h/a300/format/jpg",
		SW{Val: &mode},
		SW{Name: "w", Val: &w, Has: &hasw},
		SW{Name: "h", Val: &h, Has: &hash},
		SW{Name: "format", Val: &format},
	)
	if err == nil {
		ts.Fatal(t, "Parse failed:", err)
	}
	fmt.Println(errors.Detail(err))

	{
		var version, thumbnail, format string
		var autoOrient bool
		var rotate float64
		err := Parse(
			"imageMogr/v2/auto-orient/true/thumbnail/300x300/rotate/90.102/format/JPG",
			SW{Val: &version},
			SW{Name: "auto-orient", Val: &autoOrient},
			SW{Name: "thumbnail", Val: &thumbnail},
			SW{Name: "format", Val: &format},
			SW{Name: "rotate", Val: &rotate},
		)
		if err != nil {
			ts.Fatal(t, "Parse failed:", err)
		}
		if version != "v2" || thumbnail != "300x300" || format != "JPG" || !autoOrient || rotate != 90.102 {
			ts.Fatal(t, "Parse failed:", version, thumbnail, format, autoOrient, rotate)
		}
	}
}
