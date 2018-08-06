#include <ifaddrs.h>
#include <net/ethernet.h>
#include <net/if_dl.h>
#include <net/route.h>
#include <stdio.h>

int main()
{
	struct ifaddrs* ifa;
	struct ifaddrs* ifa0;
	const struct sockaddr_dl* sdl;

	if (getifaddrs(&ifa0) != 0) {
		printf("getifaddrs failed!\n");
	}
	for (ifa = ifa0; ifa != NULL; ifa = ifa->ifa_next) {
		printf("ifa_name: %s 0x%x\n", ifa->ifa_name, ifa->ifa_flags);
		if (ifa->ifa_addr->sa_family == AF_LINK) {
			sdl = (const struct sockaddr_dl*)ifa->ifa_addr;
			if (sdl != NULL) {
				printf("sdl_nlen: %d, sdl_alen: %d, sdl_slen: %d\n",
					(int)(sdl->sdl_nlen), (int)(sdl->sdl_alen), (int)(sdl->sdl_slen));
				printf("sdl: %s\n", link_ntoa(sdl));
			}
		}
	}
	freeifaddrs(ifa0);
}