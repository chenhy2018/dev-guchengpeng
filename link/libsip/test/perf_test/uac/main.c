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
#define TEST1 0

void *pLocalSdp;
static int OtherCode[200];
static int OtherIdx;
#if TEST1
static char* PassWord[MAXACCOUNT] = {"37gP8trQ","MGCyV0Mk","kHvg5Fjp","usUznURU","l7lnQUsA","psJTzczQ","Y2GIQbi8","OdyDjgbm","sP7wa78K","nQHGpsrc","VXhadUcM","bO6WEs9K","dAvmei7H","14EolUgq","0aZWJ0Kh","3wJgSEBq","xEaZlhVi","8e3DPdX9","PDuLHx4A","yNuPC8SV","uZ4e78ea","tgR5wkyR","R3g9Pucs","HFQABlad","ZmFwiGCx","iu3WF1GA","606Y07X8","ApJyFm75","3I1k803f","i15TgIRD","ut6s30ZD","MbMSi9CC","NWbJV3Is","RK6pAa7p","fTCr67kf","DDR0LePj","o2xCLH8Z","aIoh14CG","FOKGoemD","GBuEofRe","sKSoW333","GRNTZAkI","hukgyF5D","kRLXE1qY","8f3dvSfb","9USd7VN2","IDkGJ2Fh","gMq7M8MG","1jZMWFCg","ofzJYuJc"};
static char* UserName[MAXACCOUNT] = {"1500","1501","1502","1503","1504","1505","1506","1507","1508","1509","1510","1511","1512","1513","1514","1515","1516","1517","1518","1519","1520","1521","1522","1523","1524","1525","1526","1527","1528","1529","1530","1531","1532","1533","1534","1535","1536","1537","1538","1539","1540","1541","1542","1543","1544","1545","1546","1547","1548","1549"};
#else
static char* PassWord[MAXACCOUNT] = {"bKvxaG1Y","xZUZsrAa","KKHWtvwi","CWTKB5a8","UKvONnz6","jhtT8tWA","Wu7Qo31t","bmUMJNVK","Z7eSSFom","sqEKaesr","Xkm1n9JW","CymVuDLc","IabDkgSM","la0wcacd","eQ9HTcEN","k0OVU2Y4","i1N7bCYy","bcFr4AtH","QVgw7cqN","LOH8KlOK","Xtvz2D6b","CrFjUghW","UnnTr9NR","cUgHYAfm","GCC0YVvM","UrWv9mPI","7XyPdqN3","LnthYAQD","Y6N1gAuZ","0yjoyP9q","J15i6v7O","Zf8Borz0","i7JAsSb9","pMtaqRwp","GF5DnZ8q","R9ugyeO3","wjLuaCtq","1jvozt7R","VKLRc6Wb","nXSj5259","Ra3VWjIZ","QetKQAtv","BCJr8MYp","RZOXeZ0j","TsAzPM1p","K4fIrpoN","mWL8K4DB","PPgKMoD1","3N8q4HEj","2Mk1dc6V"};

static char* UserName[MAXACCOUNT] = {"1600","1601","1602","1603","1604","1605","1606","1607","1608","1609","1610","1611","1612","1613","1614","1615","1616","1617","1618","1619","1620","1621","1622","1623","1624","1625","1626","1627","1628","1629","1630","1631","1632","1633","1634","1635","1636","1637","1638","1639","1640","1641","1642","1643","1644","1645","1646","1647","1648","1649"};
#endif

int CallState[MAXCALL];
int AccountState[MAXACCOUNT];

int RegSuccess = 0;
int Reg5XX = 0;
int Reg408 = 0;
int Reg401 = 0;

SipAnswerCode cbOnIncomingCall(int _nAccountId, const char *_pFrom, const void *_pUser, const void *_pMedia, int *_pCallId)
{
        //printf("----->incoming call From %s to %d--------------userdata = %d\n", _pFrom, _nAccountId, *(int*)_pUser);
        return OK ;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _StatusCode, const void *_pUser)
{
        if (_StatusCode != 200 && RegSuccess > 0) {
                AccountState[_nAccountId] = 0;
                RegSuccess--;
        }
        if (_StatusCode == UNAUTHORIZED)
                Reg401++;
        else if (_StatusCode == REQUEST_TIMEOUT)
                Reg408++;
        else if (_StatusCode >= INTERNAL_SERVER_ERROR && _StatusCode <= PRECONDITION_FAILURE)
                Reg5XX++;
        else if ((AccountState[_nAccountId] == 0) && (_StatusCode == 200)) {
                AccountState[_nAccountId] = 1;
                RegSuccess++;
        }
}

