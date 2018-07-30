package uri

import (
	"testing"
)

func TestEncode(t *testing.T) {
	v1 := [][2]string{
		{"/home/bar", "!home!bar"},
		{"!home$bar", "'21home$bar"},
		{"A~home@+bar", "A~home@+bar"},
		{"A*?home@#bar", "A*'3Fhome@'23bar"},
	}
	for _, v := range v1 {
		s := encode(v[0])
		if s != v[1] {
			t.Fatal("Encode:", v[0], v[1], s)
		}
		s1, err := decode(s)
		if err != nil {
			t.Fatal("Decode:", s, err)
		}
		if s1 != v[0] {
			t.Fatal("Decode:", s, v[0], len(s1), s1)
		}
	}
}
