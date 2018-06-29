#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include <pthread.h>

int callId[4] = {-1, -1, -1, -1};
pthread_mutex_t mutex1 = PTHREAD_MUTEX_INITIALIZER;
SipAnswerCode cbOnIncomingCall(int _nAccountId, const char *_pFrom, const void *_pUser, const void *_pMedia, int *_pCallId)
{
        static int id = 0;
        pthread_mutex_lock( &mutex1);
        callId[id] = id;
        *_pCallId = id++;
        printf("----->incoming call From %s to %d--------------userdata = %d\n", _pFrom, _nAccountId, *(int*)_pUser);
        pthread_mutex_unlock( &mutex1);
	return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        printf("_nAccountId = %d ---->>reg status = %d------------------------>userdata = %d\n", _nAccountId, _StatusCode,  *(int*)_pUser);
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{
        printf("Callid = %d----->state = %d, status code = %d------------>userdata = %d\n", _nCallId, _State, _StatusCode,  *(int*)_pUser);
}

void *AnswerCall(void *arg)
{
        void *pLocalSdp;
        CreateTmpSDP(&pLocalSdp);
        while (1) {
                for (int i = 0; i < 4; i++) {
                        if (callId[i] != -1) {
                                pthread_mutex_lock( &mutex1);
                                SipAnswerCall(callId[i], 200, NULL, pLocalSdp);
                                callId[i] = -1;
                                pthread_mutex_unlock( &mutex1);

                        }
                }
        }

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
        SipSetLogLevel(6);
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

        int ret = SipRegAccount(&AccountConfig, 1);
        assert(ret == SIP_SUCCESS);


        AccountConfig.pUserName = "1038";
        AccountConfig.pPassWord = "1038";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipRegAccount(&AccountConfig, 2);
        assert(ret == SIP_SUCCESS);

        AccountConfig.pUserName = "1037";
        AccountConfig.pPassWord = "1037";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipRegAccount(&AccountConfig, 3);
        assert(ret == SIP_SUCCESS);

        AccountConfig.pUserName = "1036";
        AccountConfig.pPassWord = "1036";
        AccountConfig.pDomain = "123.59.204.198";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        ret = SipRegAccount(&AccountConfig, 4);
        assert(ret == SIP_SUCCESS);
        pthread_t tid_1;
        pthread_create(&tid_1, NULL, AnswerCall, NULL);
        while(1) {

        }
        return 0;
}
