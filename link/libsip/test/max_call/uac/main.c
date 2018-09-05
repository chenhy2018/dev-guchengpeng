#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

int destroy = 0;

SipAnswerCode cbOnIncomingCall(int _nAccountId, const char *_pFrom, const void *_pUser, const void *_pMedia, int *_pCallId)
{
        static int callid = 0;

        *_pCallId = callid++;
        printf("----->incoming call From %s to %d--------------userdata = %d\n", _pFrom, _nAccountId, *(int*)_pUser);
        PrintSdp(_pMedia);
	return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        printf("_nAccountId = %d ---->>reg status = %d------------------------>userdata = %d\n", _nAccountId, _StatusCode,  *(int*)_pUser);
}


void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia){
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

        int ret = SipRegAccount(&AccountConfig, 10);
        assert(ret == SIP_SUCCESS);

        AccountConfig.pUserName = "1038";
        AccountConfig.pPassWord = "1038";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipRegAccount(&AccountConfig, 11);
        assert(ret == SIP_SUCCESS);

        AccountConfig.pUserName = "1037";
        AccountConfig.pPassWord = "1037";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipRegAccount(&AccountConfig, 14);
        assert(ret == SIP_SUCCESS);

        int nid4 = -1;
        AccountConfig.pUserName = "1036";
        AccountConfig.pPassWord = "1036";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipRegAccount(&AccountConfig, 20);
        assert(ret == SIP_SUCCESS);

        sleep(10);
        void *pLocalSdp;
        CreateTmpSDP(&pLocalSdp);

        ret = SipMakeNewCall(10, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, 10);
        assert(ret == SIP_SUCCESS);

        ret = SipMakeNewCall(11, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, 11);
        assert(ret == SIP_SUCCESS);

        ret = SipMakeNewCall(14, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, 14);
        assert(ret == SIP_SUCCESS);

        ret = SipMakeNewCall(20, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, 20);

        SipHangUp(14);
        sleep(5);
        printf("setup a new call\n");
        ret = SipMakeNewCall(20, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, 21);
        assert(ret == SIP_SUCCESS);
        sleep(20);
        SipHangUp(21);
        SipHangUp(11);
        SipHangUp(10);
        sleep(10);
        ret = SipUnRegAccount(10);
        ret = SipUnRegAccount(11);
        ret = SipUnRegAccount(14);
        ret = SipUnRegAccount(20);
        sleep(10);
        SipDestroyInstance();
        sleep(10);
}
