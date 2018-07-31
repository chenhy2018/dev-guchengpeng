package sstore

import (
	"github.com/qiniu/ts"
	"testing"
)

func TestKeyFinder(t *testing.T) {
	kf := NewSimpleKeyFinder()
	kf.Add(1, []byte("foo"))
	kf.Add(2, []byte("bar"))
	if string(kf.Find(1)) != "foo" {
		ts.Fatal(t, "SimpleKeyFinder.Find failed")
	}
	if string(kf.Find(2)) != "bar" {
		ts.Fatal(t, "SimpleKeyFinder.Find failed")
	}
}
