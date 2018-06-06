#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>

SipAnswerCode cbOnIncomingCall(int _nAccountId, int _nCallId, const char *_pFrom, const void *_pUser, const void *_pMedia)
{
        printf("----->incoming call From %s to %d--------------userdata = %d\n", _pFrom, _nAccountId, *(int*)_pUser);
        PrintSdp(_pMedia);
        void *pLocalSdp;
        CreateTmpSDP(&pLocalSdp);
        SipAnswerCall(_nCallId, 200, NULL, pLocalSdp);
	return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        printf("---->>reg status = %d------------------------>userdata = %d\n", _StatusCode,  *(int*)_pUser);
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{
        PrintSdp(_pMedia);
        printf("----->state = %d, status code = %d------------>userdata = %d\n", _State, _StatusCode,  *(int*)_pUser);
}
int main()
{
        SipInstanceConfig Config;
        Config.Cb.OnIncomingCall  = &cbOnIncomingCall;
        Config.Cb.OnCallStateChange = &cbOnCallStateChange;
        Config.Cb.OnRegStatusChange = &cbOnRegStatusChange;
        Config.nMaxCall = 15;
        Config.nMaxAccount = 20;

        SipCreateInstance(&Config);
        SipSetLogLevel(4);
        sleep(2);
        int *user = (int *)malloc(sizeof(int));
        *user = 12345;
        int nid4 = -1;
        SipAccountConfig AccountConfig;
        AccountConfig.pUserName = "1004";
        AccountConfig.pPassWord = "1004";
        AccountConfig.pDomain = "192.168.56.102";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        int ret = SipAddNewAccount(&AccountConfig, &nid4);
        if (ret != SIP_SUCCESS)
                printf("Add sip account failed");


        /*
          int nid3 = SipAddNewAccount("1003", "1003", "192.168.56.102");

        int nid5 = SipAddNewAccount("1005", "1005", "192.168.56.102");
        int nid6 = SipAddNewAccount("1006", "1006", "192.168.56.102");
        int nid7 = SipAddNewAccount("1007", "1007", "192.168.56.102");
        int nid8 = SipAddNewAccount("1008", "1008", "192.168.56.102");
        int nid9 = SipAddNewAccount("1009", "1009", "192.168.56.102");
        int nid10 = SipAddNewAccount("1010", "1010", "192.168.56.102");

        SipRegAccount(nid3, 1);
        SipRegAccount(nid4, 1);
        SipRegAccount(nid5, 1);
        SipRegAccount(nid6, 1);
        SipRegAccount(nid7, 1);
        SipRegAccount(nid8, 1);
        SipRegAccount(nid9, 1);
        SipRegAccount(nid10, 1);
        */
        int ret1 = SipRegAccount(nid4, 1);

        sleep(5);

        void *pLocalSdp;
        CreateTmpSDP(&pLocalSdp);
        int nCallId1 = -1;
        SipMakeNewCall(nid4, "<sip:1003@192.168.56.102>", pLocalSdp, &nCallId1);

        sleep(20);
        SipHangUp(nCallId1);
        //SipDestroyInstance();
        while(1) {
        }
        return 0;
}
