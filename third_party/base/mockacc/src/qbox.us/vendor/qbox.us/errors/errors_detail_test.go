package errors

import (
	"fmt"
	"testing"
)

func Foo(a int, b string) error {

	return Info(EINVAL, "Foo", a, b)
}

func Bar(a int, b string, c float32) error {

	err := Foo(a, b)
	return Info(err, "Bar", "a:", a, "b:", b, "c:", c).Detail(err).Warn()
}

func TestError(t *testing.T) {

	err := Bar(1, "hello", 3.2)
	fmt.Println("Bar error:", err)

	switch Err(err) {
	case EINVAL:
	default:
		t.Fatal("errors failed")
	}
}