int SendInvCount  = 0;
int Inv404 = 0;
int Inv408 = 0;
int Inv407 = 0;
int Inv477 = 0;
int Inv481 = 0;
int InvOthers = 0;
int Inv180 = 0;
int Inv200 = 0;
int InvAck = 0;

int SendBYE = 0;
int BYE408 = 0;
int BYE407 = 0;
int BYE477 = 0;
int BYE481 = 0;
int BYEOthers = 0;
int BYE200 = 0;

int RunningCalls = 0;
pthread_mutex_t lock;

int a = 0, b = 0, c = 0 ,d = 0, e = 0;

void UpdateInvCount(SipAnswerCode _StatusCode) {
        switch (_StatusCode) {
        case NOT_FOUND:
                Inv404++;
                break;
        case REQUEST_TIMEOUT:
                Inv408++;
                break;
        case PROXY_AUTHENTICATION_REQUIRED:
                Inv407++;
                break;

        case CALL_TSX_DOES_NOT_EXIST:
                Inv481++;
                break;
        default:
		OtherCode[OtherIdx++] = _StatusCode;
                if (OtherIdx == 200)
                        OtherIdx = 0;
                InvOthers++;
                break;
        }
}
void UpdateByeCount(SipAnswerCode _StatusCode) {
        switch (_StatusCode) {
        case REQUEST_TIMEOUT:
                BYE408++;
                break;
        case PROXY_AUTHENTICATION_REQUIRED:
                BYE407++;
                break;

        case CALL_TSX_DOES_NOT_EXIST:
                BYE481++;
                break;
        case 200:
                BYE200++;
                break;
        default:
                BYEOthers++;
                break;
        }
}
void HandleDisconnect(int _nCallId, SipAnswerCode _StatusCode) {
        if (CallState[_nCallId] != INV_STATE_CONFIRMED && CallState[_nCallId] != 7) {
                UpdateInvCount(_StatusCode);
        }
        else if (CallState[_nCallId] == 7){
                RunningCalls--;
                UpdateByeCount(_StatusCode);
        }
        CallState[_nCallId] = -1;
}
void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State, const SipAnswerCode _StatusCode, const void *_pUser, const void *_pMedia)
{

        //        printf("Callid = %d-- nAccountId = %d --->state = %d, status code = %d------------>userdata = %d\n", _nCallId, _nAccountId, _State, _StatusCode,  *(int*)_pUser);
        pthread_mutex_lock(&lock);
        switch (_State) {
        case INV_STATE_NULL:
        case INV_STATE_INCOMING:
                break;
        case INV_STATE_CALLING:
                {
                        CallState[_nCallId] = INV_STATE_CALLING;
                        break;
                }
        case INV_STATE_EARLY:
                {
                        Inv180++;
                        CallState[_nCallId] = INV_STATE_EARLY;
                        break;
                }
        case INV_STATE_CONNECTING:
                {
                        Inv200++;
                        CallState[_nCallId] = INV_STATE_CONNECTING;
                        break;
                }
        case INV_STATE_CONFIRMED:
                {
                        InvAck++;
                        RunningCalls++;
                        CallState[_nCallId] = INV_STATE_CONFIRMED;
                        break;
                }
        case INV_STATE_DISCONNECTED:
                {
                        //                if (_StatusCode != 200)
                        //printf("Callid = %d-- nAccountId = %d --->state = %d, status code = %d------------>userdata = %d\n", _nCallId, _nAccountId, _State, _StatusCode,  *(int*)_pUser);

                        HandleDisconnect(_nCallId, _StatusCode);
                        break;
                }
        }
        pthread_mutex_unlock(&lock);
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
        pthread_mutex_lock(&lock);
        for (i = lastcall; i < MAXCALL; i++) {
                if (CallState[i] == -1) {
                        lastcall++;
                        pthread_mutex_unlock(&lock);
                        return i;
                }
        }
        pthread_mutex_unlock(&lock);
        lastcall = 0;
        return -1;

}
void* MakeCall(void* arg) {
        sleep(2);

        while(1) {
                int callId = FindCallId();
                int nAccount = rand() % (MAXACCOUNT -1 + 1 - 0) + 0;
                if (callId != -1 && AccountState[nAccount] != 0) {
                        CallState[callId] = 0;
                        SendInvCount++;
#if TEST1
                        SipMakeNewCall(nAccount, "<sip:1550@180.97.147.174;transport=tcp>", pLocalSdp, callId);
#else
                        SipMakeNewCall(nAccount, "<sip:1650@180.97.147.174;transport=tcp>", pLocalSdp, callId);
#endif
                }
                usleep(10000);
        }
        return NULL;
}

