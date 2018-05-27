// Last Update:2018-05-27 17:18:34
/**
 * @file sdk_interface.c
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-25
 */

#include <string.h>
#include "sip.h"
#include "dbg.h"
#include "queue.h"
#include "sdk_interface.h"

#define MESSAGE_QUEUE_MAX 256

typedef struct {
    MessageQueue *pQueue;
    int fd;
} UA;

UA UaInstance;
UA *pUA = &UaInstance;

SipAnswerCode cbOnIncomingCall(int _nAccountId, int _nCallId, const char *_pFrom)
{
    Message message, *pMessage = &message;
    event_s *pEvent = (event_s *) malloc( sizeof(event_s) );

    DBG_LOG("incoming call From %s to %d\n", _pFrom, _nAccountId);

    memset( &message, 0, sizeof(Message) );
    memset( pEvent, 0, sizeof(event_s) );
    pMessage->nMessageID = EVENT_TYPE_INCOMING_CALL;
    if ( pEvent ) {
        pEvent->nAccountId = _nAccountId;
        pEvent->nCallId = _nCallId;
        if ( _pFrom )
            memcpy( pEvent->body.From, _pFrom, strlen(_pFrom) );
        
        pMessage->pMessage = pEvent;
    }
    if ( pMessage )
        SendMessage( pUA->pQueue, pMessage );
	return OK;
}

void cbOnRegStatusChange(int _nAccountId, SipAnswerCode _StatusCode)
{
    DBG_LOG("reg status = %d\n", _StatusCode);
}

void cbOnCallStateChange(int _nCallId, SipInviteState _State, SipAnswerCode _StatusCode)
{
    Message message, *pMessage = &message;
    event_s *pEvent = (event_s *) malloc( sizeof(event_s) );

    DBG_LOG("state = %d, status code = %d\n", _State, _StatusCode);

    memset( &message, 0, sizeof(Message) );
    memset( pEvent, 0, sizeof(event_s) );
    if ( _State == INV_STATE_CONFIRMED ) {
        pMessage->nMessageID = EVENT_TYPE_SESSION_ESTABLISHED;
        if ( pEvent ) {
            pEvent->nCallId = _nCallId;
        }
    } else if ( _State == INV_STATE_DISCONNECTED ) {
        pMessage->nMessageID = EVENT_TYPE_SESSION_ESTABLISHED;
        if ( pEvent ) {
            pEvent->nCallId = _nCallId;
        }
    } else {
    }
    if ( pEvent )
        pMessage->pMessage  = (void *)pEvent;

    if ( pMessage )
        SendMessage( pUA->pQueue, pMessage );
}

int CreateUA()
{
    SipCallBack cb;

    memset( pUA, 0, sizeof(UA) );

    if ( !pUA->pQueue ) {
        pUA->pQueue = CreateMessageQueue( MESSAGE_QUEUE_MAX );
        if ( !pUA->pQueue ) {
            DBG_ERROR("queue malloc fail\n");
            return -1;
        }
    }
    pUA->fd++;

    cb.OnIncomingCall  = &cbOnIncomingCall;
    cb.OnCallStateChange = &cbOnCallStateChange;
    cb.OnRegStatusChange = &cbOnRegStatusChange;
    SipCreateInstance(&cb);

    return pUA->fd;
}

int UA_Destroy()
{
    if ( pUA->pQueue ) {
        DestroyMessageQueue( &pUA->pQueue );
    }

    return 0;
}

int Register( const char* id, const char* host, const char* password, const int _bDeReg)
{
    int nid = 0;

    nid = SipAddNewAccount( id, password, host );
    SipRegAccount( nid, _bDeReg );

    return nid;
}

int MakeCall( int fd, int _nNid, const char* _pDestUri, const stream_s * _pStream )
{
    int CallId = 0;

    if ( !_pDestUri || !_pStream )
        return;

    // libsip need to tell CallId when cb function been called
    CallId = SipMakeNewCall( _nNid, _pDestUri );

    return CallId;
}

int PollEvents(  int* eventID, void* event, int nTimeOut)
{
    Message *pMessage = NULL;

    if (!eventID || !event ) {
        return RET_FAIL;
    }

    if ( nTimeOut ) {
        pMessage = ReceiveMessageTimeout( pUA->pQueue, nTimeOut );
    } else {
        pMessage = ReceiveMessage( pUA->pQueue );
    }

    if ( !pMessage ) {
        return RET_FAIL;
    }

    *eventID = pMessage->nMessageID;
    if ( pMessage->pMessage ) {
        memcpy( event, pMessage->pMessage, sizeof(event_s) );
        free( pMessage->pMessage );
    }

    return RET_SUCCESS;
}

int AnswerCall( int fd, int _nCallIndex )
{
}

int Reject( int fd, int _nCallIndex)
{
}

int HangupCall( int fd, int _nCallId )
{
    SipHangUp(_nCallId );

    return RET_SUCCESS;
}

int Report( int fd, const char* message, size_t length)
{
}

int SendPacket( int fd , int callIndex, int streamIndex, const char* buffer, size_t size)
{
}



