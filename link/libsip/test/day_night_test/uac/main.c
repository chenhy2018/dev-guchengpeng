#include "sip.h"

#include <sys/time.h>
//#include <ncurses.h>
#include <stdio.h>
#include <string.h>
#include <time.h>
#include <stdlib.h>
#include <unistd.h>
#include <pthread.h>

#define DELAY 3000
#define MAXACCOUNT 50
#define MAXCALL 50
void *pLocalSdp;

static char* PassWord[MAXACCOUNT] = {"37gP8trQ","MGCyV0Mk","kHvg5Fjp","usUznURU","l7lnQUsA","psJTzczQ","Y2GIQbi8","OdyDjgbm","sP7wa78K","nQHGpsrc","VXhadUcM","bO6WEs9K","dAvmei7H","14EolUgq","0aZWJ0Kh","3wJgSEBq","xEaZlhVi","8e3DPdX9","PDuLHx4A","yNuPC8SV","uZ4e78ea","tgR5wkyR","R3g9Pucs","HFQABlad","ZmFwiGCx","iu3WF1GA","606Y07X8","ApJyFm75","3I1k803f","i15TgIRD","ut6s30ZD","MbMSi9CC","NWbJV3Is","RK6pAa7p","fTCr67kf","DDR0LePj","o2xCLH8Z","aIoh14CG","FOKGoemD","GBuEofRe","sKSoW333","GRNTZAkI","hukgyF5D","kRLXE1qY","8f3dvSfb","9USd7VN2","IDkGJ2Fh","gMq7M8MG","1jZMWFCg","ofzJYuJc"};
static char* UserName[MAXACCOUNT] = {"1500","1501","1502","1503","1504","1505","1506","1507","1508","1509","1510","1511","1512","1513","1514","1515","1516","1517","1518","1519","1520","1521","1522","1523","1524","1525","1526","1527","1528","1529","1530","1531","1532","1533","1534","1535","1536","1537","1538","1539","1540","1541","1542","1543","1544","1545","1546","1547","1548","1549"};

int CallState[MAXCALL];
int AccountState[MAXACCOUNT];

static pthread_cond_t Cond = PTHREAD_COND_INITIALIZER;
static pthread_mutex_t lock = PTHREAD_MUTEX_INITIALIZER;

SipAnswerCode cbOnIncomingCall(int _nAccountId, const char *_pFrom, const void *_pUser, const void *_pMedia, int *_pCallId)
{
        //printf("----->incoming call From %s to %d--------------userdata = %d\n", _pFrom, _nAccountId, *(int*)_pUser);
        return OK ;
}

void GetCurrentDateTime(char* buffer)
{
        time_t     now;
        struct tm *ts;

        /* Get the current time */
        now = time(NULL);

        /* Format and print the time, "ddd yyyy-mm-dd hh:mm:ss zzz" */
        ts = localtime(&now);
        strftime(buffer, 80, "%m-%d %H:%M:%S", ts);
}
void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        if (_StatusCode != 200) {
                char buffer[80];
                GetCurrentDateTime(buffer);
                printf("============= %d Reg Faild at %s \n", _nAccountId, buffer);
        } else
                AccountState[_nAccountId] = 1;
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{
        CallState[_nCallId] = _State;
        if (INV_STATE_DISCONNECTED == _State) {
                CallState[_nCallId] = -1;
                char buffer[80];
                GetCurrentDateTime(buffer);
                pthread_cond_signal(&Cond);
                printf("=========AccontId %d  Callid = %d, Disconnected at %s cause of %d \n", _nAccountId, _nCallId, buffer, _StatusCode);
        }
}

void registerAccount() {
        int *user = (int *)malloc(sizeof(int));
        *user = 12345;
        SipAccountConfig AccountConfig;
        int i;
        for (i = 0; i < MAXACCOUNT; i++) {
                AccountConfig.pUserName = UserName[i];
                AccountConfig.pPassWord = PassWord[i];
                AccountConfig.pDomain = "180.97.147.174";
                AccountConfig.pUserData = (void *)user;
                AccountConfig.nMaxOngoingCall = 10;
                SipRegAccount(&AccountConfig, i);
                usleep(1000);
        }
}

void initCallState()
{
        int i;
        for (i = 0; i < MAXCALL; i++) {
                CallState[i] = -1;
         }

        for (i = 0; i < MAXACCOUNT; i++) {
                AccountState[i] = 0;
        }
}

static int lastcall = 0;
int FindCallId() {
        int i;
        for (i = lastcall; i < MAXCALL; i++) {
                if (CallState[i] == -1) {
                        lastcall++;
                        return i;
                }
        }
        lastcall = 0;
        return -1;

}
void* MakeCall(void* arg) {
        sleep(5);

	int i;
        while(1) {
                char buffer[80];
                GetCurrentDateTime(buffer);
		for (i = 0; i < MAXCALL; i++) {
			if (CallState[i] == -1) {
                        	printf("=========AccontId %d  Callid = %d, Make the Call at %s \n", i, i,  buffer);
                        	SipMakeNewCall(i, "<sip:1550@180.97.147.174;transport=tcp>", pLocalSdp, i);
                                sleep(1);
			}
                }
                pthread_cond_wait(&Cond, &lock);
        }
        return NULL;
}

int main(int argc, char *argv[]) {
        SipInstanceConfig Config;
        Config.Cb.OnIncomingCall  = &cbOnIncomingCall;
        Config.Cb.OnCallStateChange = &cbOnCallStateChange;
        Config.Cb.OnRegStatusChange = &cbOnRegStatusChange;
        Config.nMaxCall = 50;
        Config.nMaxAccount = 50;

        SipCreateInstance(&Config);
        SipSetLogLevel(0);
        CreateTmpSDP(&pLocalSdp);
        initCallState();
        registerAccount();
        pthread_t tid1 ,tid2;
        pthread_create(&tid1, NULL, MakeCall, NULL);

        pthread_join(tid1, NULL);
}
