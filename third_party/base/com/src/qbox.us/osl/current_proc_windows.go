// +build ignore

package osl

/*
#include <windows.h>
*/
import "C"

func GetCurrentProcessId() int {
	return int(C.GetCurrentProcessId())
}
