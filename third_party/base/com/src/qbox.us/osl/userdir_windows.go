// +build windows

package osl

import (
	"os"
	"strings"
)

func GetAppDataDir(app string) (dir string, err error) {

	hidden := false
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		appData = os.Getenv("APPDATA")
	}

	if appData != "" {
		dir = appData + "/" + app + "/conf"
	} else {
		home := os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		}
		dir = home + "/." + strings.ToLower(app)
		hidden = true
	}
	err = os.MkdirAll(dir, 0700)
	if hidden {
		SetHiddenFile(dir)
	}
	return
}
