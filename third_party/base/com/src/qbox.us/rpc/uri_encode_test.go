package rpc

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestEncodeURI(t *testing.T) {
	v1 := [][3]string{
		{"/home/bar", "!home!bar", "home/bar"},
		{"foo:bar", "!foo:bar", "foo:bar"},
		{"foo:", "!foo:", "foo:"},
		{":foo:bar", ":foo:bar", ":foo:bar"},
		{"http://bar", "http:!!bar", "http://bar"},
		{"https://bar", "https:!!bar", "https://bar"},
		{"!home$bar", "!'21home$bar", "!home$bar"},
		{"A~home@+bar", "!A~home@+bar", "A~home@+bar"},
		{"A*?home@#bar", "!A*'3Fhome@'23bar", "A*?home@#bar"},
	}
	for _, v := range v1 {

		s := EncodeURI(v[0])
		fmt.Println("EncodeURI:", v[0], "=>", s)
		if s != v[1] {
			t.Error("EncodeURI:", v[0], v[1], s)
			continue
		}

		s1, err := DecodeURI(s)
		if err != nil {
			t.Error("DecodeURI:", s, err)
		} else if s1 != v[2] {
			t.Error("DecodeURI:", s, v[2], len(s1), s1)
		}

		s2 := base64.URLEncoding.EncodeToString([]byte(v[0]))
		s3, err := DecodeURI(s2)
		if err != nil {
			t.Error("DecodeURI:", s, err)
		} else if s3 != v[0] {
			t.Error("DecodeURI:", s, v[0], len(s3), s3)
		}
	}
}
