#include <winsock2.h>
#include <iphlpapi.h>
#include <stdlib.h>

#pragma comment(lib, "IPHLPAPI.lib")

int GetNicSerialNumber(unsigned char szSerial[6])
{
	PIP_ADAPTER_INFO pAdapterInfo;
	PIP_ADAPTER_INFO pAdapter = NULL;
	DWORD dwRetVal = 0;

	ULONG ulOutBufLen = sizeof(IP_ADAPTER_INFO);
	pAdapterInfo = (IP_ADAPTER_INFO*)malloc(sizeof (IP_ADAPTER_INFO));

	// Make an initial call to GetAdaptersInfo to get
	// the necessary size into the ulOutBufLen variable
	if (GetAdaptersInfo(pAdapterInfo, &ulOutBufLen) == ERROR_BUFFER_OVERFLOW) {
		free(pAdapterInfo);
		pAdapterInfo = (IP_ADAPTER_INFO*)malloc(ulOutBufLen);
	}

	if ((dwRetVal = GetAdaptersInfo(pAdapterInfo, &ulOutBufLen)) == NO_ERROR) {
		pAdapter = pAdapterInfo;
		while (pAdapter) {
			if (pAdapter->AddressLength > 6) {
				pAdapter->AddressLength = 6;
			}
			memcpy(szSerial, pAdapter->Address, pAdapter->AddressLength);
			if (pAdapter->Type == MIB_IF_TYPE_ETHERNET)
				break;
			pAdapter = pAdapter->Next;
		}
	}
	if (pAdapterInfo)
		free(pAdapterInfo);
	return (int)(dwRetVal);
}

#include <stdio.h>

int main() {
	unsigned char sn[6];
	int err;
	err = GetNicSerialNumber(sn);
	if (err != 0) {
		fprintf(stderr, "Error: %d\n", err);
		return -1;
	}
	printf("%02x:%02x:%02x:%02x:%02x:%02x\n", sn[0], sn[1], sn[2], sn[3], sn[4], sn[5]);
}
