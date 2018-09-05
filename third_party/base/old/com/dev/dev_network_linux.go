package dev

/*
#include <sys/ioctl.h>
#include <sys/socket.h>
#include <net/if.h>
#include <netinet/in.h>
#include <net/if_arp.h>
#include <string.h>
#include <unistd.h>

#define MAXINTERFACES  32

typedef const char* LPCSTR;
typedef unsigned char* LPUSTR;

LPCSTR GetNicSerialNumber(LPUSTR szSerial)
{
	int fd, i, intrface;
	struct ifreq buf[MAXINTERFACES];
	struct ifconf ifc;

	if ((fd = socket(AF_INET, SOCK_DGRAM, 0)) < 0) {
		return "Socket error";
	}

	ifc.ifc_len = sizeof buf;
	ifc.ifc_buf = (caddr_t)buf;
	if (ioctl(fd, SIOCGIFCONF, (char*)&ifc) != 0) {
		close(fd);
		return "Get lana state failure";
	}

	intrface = ifc.ifc_len / sizeof(struct ifreq);
	for (i = 0; i < intrface; i++) {
		if (ioctl(fd, SIOCGIFFLAGS, (char*)&buf[i]) != 0) {
			continue;
		}
		if (buf[i].ifr_flags & (IFF_LOOPBACK | IFF_NOARP)) {
			continue;
		}
		if (ioctl(fd, SIOCGIFHWADDR, (char*)&buf[i]) != 0) {
			continue;
		}
		memcpy(szSerial, buf[i].ifr_hwaddr.sa_data, 6);
		close(fd);
		return NULL;
	}
	close(fd);
	return "Network device not found";
}
*/
import "C"
import "os"
import "unsafe"

type Error struct {
	msg C.LPCSTR
}

func (e Error) String() string {
	return C.GoString(e.msg)
}

func GetNicSerialNumber() (sn []byte, err os.Error) {

	var buf [6]byte
	emsg := C.GetNicSerialNumber(C.LPUSTR(unsafe.Pointer(&buf[0])))
	if emsg != nil {
		err = Error{emsg}
		return
	}
	sn = buf[:]
	return
}
