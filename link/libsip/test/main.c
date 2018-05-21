#include "sip.h"
#include <unistd.h>
#include <stdio.h>

SipAnswerCode cbOnIncomingCall(int _nAccountId, int _nCallId, const char *_pFrom)
{
        printf("------------------------------------------------------------------->incoming call From %s to %d\n", _pFrom, _nAccountId);
	return OK ;
}

void cbOnRegStatusChange(int _nAccountId, SipAnswerCode _StatusCode)
{
        printf("------------------------------------------------------------------->reg status = %d\n", _StatusCode);
}

void cbOnCallStateChange(int _nCallId, SipInviteState _State, SipAnswerCode _StatusCode)
{
        printf("------------------------------------------------------------------->state = %d, status code = %d\n", _State, _StatusCode);
}
int main()
{
        SipCallBack cb;
        cb.OnIncomingCall  = &cbOnIncomingCall;
        cb.OnCallStateChange = &cbOnCallStateChange;
        cb.OnRegStatusChange = &cbOnRegStatusChange;

        SipCreateInstance(&cb);
        sleep(2);
        int nid1 = SipAddNewAccount("1001", "1001", "192.168.56.102");
        int nid3 = SipAddNewAccount("1003", "1003", "192.168.56.102");
        int nid4 = SipAddNewAccount("1004", "1004", "192.168.56.102");
        int nid5 = SipAddNewAccount("1005", "1005", "192.168.56.102");
        int nid6 = SipAddNewAccount("1006", "1006", "192.168.56.102");
        int nid7 = SipAddNewAccount("1007", "1007", "192.168.56.102");
        int nid8 = SipAddNewAccount("1008", "1008", "192.168.56.102");
        int nid9 = SipAddNewAccount("1009", "1009", "192.168.56.102");
        int nid10 = SipAddNewAccount("1010", "1010", "192.168.56.102");

        SipRegAccount(nid1, 1);
        SipRegAccount(nid3, 1);
        SipRegAccount(nid4, 1);
        SipRegAccount(nid5, 1);
        SipRegAccount(nid6, 1);
        SipRegAccount(nid7, 1);
        SipRegAccount(nid8, 1);
        SipRegAccount(nid9, 1);
        SipRegAccount(nid10, 1);

        sleep(20);
        //int nCallId1 = SipMakeNewCall(nid3, "<sip:1004@192.168.56.102>");
        //sleep(20);
        //SipHangUp(nCallId1);
        //SipDestroyInstance();
        while(1) {
        }
        return 0;
}