static int last;
int  RRID() {
        int i;

        pthread_mutex_lock(&lock);
        for (i = last; i < MAXCALL; i++) {
                if (CallState[i] == 5) {
                        last += 1;
                        pthread_mutex_unlock(&lock);
                        return i;
                }
        }
        pthread_mutex_unlock(&lock);
        last = 0;
        return -1;
}
void* HangUpCall(void* arg) {
        while(1) {
                int CallId = RRID();
                if (CallId != -1) {
                        CallState[CallId] = 7;
                        SipHangUp(CallId);
                        SendBYE++;
                }
                usleep(50000);
        }
}

void GetCurrentDateTime(char* buffer)
{
        time_t     now;
        struct tm *ts;
        char       buf[80];

        /* Get the current time */
        now = time(NULL);

        /* Format and print the time, "ddd yyyy-mm-dd hh:mm:ss zzz" */
        ts = localtime(&now);
        strftime(buffer, 80, "%m-%d %H:%M:%S", ts);
}
/*
void updateOnNcuser() {
        char StartTime[80];
        GetCurrentDateTime(StartTime);
        initscr();
        noecho();
        curs_set(FALSE);
        time_t start_t, end_t;
        time(&start_t);
        int lastInvAcc = 0;
        int x = 30, y = 30;
        while(1) {
                clear(); // Clear the screen of all
                // previously-printed characters
                char CurrentTime[80];
                GetCurrentDateTime(CurrentTime);

                time(&end_t);
                double elapsed = difftime(end_t, start_t);

                mvprintw(y + 0, x, "|=====================================================================================================+");
                mvprintw(y + 1, x, "|=====================================================================================================+");
                mvprintw(y + 2, x, "|=======================|Message |407/401 |  404   |408     | 477    | 481    | Others |==============+");
                mvprintw(y + 3, x, "|=======================|========|========|========|========|========|========|=======================+");
                mvprintw(y + 4, x, "| Register ------------>|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+", 50, Reg401, 0, Reg408, 0, 0, Reg5XX);
                mvprintw(y + 5, x, "| 200      <------------|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+", RegSuccess, 0, 0, 0, 0, 0, 0);
                mvprintw(y + 6, x, "| INVITE -------------->|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+", SendInvCount, Inv407, Inv404, Inv408, Inv477, Inv481, InvOthers);
                mvprintw(y + 7, x, "| 180    <--------------|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+", Inv180, 0, 0, 0, 0, 0, 0);
                mvprintw(y + 8, x, "| 200    <--------------|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+", Inv200, 0, 0, 0, 0, 0, 0);
                mvprintw(y + 9, x, "| ACK    -------------->|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+", InvAck, 0, 0, 0, 0, 0, 0);
                mvprintw(y + 10, x, "|===================================PAUSE    =========================================================+");
                mvprintw(y + 11, x, "| BYE    -------------->|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+", SendBYE, BYE407, 0, BYE408, BYE477, BYE481, BYEOthers);
                mvprintw(y + 12, x, "| 200    <--------------|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+", BYE200, 0, 0, 0, 0, 0);
                mvprintw(y + 13, x, "|==========================Statistics Info============================================================+");
                mvprintw(y + 14, x, "|                   Total-time(s) | Total-calls | running-calls | success-rate | call-rate(cps)       +");
                mvprintw(y + 15, x, "|                      %-11.2f|%13d|%-15d|%14.5f%|%-11.2f          +", elapsed, InvAck, RunningCalls, (1 - ((Inv407 + Inv404 + Inv408 + Inv477 + Inv481 + InvOthers) * 1.0) / SendInvCount) * 100, SendInvCount * 1.0 / elapsed);
                mvprintw(y + 16, x, "|=====================================================================================================+");
                mvprintw(y + 17, x, "|==StartTime = %s,    CurrentTime = %s========================================+", StartTime, CurrentTime);
                mvprintw(y + 18, x, "| Debug-----------------|%-8d|%8d|%-8d|%8d|%-8d|%8d|=======================+", a, b, c, d, e, last);
                mvprintw(y + 19, x,"|=====================================================================================================+");
                lastInvAcc = SendInvCount;
                for(int i = 0; i < 25; i++) {
                        for (int j = 0; j < 8; j++)
                                mvprintw(y + 20 + j, x+ 4 *i, "%3d,", OtherCode[j * 25 + i]);
                }

                refresh();

                usleep(DELAY); // Shorter delay between movements
        }

        endwin();
}
*/
void writeToFile() {
        char StartTime[80];
        GetCurrentDateTime(StartTime);
        time_t start_t, end_t;
        time(&start_t);
        int lastInvAcc = 0;
        FILE *pFile = NULL;
        pFile = fopen("./uac.log", "w");

        while(1) {
                char CurrentTime[80];
                GetCurrentDateTime(CurrentTime);

                time(&end_t);
                double elapsed = difftime(end_t, start_t);

                fprintf(pFile, "|=====================================================================================================+\n");
                fprintf(pFile, "|=====================================================================================================+\n");
                fprintf(pFile, "|=======================|Message |407/401 |  404   |408     | 477    | 481    | Others |==============+\n");
                fprintf(pFile, "|=======================|========|========|========|========|========|========|=======================+\n");
                fprintf(pFile, "| Register ------------>|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+\n", 50, Reg401, 0, Reg408, 0, 0, Reg5XX);
                fprintf(pFile, "| 200      <------------|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+\n", RegSuccess, 0, 0, 0, 0, 0, 0);
                fprintf(pFile, "| INVITE -------------->|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+\n", SendInvCount, Inv407, Inv404, Inv408, Inv477, Inv481, InvOthers);
                fprintf(pFile, "| 180    <--------------|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+\n", Inv180, 0, 0, 0, 0, 0, 0);
                fprintf(pFile, "| 200    <--------------|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+\n", Inv200, 0, 0, 0, 0, 0, 0);
                fprintf(pFile, "| ACK    -------------->|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+\n", InvAck, 0, 0, 0, 0, 0, 0);
                fprintf(pFile, "|===================================PAUSE    =========================================================+\n");
                fprintf(pFile, "| BYE    -------------->|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+\n", SendBYE, BYE407, 0, BYE408, BYE477, BYE481, BYEOthers);
                fprintf(pFile, "| 200    <--------------|%-8d|%8d|%-8d|%8d|%-8d|%8d|%-8d|==============+\n", BYE200, 0, 0, 0, 0, 0, 0);
                fprintf(pFile, "|==========================Statistics Info============================================================+\n");
                fprintf(pFile, "|                   Total-time(s) | Total-calls | running-calls | success-rate | call-rate(cps)       +\n");
                fprintf(pFile, "|                      %-11.2f|%13d|%-15d|%14.5f%|%-11.2f          +\n", elapsed, InvAck, RunningCalls, (1 - ((Inv407 + Inv404 + Inv408 + Inv477 + Inv481 + InvOthers) * 1.0) / SendInvCount) * 100, SendInvCount * 1.0 / elapsed);
                fprintf(pFile, "|=====================================================================================================+\n");
                fprintf(pFile, "|==StartTime = %s,    CurrentTime = %s========================================+\n", StartTime, CurrentTime);
                fprintf(pFile, "| Debug-----------------|%-8d|%8d|%-8d|%8d|%-8d|%8d|=======================+\n", a, b, c, d, e, last);
                fprintf(pFile,"|=====================================================================================================+\n");
                lastInvAcc = SendInvCount;
                sleep(1); // Shorter delay between movements
                fseek ( pFile , 0 , SEEK_SET);
        }
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
        registerAccount();
        initCallState();


        pthread_t tid1 ,tid2;
        pthread_create(&tid1, NULL, MakeCall, NULL);
        pthread_create(&tid2, NULL, HangUpCall, NULL);

        //updateOnNcuser();
        writeToFile();
        pthread_join(tid1, NULL);
        pthread_join(tid2, NULL);

}
