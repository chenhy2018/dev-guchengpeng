#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>

int incommingcall = 0;
int nCallid = 0;
int destroy = 0;
int confirmed = 0;
int disconnected = 0;
int answered = 0;
#define  MAXCALL  200

static pthread_mutex_t lock = PTHREAD_MUTEX_INITIALIZER;

static pthread_cond_t Cond = PTHREAD_COND_INITIALIZER;
int CallState[MAXCALL];

void initCallState() {
        int i;
        for(i = 0; i < MAXCALL; i++) {
                CallState[i] = -1;
        }
}

int  GetCallId() {
        int i;
        for (i = 0; i < MAXCALL; i++) {
                if (CallState[i] == -1) {
                        return i;
                }
        }
        return -1;
}

SipAnswerCode cbOnIncomingCall(int _nAccountId, const char *_pFrom, const void *_pUser, const void *_pMedia, int *_pCallId)
{

        pthread_mutex_lock(&lock);
        int rrid = GetCallId();
        pthread_mutex_unlock(&lock);

        if (rrid == -1)
                return 500;

        *_pCallId = rrid;
        incommingcall++;
        return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        printf("_nAccountId = %d ---->>reg status = %d------------------------>userdata = %d\n", _nAccountId, _StatusCode,  *(int*)_pUser);
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{
        pthread_mutex_lock(&lock);
        CallState[_nCallId] = _State;
        if (_State == 6) {
                disconnected++;
                CallState[_nCallId] = -1;
        }
        if (_State == 5)
                confirmed++;

        if (_State == 6 || _State == 5)
                printf("Callid = %d---->ongoing call = %d, total confirmed = %d, disconnected = %d\n", _nCallId, incommingcall - disconnected, confirmed, disconnected);
        pthread_mutex_unlock(&lock);
}


static int lastAnswer;
int  getNextAnsweredCallId() {
        int i;
        for (i = lastAnswer; i < MAXCALL; i++) {
                if (CallState[i] == INV_STATE_EARLY) {
                        lastAnswer += 1;
                        return i;
                }
        }
        lastAnswer = 0;
        return -1;
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
        initCallState();
        sleep(2);
        int *user = (int *)malloc(sizeof(int));
        *user = 12345;
        int nid = -1;
        SipAccountConfig AccountConfig;
        AccountConfig.pUserName = "1550";
        AccountConfig.pPassWord = "qUPaxZD6";
        AccountConfig.pDomain = "180.97.147.174";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 200;

        int ret = SipRegAccount(&AccountConfig, 2);
        void *pLocalSdp;
        CreateTmpSDP(&pLocalSdp);

        while(1) {
                pthread_mutex_lock(&lock);
                int callId = getNextAnsweredCallId();
                pthread_mutex_unlock(&lock);

                if (callId != -1) {
                        pthread_mutex_lock(&lock);
                        CallState[callId] = 7;
                        pthread_mutex_unlock(&lock);
                        usleep(50);
                        SipAnswerCall(callId, 200, NULL, pLocalSdp);
                }
                sleep(1);
                /*
                pthread_mutex_lock(&lock);
                pthread_cond_wait(&Cond, &lock);
                SipAnswerCall(nCallid, 200, NULL, pLocalSdp);
                answered = 0;
                pthread_mutex_unlock(&lock);
                */
        }
        SipDestroyInstance();
        sleep(4);
        return 0;
}
