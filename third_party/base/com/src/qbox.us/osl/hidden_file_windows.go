// +build windows

package osl

import (
	"syscall"
)

func SetHiddenFile(file string) (err error) {
	fileW := syscall.StringToUTF16Ptr(file)
	attrs, err := syscall.GetFileAttributes(fileW)
	if err != nil {
		return
	}
	if (attrs & syscall.FILE_ATTRIBUTE_HIDDEN) != 0 {
		return
	}
	err = syscall.SetFileAttributes(fileW, attrs|syscall.FILE_ATTRIBUTE_HIDDEN)
	return
}
