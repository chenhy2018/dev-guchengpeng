#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#define TEST1 0

int incommingcall = 0;
int nCallid = 0;
int destroy = 0;
int confirmed = 0;
int normaldisconnected = 0;
int abnormalDisconnected = 0;
static pthread_mutex_t lock = PTHREAD_MUTEX_INITIALIZER;

static pthread_cond_t Cond = PTHREAD_COND_INITIALIZER;
SipAnswerCode cbOnIncomingCall(int _nAccountId, const char *_pFrom, const void *_pUser, const void *_pMedia, int *_pCallId)
{

        static int callid = 0;
        pthread_mutex_lock(&lock);
        nCallid = callid;
        pthread_cond_signal(&Cond);
        *_pCallId = callid++;
        incommingcall++;
        pthread_mutex_unlock(&lock);
        return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        printf("_nAccountId = %d ---->>reg status = %d------------------------>userdata = %d\n", _nAccountId, _StatusCode,  *(int*)_pUser);
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{
        if (_State == 6) {
                if (_StatusCode == 200)
                        normaldisconnected++;
                else
                        abnormalDisconnected++;
        }
        if (_State == 5)
                confirmed++;

        if (_State == 6 || _State == 5)
                printf("Callid = %d---->ongoing call = %d, total confirmed = %d, normal dis = %d, abnormal dis = %d\n", _nCallId, incommingcall - (normaldisconnected + abnormalDisconnected), confirmed, normaldisconnected, abnormalDisconnected);
}

int main()
{
        SipInstanceConfig Config;
        Config.Cb.OnIncomingCall  = &cbOnIncomingCall;
        Config.Cb.OnCallStateChange = &cbOnCallStateChange;
        Config.Cb.OnRegStatusChange = &cbOnRegStatusChange;
        Config.nMaxCall = 200;
        Config.nMaxAccount = 20;

        SipCreateInstance(&Config);
        SipSetLogLevel(1);
        sleep(2);
        int *user = (int *)malloc(sizeof(int));
        *user = 12345;
        int nid = -1;
        SipAccountConfig AccountConfig;
        #if TEST1
        AccountConfig.pUserName = "1550";
        AccountConfig.pPassWord = "qUPaxZD6";
        #else
        AccountConfig.pUserName = "1650";
        AccountConfig.pPassWord = "sbemuRpa";
        #endif

        AccountConfig.pDomain = "180.97.147.174";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 200;

        int ret = SipRegAccount(&AccountConfig, 2);
        void *pLocalSdp;
        CreateTmpSDP(&pLocalSdp);

        while(1) {

                pthread_mutex_lock(&lock);
                pthread_cond_wait(&Cond, &lock);
                SipAnswerCall(nCallid, 200, NULL, pLocalSdp);
                pthread_mutex_unlock(&lock);
        }
        SipDestroyInstance();
        sleep(4);
        return 0;
}
