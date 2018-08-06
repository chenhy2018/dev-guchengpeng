package main

import (
	"errors"
	"fmt"
	gioutil "io/ioutil"
	"os"
	"qbox.us/cc/ioutil"
	"strings"
)

// ----------------------------------------------------------

var ErrNotFound = errors.New("go package not found in $GOPATH")

func findPackage(pkg string) error {

	d, err := os.Stat(pkg)
	if err != nil {
		return err
	}
	if m := d.Mode(); m.IsDir() {
		return nil
	}
	return os.ErrPermission
}

func LookPackage(pkg string) (string, error) {

	pathenv := os.Getenv("GOPATH")
	for _, dir := range strings.Split(pathenv, ":") {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		path := dir + "/src/" + pkg
		if err := findPackage(path); err == nil {
			return path, nil
		}
	}
	return "", ErrNotFound
}

// ----------------------------------------------------------

func CopyFile(dest, src string) (err error) {

	data, err := gioutil.ReadFile(src)
	if err != nil {
		return
	}

	return gioutil.WriteFile(dest, data, 0666)
}

func CopyPackage(dest, src string) (err error) {

	err = os.MkdirAll(dest, 0777)
	if err != nil {
		return
	}
	dest += "/"

	fis, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}
	src += "/"

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		err1 := CopyFile(dest+name, src+name)
		if err1 != nil {
			err = err1
			fmt.Fprintln(os.Stderr, "[WARN] Copy file failed:", err)
		}
	}
	return
}

// ----------------------------------------------------------

func main() {

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: gofetch <DestPath> <Pkg1> ... <PkgN>")
		return
	}

	destdir := os.Args[1] + "/src"
	err := os.MkdirAll(destdir, 0777)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	destdir += "/"

	for _, pkg := range os.Args[2:] {
		srcpkg, err := LookPackage(pkg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[WARN] Package not found:", pkg)
			continue
		}
		err = CopyPackage(destdir+pkg, srcpkg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[WARN] Copy package failed:", err)
			continue
		}
	}
}

// ----------------------------------------------------------
