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
        PrintSdp(_pMedia);
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
        PrintSdp(_pMedia);
        printf("----->state = %d, status code = %d------------>userdata = %d\n", _State, _StatusCode,  *(int*)_pUser);
}
int main()
{
        SipInstanceConfig Config;
        Config.Cb.OnIncomingCall  = &cbOnIncomingCall;
        Config.Cb.OnCallStateChange = &cbOnCallStateChange;
        Config.Cb.OnRegStatusChange = &cbOnRegStatusChange;
        Config.nMaxCall = 3;
        Config.nMaxAccount = 3;

        SipCreateInstance(&Config);
        SipSetLogLevel(6);
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

        int ret = SipRegAccount(&AccountConfig, 0);
        if (ret != SIP_SUCCESS)
                printf("Add 1039 failed, ret = %d\n", ret);
        printf("Add 1039 success\n");

        AccountConfig.pUserName = "1038";
        AccountConfig.pPassWord = "1038";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipRegAccount(&AccountConfig, 1);
        if (ret != SIP_SUCCESS)
                printf("Add 1038 failed ret = %d\n", ret);
        printf("Add 1038 success\n");

        AccountConfig.pUserName = "1037";
        AccountConfig.pPassWord = "1037";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret =  SipRegAccount(&AccountConfig, 5);
        if (ret != SIP_SUCCESS)
                printf("Add 1037 failed, ret = %d\n", ret);
        printf("Add 1037 success\n");

        AccountConfig.pUserName = "1036";
        AccountConfig.pPassWord = "1036";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipRegAccount(&AccountConfig, 3);
        assert(ret == SIP_SUCCESS);

        /*Add a already exist user */

        sleep(5);
        ret = SipUnRegAccount(0);
        assert(ret == SIP_SUCCESS);
        ret = SipUnRegAccount(1);
        assert(ret == SIP_SUCCESS);


        AccountConfig.pUserName = "1037";
        AccountConfig.pPassWord = "1037";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret =  SipRegAccount(&AccountConfig, 6);
        if (ret != SIP_SUCCESS)
                printf("Add 1037 failed, ret = %d\n", ret);
        printf("Add 1037 success\n");
        sleep(1);
        ret = SipUnRegAccount(5);
        assert(ret == SIP_SUCCESS);
        sleep(10);
        return 0;
}
