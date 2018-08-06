package serverstat

import (
	"fmt"
	"github.com/qiniu/ts"
	"testing"
	"time"
)

// ----------------------------------------------------------

type foo struct {
}

func (r foo) ServiceStat() interface{} {

	return map[string]interface{}{
		"a": 1,
		"b": map[string]interface{}{
			"c": 2,
		},
	}
}

// ----------------------------------------------------------

func doTest(t *testing.T) {

	data, err := Get("http://localhost:3579/stat")
	if err != nil {
		ts.Fatal(t, err)
	}
	fmt.Println(string(data))

	if string(data) != `{"a":1,"b":{"c":2}}` {
		ts.Fatal(t, "get failed")
	}

	data, err = Get("http://localhost:3579/stat?q=b.c")
	if err != nil {
		ts.Fatal(t, err)
	}
	fmt.Println(string(data))

	if string(data) != `2` {
		ts.Fatal(t, "get failed")
	}

	data, err = Get("http://localhost:3579/stat?dividend=a&divisor=b.c")
	if err != nil {
		ts.Fatal(t, err)
	}
	fmt.Println(string(data))

	if string(data) != `0.5` {
		ts.Fatal(t, "get failed")
	}
}

func TestServerStat(t *testing.T) {

	go Run("localhost:3579", foo{})
	time.Sleep(1e9)

	doTest(t)
}

// ----------------------------------------------------------
