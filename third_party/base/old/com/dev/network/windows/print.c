#include <winsock2.h>
#include <iphlpapi.h>
#include <stdlib.h>

#pragma comment(lib, "IPHLPAPI.lib")

int main()
{
	PIP_ADAPTER_INFO pAdapterInfo;
	PIP_ADAPTER_INFO pAdapter = NULL;
	DWORD dwRetVal = 0;
	UINT i;

	ULONG ulOutBufLen = sizeof(IP_ADAPTER_INFO);
	pAdapterInfo = (IP_ADAPTER_INFO*)malloc(sizeof (IP_ADAPTER_INFO));
	if (pAdapterInfo == NULL) {
		printf("Error allocating memory needed to call GetAdaptersinfo\n");
		return 1;
	}

	// Make an initial call to GetAdaptersInfo to get
	// the necessary size into the ulOutBufLen variable
	if (GetAdaptersInfo(pAdapterInfo, &ulOutBufLen) == ERROR_BUFFER_OVERFLOW) {
		free(pAdapterInfo);
		pAdapterInfo = (IP_ADAPTER_INFO*)malloc(ulOutBufLen);
		if (pAdapterInfo == NULL) {
			printf("Error allocating memory needed to call GetAdaptersinfo\n");
			return 1;
		}
	}

	if ((dwRetVal = GetAdaptersInfo(pAdapterInfo, &ulOutBufLen)) == NO_ERROR) {
		pAdapter = pAdapterInfo;
		while (pAdapter) {
			printf("\tComboIndex: \t5d\n", pAdapter->ComboIndex);
			printf("\tAdapter Name: \t%s\n", pAdapter->AdapterName);
			printf("\tAdapter Desc: \t%s\n", pAdapter->Description);
			printf("\tAdapter Addr: \t");
			for (i = 0; i < pAdapter->AddressLength; i++) {
				if (i == (pAdapter->AddressLength - 1))
					printf("%.2X\n", (int) pAdapter->Address[i]);
				else
					printf("%.2X-", (int) pAdapter->Address[i]);
			}
			printf("\tIndex: \t%d\n", pAdapter->Index);
			printf("\tType: \t");
			switch (pAdapter->Type) {
			case MIB_IF_TYPE_OTHER:
				printf("Other\n");
				break;
			case MIB_IF_TYPE_ETHERNET:
				printf("Ethernet\n");
				break;
			case MIB_IF_TYPE_TOKENRING:
				printf("Token Ring\n");
				break;
			case MIB_IF_TYPE_FDDI:
				printf("FDDI\n");
				break;
			case MIB_IF_TYPE_PPP:
				printf("PPP\n");
				break;
			case MIB_IF_TYPE_LOOPBACK:
				printf("Lookback\n");
				break;
			case MIB_IF_TYPE_SLIP:
				printf("Slip\n");
				break;
			default:
				printf("Unknown type %ld\n", pAdapter->Type);
				break;
			}

			printf("\tIP Address: \t%s\n",
				   pAdapter->IpAddressList.IpAddress.String);
			printf("\tIP Mask: \t%s\n", pAdapter->IpAddressList.IpMask.String);

			printf("\tGateway: \t%s\n", pAdapter->GatewayList.IpAddress.String);
			printf("\t***\n");

			pAdapter = pAdapter->Next;
			printf("\n");
		}
	} else {
		printf("GetAdaptersInfo failed with error: %d\n", dwRetVal);
	}
	if (pAdapterInfo)
		free(pAdapterInfo);
}