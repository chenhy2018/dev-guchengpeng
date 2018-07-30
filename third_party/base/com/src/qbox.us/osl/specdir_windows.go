// +build windows

package osl

/*
#define _WIN32_IE 0x0500
#include <shlobj.h>
#include <stdio.h>

static LPWSTR getLocalAppDataDir() {
	static WCHAR szPath[MAX_PATH];
	if (*szPath == '\0') {
		SHGetFolderPathW(NULL, CSIDL_PERSONAL, NULL, SHGFP_TYPE_CURRENT, szPath);
	}
	printf("szPath: %S\n", szPath);
	return szPath;
}
*/
import "C"
import "syscall"
import "unsafe"

func GetLocalAppDataDir() string {
	path := C.getLocalAppDataDir()
	path1 := ((*[1 << 20]uint16)(unsafe.Pointer(path)))[:]
	return syscall.UTF16ToString(path1)
}
