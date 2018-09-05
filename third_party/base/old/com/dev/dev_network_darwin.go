package dev

/*
#include <ifaddrs.h>
#include <net/ethernet.h>
#include <net/if_dl.h>
#include <net/route.h>
#include <string.h>

typedef unsigned char* LPUSTR;

int GetNicSerialNumber(LPUSTR szSerial, int* nLen)
{
	int rv;
	struct ifaddrs* ifa;
	struct ifaddrs* ifa0;
	const struct sockaddr_dl* sdl;

	rv = getifaddrs(&ifa0);
	if (rv != 0) {
		return rv;
	}
	for (ifa = ifa0; ifa != NULL; ifa = ifa->ifa_next) {
		if (ifa->ifa_addr->sa_family == AF_LINK) {
			sdl = (const struct sockaddr_dl*)ifa->ifa_addr;
			if (sdl != NULL) {
				rv = (int)(sdl->sdl_alen);
				if (rv < 6) continue;
				if (rv > 16)
					rv = 16;
				memcpy(szSerial, sdl->sdl_data + sdl->sdl_nlen, rv);
				freeifaddrs(ifa0);
				*nLen = rv;
				return 0;
			}
		}
	}
	freeifaddrs(ifa0);
	return -1;
}
*/
import "C"
import "os"
import "unsafe"

func GetNicSerialNumber() (sn []byte, err os.Error) {

	var buf [16]byte
	var nLen C.int
	rv := C.GetNicSerialNumber(C.LPUSTR(unsafe.Pointer(&buf[0])), &nLen)
	if rv != 0 {
		err = os.Errno(int(rv))
		return
	}
	sn = buf[:int(nLen)]
	return
}
