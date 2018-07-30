package ioutil

import (
	"io/ioutil"
	"os"
)

// -------------------------------------------------------------------------

const (
	NoFollowSymlink  = 0
	FollowSymlink    = 1
	FollowDirSymlink = 2
)

func ReadDir(path string, followSymlink int) (fis []os.FileInfo, err error) {

	fis, err = ioutil.ReadDir(path)
	if err != nil {
		return
	}

	if followSymlink != 0 {
		for i, fi := range fis {
			if (fi.Mode() & os.ModeSymlink) != 0 {
				fi2, err2 := os.Stat(path + "/" + fi.Name())
				if err2 == nil {
					if followSymlink == FollowDirSymlink && !fi2.IsDir() {
						continue
					}
					fis[i] = fi2
				}
			}
		}
	}
	return
}

// -------------------------------------------------------------------------
