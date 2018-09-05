#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>

int incommingcall = 0;
int nCallid = 0;
int destroy = 0;
int confirmed = 0;

SipAnswerCode cbOnIncomingCall(int _nAccountId, const char *_pFrom, const void *_pUser, const void *_pMedia, int *_pCallId)
{
        static int callid = 0;
        nCallid = callid;
        *_pCallId = callid++;
        incommingcall = 1;

        printf("----->incoming call From %s to %d--------------userdata = %d\n", _pFrom, _nAccountId, *(int*)_pUser);
        return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        printf("_nAccountId = %d ---->>reg status = %d------------------------>userdata = %d\n", _nAccountId, _StatusCode,  *(int*)_pUser);
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{
        printf("Callid = %d-- nAccountId = %d --->state = %d, status code = %d------------>userdata = %d\n", _nCallId, _nAccountId, _State, _StatusCode,  *(int*)_pUser);
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
        int nid = -1;
        SipAccountConfig AccountConfig;
        AccountConfig.pUserName = "1001";
        AccountConfig.pPassWord = "d5YuLBNx";
        AccountConfig.pDomain = "180.97.147.174";
        AccountConfig.pUserData = (void *)user;
        AccountConfig.nMaxOngoingCall = 2;

        int ret = SipRegAccount(&AccountConfig, 2);

        while(1) {
                if (incommingcall) {
                        void *pLocalSdp;
                        CreateTmpSDP(&pLocalSdp);
                        SipAnswerCall(nCallid, 200, NULL, pLocalSdp);
                        break;
                }
        }
        SipDestroyInstance();
        sleep(4);
        return 0;
}
