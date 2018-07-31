package cc

import (
	"os"
)

func GetConfigDir(app string) (dir string, err os.Error) {

	home := os.Getenv("HOME")

	dir = home + "/." + app
	err = os.MkdirAll(dir, 0777)
	return
}
