#include <sys/ioctl.h>
#include <sys/socket.h>
#include <net/if.h>
#include <netinet/in.h>
#include <net/if_arp.h>
#include <string.h>
#include <stdio.h>

#define MAXINTERFACES  32

int main()
{
	int fd, i, intrface;
	struct ifreq buf[MAXINTERFACES];
	struct ifconf ifc;
	unsigned char* sn;

	if ((fd = socket(AF_INET, SOCK_DGRAM, 0)) < 0) {
		printf("Socket error\n");
		return -1;
	}

	ifc.ifc_len = sizeof buf;
	ifc.ifc_buf = (caddr_t)buf;
	if (ioctl(fd, SIOCGIFCONF, (char*)&ifc) != 0) {
		printf("Get lana state failure\n");
		close(fd);
		return -2;
	}

	intrface = ifc.ifc_len / sizeof(struct ifreq);
	for (i = 0; i < intrface; i++) {
		printf("%s\n", buf[i].ifr_name);
		if (ioctl(fd, SIOCGIFFLAGS, (char*)&buf[i]) != 0) {
			continue;
		}
		printf("0x%x\n", buf[i].ifr_flags);
		if (buf[i].ifr_flags & (IFF_LOOPBACK | IFF_NOARP)) {
			continue;
		}
		if (ioctl(fd, SIOCGIFHWADDR, (char*)&buf[i]) != 0) {
			continue;
		}
		sn = buf[i].ifr_hwaddr.sa_data;
		printf("%d) %02x:%02x:%02x:%02x:%02x:%02x\n", (int)buf[i].ifr_hwaddr.sa_family, sn[0], sn[1], sn[2], sn[3], sn[4], sn[5]);
	}
	return 0;
}

