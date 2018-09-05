package urlutil

import (
	"fmt"
	"net/url"
	"github.com/qiniu/ts"
	"testing"
)

func TestValues(t *testing.T) {

	v := url.Values{
		"a.b": {"1/2"},
	}
	vstr := v.Encode()
	fmt.Println(vstr)

	v2, err := url.ParseQuery(vstr)
	if err != nil {
		ts.Fatal(t, "ParseQuery failed:", err)
	}
	fmt.Println(v2)
	vstr2 := v2.Encode()
	if vstr != vstr2 {
		ts.Fatal(t, "url.Values.Encode failed:", vstr, vstr2)
	}
}

type Foo struct {
	B string  `json:"b"`
	C byte    `json:"c"`
	D float32 `json:"d"`
}

func TestEncodeQuery(t *testing.T) {

	v := map[string]interface{}{
		"a": Foo{B: "a.b"},
		"b": &Foo{B: "b.b"},
		"c": map[string]interface{}{"d": "c.d", "e": 1},
		"d": []string{"d1", "d2"},
		"e": true,
	}
	vstr := EncodeQuery(v)
	fmt.Println(vstr)
}
