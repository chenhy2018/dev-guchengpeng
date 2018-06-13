#include "sip.h"
#include <unistd.h>
#include <stdio.h>
#include "rtmp.h"
#include "main.h"

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
        printf("------------------------------------------------>state = %d, status code = %d\n", _State, _StatusCode);
        if ( _State == INV_STATE_CONFIRMED ) {
            DBG_LOG("INV_STATE_CONFIRMED\n");
            RTMPStat( RTMP_START );
        } else if ( _State == INV_STATE_DISCONNECTED ) {
            DBG_LOG("INV_STATE_DISCONNECTED\n");
            RTMPStat( RTMP_STOP );
        } else {
            DBG_LOG("other state\n");
        }
}

int main()
{
        SipCallBack cb;
        cb.OnIncomingCall  = &cbOnIncomingCall;
        cb.OnCallStateChange = &cbOnCallStateChange;
        cb.OnRegStatusChange = &cbOnRegStatusChange;

        Rtmp_Init();

        SipCreateInstance(&cb);
        sleep(2);
        int nid1 = SipAddNewAccount("1001", "1001", "123.59.204.198");
        SipRegAccount(nid1, 1);
        //sleep(20);
        //SipHangUp(nCallId1);
        //SipDestroyInstance();
        while(1) {
        }
        return 0;
}
