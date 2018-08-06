// +build !windows

package osl

import (
	"os"
	"strings"
)

func GetAppDataDir(app string) (dir string, err error) {

	home := os.Getenv("HOME")

	dir = home + "/." + strings.ToLower(app)
	err = os.MkdirAll(dir, 0700)
	return
}
