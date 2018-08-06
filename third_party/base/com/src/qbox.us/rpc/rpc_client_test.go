package rpc

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

func foo(w ResponseWriter, req Request) {
	w.Reply(map[string]interface{}{"info": "Call method foo", "query": req.Query})
}

type Object struct {
}

func (p *Object) ServeHTTP(w ResponseWriter, req Request) {
	req2, _ := ioutil.ReadAll(req.Body)
	w.Reply(map[string]interface{}{"info": "Call method object", "req": string(req2)})
}

var done = make(chan bool)

func server(t *testing.T) {
	HandleFunc("foo", foo)
	Handle("object", new(Object))
	err := Run(":8789", nil)
	if err != nil {
		t.Fatal("ListenAndServe:", err)
	}
}

func TestCall(t *testing.T) {
	go server(t)
	time.Sleep(1e9)
	param := "http:**localhost:8888*abc:def,g;+&$=foo*~!*~!"
	var r interface{}
	_, err := Call(&r, "http://localhost:8789/foo/"+param)
	if err != nil {
		t.Fatal(err)
	}
	if r2, ok := r.(map[string]interface{}); ok {
		if info, ok := r2["info"]; ok {
			if info.(string) != "Call method foo" {
				t.Fatal("Info fail")
			}
		} else {
			t.Fatal("Error get info")
		}
		if query, ok := r2["query"]; ok {
			if q := query.([]interface{}); ok {
				fmt.Println(q)
				if q[0].(string) != "foo" || q[1].(string) != param {
					t.Fatal("query fail:", q[1].(string))
				}
			} else {
				t.Fatal("Query fail2")
			}
		} else {
			t.Fatal("Error get query")
		}
	} else {
		t.Fatal("Error to map[string] interface{}")
	}

	r = nil
	_, err = CallWithJson(&r, "http://localhost:8789/object", map[string]int{"a": 1, "b": 2})
	if err != nil {
		t.Fatal(err)
	}
	if r2, ok := r.(map[string]interface{}); ok {
		if info, ok := r2["info"]; ok {
			if info.(string) != "Call method object" {
				t.Fatal("Info fail")
			}
		} else {
			t.Fatal("Error get info")
		}
		if req, ok := r2["req"]; ok {
			if req != "{\"a\":1,\"b\":2}" {
				t.Fatal(req)
			}
		} else {
			t.Fatal("Error get query")
		}
	} else {
		t.Fatal("Error to map[string] interface{}")
	}
}
