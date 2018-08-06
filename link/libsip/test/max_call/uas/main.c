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
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia) {
        printf("ncallId = %d ------>_nAccountId = %d ---->state = %d, status code = %d----->userdata = %d\n", _nCallId, _nAccountId, _State, _StatusCode,  *(int*)_pUser);
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

        int ret = SipRegAccount(&AccountConfig, 1);
        assert(ret == SIP_SUCCESS);


        while(1) {
        }
        return 0;
}
