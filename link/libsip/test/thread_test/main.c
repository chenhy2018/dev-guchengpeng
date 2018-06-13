#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include <pthread.h>

int nid = -1;

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
        printf("----->state = %d, status code = %d------------>userdata = %d\n", _State, _StatusCode,  *(int*)_pUser);
}
void *print_message_function( void *ptr )
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

        int ret = SipAddNewAccount(&AccountConfig, &nid);
        if (ret != SIP_SUCCESS)
                printf("Add sip account failed ret = %d\n", ret);
        return NULL;
}
void* unregister(void *ptr){
        SipRegAccount(nid, 0);
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
        SipSetLogLevel(4);
        sleep(2);

        pthread_t thread1;
        pthread_create( &thread1, NULL, print_message_function, NULL);
        pthread_join( thread1, NULL);

        sleep(4);
        SipRegAccount(nid, 1);
        sleep(4);

        pthread_t thread2;
        pthread_create( &thread2, NULL, unregister, NULL);
        pthread_join( thread2, NULL);

        SipDestroyInstance();
}
