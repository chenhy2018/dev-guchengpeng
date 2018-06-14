#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>

int incommingcall = 0;
int nCallid = 0;
int destroy = 0;
int confirmed = 0;
SipAnswerCode cbOnIncomingCall(int _nAccountId, int _nCallId, const char *_pFrom, const void *_pUser, const void *_pMedia)
{
        printf("ncallId = %d ------>_nAccountId = %d----->incoming call From %s to %d--------------userdata = %d\n",  _nCallId, _nAccountId, _pFrom, _nAccountId, *(int*)_pUser);
        incommingcall = 1;
        nCallid = _nCallId;
        void *pLocalSdp;
        CreateTmpSDP(&pLocalSdp);
        SipAnswerCall(nCallid, 200, NULL, pLocalSdp);
	return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        printf("Account Id = %d ---->>reg status = %d------------------------>userdata = %d\n",_nAccountId,  _StatusCode,  *(int*)_pUser);
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{
        printf("ncallId = %d ------>_nAccountId = %d ---->state = %d, status code = %d----->userdata = %d\n", _nCallId, _nAccountId, _State, _StatusCode,  *(int*)_pUser);
        if (_State == INV_STATE_DISCONNECTED)
                destroy = 1;
        if (_State == INV_STATE_CONFIRMED)
                confirmed = 1;
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
        SipSetLogLevel(1);
        sleep(2);
        int *user = (int *)malloc(sizeof(int));
        *user = 12345;
        int nid = -1;
        SipAccountConfig AccountConfig;
        AccountConfig.pUserName = "1040";
        AccountConfig.pPassWord = "1040";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 4;

        int ret = SipAddNewAccount(&AccountConfig, &nid);
        if (ret != SIP_SUCCESS)
                printf("Add sip account failed");
        int ret1 = SipRegAccount(nid, 1);

        while(1) {
                if (destroy) {
                        sleep(10);
                        ret = SipRegAccount(nid, 0);
                        sleep(10);
                        SipDestroyInstance();
                        break;
                }
                if (incommingcall) {
                        incommingcall = 0;
                }
        }
        return 0;
}
