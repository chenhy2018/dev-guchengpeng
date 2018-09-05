#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include <pthread.h>

int nid = -1;

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

void *RegThread( void *ptr )
{
        printf("--------------------------------------\n");
        int *user = (int *)malloc(sizeof(int));
        *user = 12345;
        SipAccountConfig AccountConfig;
        AccountConfig.pUserName = "1039";
        AccountConfig.pPassWord = "1039";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        int ret = SipRegAccount(&AccountConfig, 2);
        if (ret != SIP_SUCCESS)
                printf("Add sip account failed ret = %d\n", ret);
        return NULL;
}
void* unregister(void *ptr){
        SipUnRegAccount(2);
        return NULL;
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

        pthread_t thread1;
        pthread_create( &thread1, NULL, RegThread, NULL);
        pthread_join( thread1, NULL);
        sleep(2);
        pthread_t thread2;
        pthread_create( &thread2, NULL, unregister, NULL);
        pthread_join( thread2, NULL);
        sleep(2);
        SipDestroyInstance();
}
