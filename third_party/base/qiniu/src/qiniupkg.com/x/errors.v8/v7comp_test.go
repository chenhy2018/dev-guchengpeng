package errors

import (
	"syscall"
	"testing"

	gerrors "github.com/qiniu/errors"
	qerrors "qbox.us/errors"
	errors "qiniupkg.com/x/errors.v7"
)

func TestSys(t *testing.T) {

	if !IsBadRequest(syscall.EINVAL) {
		t.Fatal("IsBadRequest?")
	}

	if !IsAlreadyExists(syscall.EEXIST) {
		t.Fatal("IsAlreadyExists?")
	}

	if !IsNotFound(syscall.ENOENT) {
		t.Fatal("IsNotFound?")
	}
}

func TestV7(t *testing.T) {

	e := qerrors.Info(syscall.EINVAL, "bad request")
	if !IsBadRequest(e) {
		t.Fatal("IsBadRequest?")
	}

	e2 := gerrors.Info(syscall.EINVAL, "gerror")
	if !IsBadRequest(e2) {
		t.Fatal("IsBadRequest?")
	}

	e3 := errors.Info(syscall.EINVAL, "error")
	if !IsBadRequest(e3) {
		t.Fatal("IsBadRequest?")
	}
}
