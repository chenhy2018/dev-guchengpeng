#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

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
        int nid1 = -1;
        SipAccountConfig AccountConfig;
        AccountConfig.pUserName = "1040";
        AccountConfig.pPassWord = "1040";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        int ret = SipAddNewAccount(&AccountConfig, &nid1);
        assert(ret == SIP_SUCCESS);
        ret = SipRegAccount(nid1, 1);
        assert(ret == SIP_SUCCESS);

        int nid2 = -1;
        AccountConfig.pUserName = "1038";
        AccountConfig.pPassWord = "1038";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipAddNewAccount(&AccountConfig, &nid2);
        assert(ret == SIP_SUCCESS);
        ret = SipRegAccount(nid2, 1);
        assert(ret == SIP_SUCCESS);

        int nid3 = -1;
        AccountConfig.pUserName = "1037";
        AccountConfig.pPassWord = "1037";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipAddNewAccount(&AccountConfig, &nid3);
        assert(ret == SIP_SUCCESS);
        ret = SipRegAccount(nid3, 1);
        assert(ret == SIP_SUCCESS);

        int nid4 = -1;
        AccountConfig.pUserName = "1036";
        AccountConfig.pPassWord = "1036";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipAddNewAccount(&AccountConfig, &nid4);
        assert(ret == SIP_SUCCESS);
        ret = SipRegAccount(nid4, 1);
        assert(ret == SIP_SUCCESS);

        sleep(10);
        while(1) {
                if (destroy) {
                        sleep(15);
                        SipDestroyInstance();
                        break;
                }
                if (incommingcall) {
                        incommingcall = 0;
                }
        }
        return 0;
}
