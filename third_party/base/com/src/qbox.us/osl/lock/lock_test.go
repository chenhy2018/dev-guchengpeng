package lock

import (
	"os"
	"github.com/qiniu/ts"
	"reflect"
	"testing"
)

func TestLock(t *testing.T) {

	home := os.Getenv("HOME")
	lockFile := home + "/lockTestFile"

	os.Remove(lockFile)

	err := Acquire(lockFile)
	if err != nil {
		ts.Fatal(t, "Acquire failed:", err)
	}

	err = Acquire(lockFile)
	if err == nil || err != ErrLocked {
		ts.Fatal(t, "Acquire 2 failed:", reflect.TypeOf(err), err)
	}

	err = Release(lockFile)
	if err != nil {
		ts.Fatal(t, "Release failed:", err)
	}
}
