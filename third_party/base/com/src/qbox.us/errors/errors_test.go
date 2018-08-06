package errors

import (
	"io"
	"reflect"
	"syscall"
	"testing"

	"github.com/qiniu/ts"

	qerrors "github.com/qiniu/errors"
)

func TestErrors(t *testing.T) {

	if EIO != syscall.EIO {
		ts.Fatal(t, "not equal")
	}

	if EIO != RegisterError(syscall.EIO) {
		ts.Fatal(t, "RegisterError: not equal")
	}

	if ErrUnexpectedEOF != io.ErrUnexpectedEOF {
		ts.Fatal(t, "io err not equal")
	}

	if ErrUnexpectedEOF != RegisterError(io.ErrUnexpectedEOF) {
		ts.Fatal(t, "RegisterError: io err not equal")
	}
}

func TestErr(t *testing.T) {

	var err error

	err = Info(EIO, "abc")
	err = qerrors.Info(err, "efg").Detail(err)
	err = Info(err, "hijk").Detail(err)

	e := Err(err)
	if e != EIO {
		t.Fatal("Err(err) != EIO:", reflect.TypeOf(e))
	}
}
