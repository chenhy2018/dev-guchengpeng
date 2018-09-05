#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

SipAnswerCode cbOnIncomingCall(int _nAccountId, const char *_pFrom, const void *_pUser, const void *_pMedia, int *_pCallId)
{
        static int callid = 0;

        *_pCallId = callid++;
        printf("----->incoming call From %s to %d--------------userdata = %d\n", _pFrom, _nAccountId, *(int*)_pUser);
        return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        printf("_nAccountId = %d ---->>reg status = %d------------------------>userdata = %d\n", _nAccountId, _StatusCode,  *(int*)_pUser);
        if (_nAccountId == 3)
                assert(_StatusCode == SIP_TOO_MANY_ACCOUNT);
        if (_nAccountId == 6)
                assert(_StatusCode == SIP_USR_ALREADY_EXIST);
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{
        printf("Callid = %d-- nAccountId = %d --->state = %d, status code = %d------------>userdata = %d\n", _nCallId, _nAccountId, _State, _StatusCode,  *(int*)_pUser);
        if (_nCallId == 6)
                assert(_StatusCode == SIP_TOO_MANY_CALLS_FOR_ACCOUNT);
}

int main()
{
        SipInstanceConfig Config;
        Config.Cb.OnIncomingCall  = &cbOnIncomingCall;
        Config.Cb.OnCallStateChange = &cbOnCallStateChange;
        Config.Cb.OnRegStatusChange = &cbOnRegStatusChange;
        Config.nMaxCall = 4;
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
        AccountConfig.nMaxOngoingCall = 3;

        int ret = SipRegAccount(&AccountConfig, 2);
        assert(ret == SIP_SUCCESS);
        sleep(4);
        void *pLocalSdp;
        CreateTmpSDP(&pLocalSdp);
        ret = SipMakeNewCall(2, "<sip:1040@123.59.204.198;transport=tcp>", pLocalSdp, 3);
        assert(ret == SIP_SUCCESS);

        ret = SipMakeNewCall(2, "<sip:1038@123.59.204.198;transport=tcp>", pLocalSdp, 4);
        assert(ret == SIP_SUCCESS);

        ret = SipMakeNewCall(2, "<sip:1037@123.59.204.198;transport=tcp>", pLocalSdp, 5);
        assert(ret == SIP_SUCCESS);

        ret = SipMakeNewCall(2, "<sip:1036@123.59.204.198;transport=tcp>", pLocalSdp, 6);
        assert(ret == SIP_SUCCESS);
        sleep(5);
        SipHangUp(5);
        sleep(5);
        printf("setup a new call\n");
        ret = SipMakeNewCall(2, "<sip:1036@123.59.204.198;transport=tcp>", pLocalSdp, 7);
        assert(ret == SIP_SUCCESS);
        sleep(20);
        SipHangUp(3);
        SipHangUp(4);
        SipHangUp(7);
        sleep(10);
        ret = SipUnRegAccount(2);
        sleep(10);
        SipDestroyInstance();
        sleep(10);
}
