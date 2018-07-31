package urlquery

import (
	"fmt"
	"testing"
)

func TestURLQuery(t *testing.T) {
	queryCase := []string{
		"",
		"token=a/b/c",
		"exif",
		"imageView/xcf0==/w/500/h/400",
		"imageView/xcf0=/w/500/h/400&token=a/b/c==&sp=yyy",
		"imageMogr/v2/thumbnail/200x100%3C&sp=yyy",
		"imageMogr/v2/thumbnail/200x100%3E&sp=yyy",
		"token=a/b/c&imageView/xcf0==/w/500/h/400&sp=yyy==",
		"token=xxx=&sp=yyy&imageView/xcf0/w/500/h/400",
		"md2html&X-Qiniu-Redirect-Token=xxx",
	}
	wantCmd := []string{
		"",
		"",
		"exif",
		"imageView/xcf0==/w/500/h/400",
		"imageView/xcf0=/w/500/h/400",
		"imageMogr/v2/thumbnail/200x100<",
		"imageMogr/v2/thumbnail/200x100>",
		"imageView/xcf0==/w/500/h/400",
		"imageView/xcf0/w/500/h/400",
		"md2html",
	}
	wantM := []map[string]string{
		map[string]string{
			"token": "", "sp": "",
		},
		map[string]string{
			"token": "a/b/c", "sp": "",
		},
		map[string]string{
			"token": "", "sp": "",
		},
		map[string]string{
			"token": "", "sp": "",
		},
		map[string]string{
			"token": "a/b/c==", "sp": "yyy",
		},
		map[string]string{
			"token": "", "sp": "yyy",
		},
		map[string]string{
			"token": "", "sp": "yyy",
		},
		map[string]string{
			"token": "a/b/c", "sp": "yyy==",
		},
		map[string]string{
			"token": "xxx=", "sp": "yyy",
		},
		map[string]string{
			"token": "", "sp": "",
		},
	}
	for i, query := range queryCase {
		m, cmd, err := ParseQuery(query)
		if err != nil {
			t.Fatal("ParseQuery failed:", err)
		}
		fmt.Printf("cmd:%s, m:%v\n", cmd, m)
		if cmd != wantCmd[i] {
			t.Fatalf("ParseQuery get cmd failed, got %s, want %s", cmd, wantCmd[i])
		}
		if m.Get("token") != wantM[i]["token"] {
			t.Fatalf("ParseQuery get token failed, got %s, want %s", m.Get("token"), wantM[i]["token"])
		}
		if m.Get("sp") != wantM[i]["sp"] {
			t.Fatalf("ParseQuery get sp failed, got %s, want %s", m.Get("sp"), wantM[i]["sp"])
		}
	}
}
