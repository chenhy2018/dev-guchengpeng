// Last Update:2018-05-25 21:30:58
/**
 * @file sdk_interface.c
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-25
 */

#include "sip.h"
#include "queue.h"

#define UA_MAX 64
#define MESSAGE_QUEUE_MAX 512

typedef struct 
{
    int fd;
    stream streams;
    void *instance;
} UA;

typedef struct {
    int current;
    UA UA_List[UA_MAX];
    MessageQueue *pQueue;
} UA_Manager_s;

UA_Manager_s g_UA_Manager;
g_UA_Manager *pManager = &g_UA_Manager;

SipAnswerCode cbOnIncomingCall(int _nAccountId, int _nCallId, const char *_pFrom)
{
    printf("-------------------->incoming call From %s to %d\n", _pFrom, _nAccountId);
	return OK ;
}

void cbOnRegStatusChange(int _nAccountId, SipAnswerCode _StatusCode)
{
    printf("------------------->reg status = %d\n", _StatusCode);
}

void cbOnCallStateChange(int _nCallId, SipInviteState _State, SipAnswerCode _StatusCode)
{
    printf("------------->state = %d, status code = %d\n", _State, _StatusCode);
}

int CreateUA( stream_s *stream )
{
    UA ua;
    SipCallBack cb;

    if (!stream)
        return;

    memset( pManager, 0, sizeof(UA_Manager_s) );
    ua.fd = pManager->current+1; 
    memcpy( ua.streams, *stream, sizeof(stream) );
    pManager->UA_List[pManager->current++] = ua;
    if ( !pManager->pQueue ) {
        pManager->pQueue = CreateMessageQueue( MESSAGE_QUEUE_MAX );
        if ( !pManager->pQueue ) {
            DBG_ERROR("queue malloc fail\n");
            return -1;
        }
    }

    cb.OnIncomingCall  = &cbOnIncomingCall;
    cb.OnCallStateChange = &cbOnCallStateChange;
    cb.OnRegStatusChange = &cbOnRegStatusChange;
    SipCreateInstance(&cb);

    return ua.fd;
}

int DestoryUA( int fd )
{
}

int FindUA( int fd, UA **ua )
{
    int i = 0;

    for ( i=0; i<pManager->current; i++ ) {
        if ( fd == pManager->UA_List[i].fd ) {
            ua = &pManager->UA_List[i];
            return RET_SUCESS;
        }
    }

    return RET_FAIL;
}

int Register( int fd, const char* id, const char* host, const char* password)
{
    int nid1 = SipAddNewAccount( id, password, host );

    return RET_SUCESS;
}

int MakeCall(const struct UA* ua, const char* id, const char* host)
{
}

int AnswerCall(const struct UA* ua, int callIndex)
{
}

int Reject(const struct UA* ua, int callIndex)
{
}

int HangupCall(const struct UA* ua, int callIndex)
{
}

int Report(struct UA* ua, const char* message, size_t length)
{
}

int SendPacket(const struct UA* ua, int callIndex, int streamIndex, const char* buffer, size_t size)
{
}

int PollEvents(const struct UA* ua, int* eventID, void* event)
{
}
