#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

int destroy = 0;
SipAnswerCode cbOnIncomingCall(int _nAccountId, int _nCallId, const char *_pFrom, const void *_pUser, const void *_pMedia)
{
        printf("ncallId = %d ------>_nAccountId = %d----->incoming call From %s to %d--------------userdata = %d\n",  _nCallId, _nAccountId, _pFrom, _nAccountId, *(int*)_pUser);
       return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        printf("Account Id = %d ---->>reg status = %d------------------------>userdata = %d\n",_nAccountId,  _StatusCode,  *(int*)_pUser);
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{
       if (_State == INV_STATE_DISCONNECTED)
                destroy = 1;
        printf("ncallId = %d ------>_nAccountId = %d ---->state = %d, status code = %d----->userdata = %d\n", _nCallId, _nAccountId, _State, _StatusCode,  *(int*)_pUser);
}
int main()
{
        SipInstanceConfig Config;
        Config.Cb.OnIncomingCall  = &cbOnIncomingCall;
        Config.Cb.OnCallStateChange = &cbOnCallStateChange;
        Config.Cb.OnRegStatusChange = &cbOnRegStatusChange;
        Config.nMaxCall = 3;
        Config.nMaxAccount = 4;

        SipCreateInstance(&Config);
        SipSetLogLevel(1);
        sleep(2);
        int *user = (int *)malloc(sizeof(int));
        *user = 12345;
        int nid1 = -1;
        SipAccountConfig AccountConfig;
        AccountConfig.pUserName = "1039";
        AccountConfig.pPassWord = "1039";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        int ret = SipAddNewAccount(&AccountConfig, &nid1);
        assert(ret == SIP_SUCCESS);
        if (ret != SIP_SUCCESS)
                printf("Add sip account failed");
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

        if (ret != SIP_SUCCESS)
                printf("Add sip account failed");
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
        if (ret != SIP_SUCCESS)
                printf("Add sip account failed");
        ret = SipRegAccount(nid3, 1);
        assert(ret == SIP_SUCCESS);

        int nid4 = -1;
        AccountConfig.pUserName = "1036";
        AccountConfig.pPassWord = "1036";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipAddNewAccount(&AccountConfig, &nid4);
        if (ret != SIP_SUCCESS)
                printf("Add sip account failed");
        SipRegAccount(nid4, 1);

        sleep(10);
        void *pLocalSdp;
        CreateTmpSDP(&pLocalSdp);
        int nCallId1 = -1;
        int nCallId2 = -1;
        int nCallId3 = -1;
        int nCallId4 = -1;

        ret = SipMakeNewCall(nid1, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, &nCallId1);
        assert(ret == SIP_SUCCESS);

        ret = SipMakeNewCall(nid2, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, &nCallId2);
        assert(ret == SIP_SUCCESS);

        ret = SipMakeNewCall(nid3, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, &nCallId3);
        assert(ret == SIP_SUCCESS);

        ret = SipMakeNewCall(nid4, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, &nCallId4);
        assert(ret == SIP_TOO_MANY_CALLS_FOR_INSTANCE);

        SipHangUp(nCallId3);
        sleep(5);
        printf("setup a new call\n");
        ret = SipMakeNewCall(nid4, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, &nCallId4);
        assert(ret == SIP_SUCCESS);
        sleep(20);
        SipHangUp(nCallId1);
        SipHangUp(nCallId2);
        SipHangUp(nCallId4);
        sleep(10);
        ret = SipRegAccount(nid1, 0);
        ret = SipRegAccount(nid2, 0);
        ret = SipRegAccount(nid3, 0);
        ret = SipRegAccount(nid4, 0);
        sleep(10);
        SipDestroyInstance();
        sleep(10);
}
