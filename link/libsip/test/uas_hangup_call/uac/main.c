#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>

int destroy = 0;
SipAnswerCode cbOnIncomingCall(int _nAccountId, int _nCallId, const char *_pFrom, const void *_pUser, const void *_pMedia)
{
        printf("----->incoming call From %s to %d--------------userdata = %d\n", _pFrom, _nAccountId, *(int*)_pUser);
        PrintSdp(_pMedia);
	return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        printf("---->>reg status = %d------------------------>userdata = %d\n", _StatusCode,  *(int*)_pUser);
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{
        PrintSdp(_pMedia);
        if (_State == INV_STATE_DISCONNECTED)
                destroy = 1;
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
        int nid = -1;
        SipAccountConfig AccountConfig;
        AccountConfig.pUserName = "1039";
        AccountConfig.pPassWord = "1039";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        int ret = SipAddNewAccount(&AccountConfig, &nid);
        if (ret != SIP_SUCCESS)
                printf("Add sip account failed");
        SipRegAccount(nid, 1);

        sleep(5);

        void *pLocalSdp;
        CreateTmpSDP(&pLocalSdp);
        int nCallId = -1;
        SipMakeNewCall(nid, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, &nCallId);
        sleep(2);
        sleep(10);
        while(1) {
                if (destory) {
                        sleep(10);
                        //SipDestroyInstance();
                        break;
                }
        }
        return 0;
}
