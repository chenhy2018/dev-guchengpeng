package qconfenc

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestEscape(t *testing.T) {

	cases := []string{"a", "aa", "aaa", "aaaa"}
	for _, c := range cases {
		venc := base64.URLEncoding.EncodeToString([]byte(c))
		venc2 := strings.TrimRight(venc, "=")
		venc3 := Escape(c)
		if venc2 != venc3 {
			t.Fatal("Escape:", c, venc, venc2, venc3)
		}
		c2, err := Unescape(venc3)
		if err != nil || c2 != c {
			t.Fatal("Unescape:", venc3, c2, c, err)
		}
	}
}

func TestUnescapeMapKeys(t *testing.T) {

	maps := make(map[string]int)
	cases := []string{"a", "aa", "aaa", "aaaa"}
	for i, c := range cases {
		maps[Escape(c)] = i
	}
	fmt.Println(maps)

	UnescapeMapKeys(&maps)
	fmt.Println(maps)
	if len(maps) != len(cases) {
		t.Fatal("UnescapeMapKeys:", maps)
	}
	for i, c := range cases {
		if maps[c] != i {
			t.Fatal("UnescapeMapKeys:", c, i, maps[c])
		}
	}
}
