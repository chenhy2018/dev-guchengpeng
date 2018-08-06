package lock

import (
	"os"
	"syscall"
)

var ErrLocked = syscall.EEXIST

func Acquire(file string) (err error) {

	f, err := os.OpenFile(file, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err == nil {
		f.Close()
	} else if e, ok := err.(*os.PathError); ok {
		err = e.Err
	}
	return
}

func Release(file string) (err error) {

	return os.Remove(file)
}
