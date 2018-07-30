package dev

/*
#include <winsock2.h>
#include <iphlpapi.h>
#include <stdlib.h>

//#pragma comment(lib, "IPHLPAPI.lib")

typedef unsigned char* LPUSTR;

int GetNicSerialNumber(LPUSTR szSerial, int* nLen)
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
		for (pAdapter = pAdapterInfo; pAdapter != NULL; pAdapter = pAdapter->Next) {
			if (pAdapter->Type == MIB_IF_TYPE_ETHERNET) {
				dwRetVal = pAdapter->AddressLength;
				if (dwRetVal < 6) continue;
				if (dwRetVal > 16) {
					dwRetVal = 16;
				}
				memcpy(szSerial, pAdapter->Address, dwRetVal);
				free(pAdapterInfo);
				*nLen = dwRetVal;
				return 0;
			}
		}
		dwRetVal = -1;
	}
	free(pAdapterInfo);
	return (int)(dwRetVal);
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
