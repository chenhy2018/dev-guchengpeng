package filepath

import (
	"errors"
	"os"
	"path/filepath"
	"qbox.us/path/ioutil"
)

var SkipDir = errors.New("skip this directory")

func Walk(path string, walkFn func(path string, info os.FileInfo, err error) error) error {

	info, err := os.Lstat(path)
	if err != nil {
		return walkFn(path, nil, err)
	}
	return walk(path, info, walkFn, 0)
}

func WalkFollowSymlink(path string, walkFn func(path string, info os.FileInfo, err error) error) error {

	info, err := os.Stat(path)
	if err != nil {
		return walkFn(path, nil, err)
	}
	return walk(path, info, walkFn, ioutil.FollowSymlink)
}

func WalkFollowDirSymlink(path string, walkFn func(path string, info os.FileInfo, err error) error) error {

	info, err := os.Stat(path)
	if err != nil {
		return walkFn(path, nil, err)
	}
	return walk(path, info, walkFn, ioutil.FollowDirSymlink)
}

func walk(
	path string, info os.FileInfo,
	walkFn func(path string, info os.FileInfo, err error) error, followSymlink int) error {

	err := walkFn(path, info, nil)
	if err != nil {
		if info.IsDir() && err == SkipDir {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return nil
	}

	infos, err := ioutil.ReadDir(path, followSymlink)
	if err != nil {
		return walkFn(path, info, err)
	}

	for _, e := range infos {
		if err = walk(filepath.Join(path, e.Name()), e, walkFn, followSymlink); err != nil {
			return err
		}
	}
	return nil
}
