// Last Update:2018-05-31 15:34:58
/**
 * @file framework.c
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-25
 */
#include "queue.h"
#include "sdk_interface.h"
#include "sip.h"

SipAnswerCode cbOnIncomingCall(int _nAccountId, int _nCallId, const char *_pFrom)
{
    Message *pMessage = (Message *) malloc( sizeof(Message) );
    EventData *pData = (EventData *) malloc( sizeof(EventData) );

    DBG_LOG("incoming call From %s to %d\n", _pFrom, _nAccountId);

    if ( !pMessage || !pData ) {
        DBG_ERROR("malloc error\n");
        return;
    }

    memset( pMessage, 0, sizeof(Message) );
    memset( pData, 0, sizeof(pData) );
    pMessage->nMessageID = EVENT_TYPE_INCOMING_CALL;
    pData->nAccountId = _nAccountId;
    pData->nCallId = _nCallId;
    if ( _pFrom )
        memcpy( pData->body.callEvent.From, _pFrom, strlen(_pFrom) );

    pMessage->pMessage = pData;
    SendMessage( pUA->pQueue, pMessage );

	return OK;
}

void cbOnRegStatusChange(int _nAccountId, SipAnswerCode _StatusCode)
{
    DBG_LOG("reg status = %d\n", _StatusCode);
}

void cbOnCallStateChange(int _nCallId, IN const int _nAccountId, SipInviteState _State, SipAnswerCode _StatusCode)
{
    Message *pMessage = (Message *) malloc ( sizeof(Message) );
    EventData *pData = (EventData *) malloc( sizeof(EventData) );

    DBG_LOG("state = %d, status code = %d\n", _State, _StatusCode);

    if ( !pMessage || !pData ) {
        DBG_ERROR("malloc error\n");
        return;
    }

    memset( pMessage, 0, sizeof(Message) );
    memset( pData, 0, sizeof(pData) );
    if ( _State == INV_STATE_CONFIRMED ) {
        pMessage->nMessageID = EVENT_TYPE_SESSION_ESTABLISHED;
        pData->nCallId = _nCallId;
    } else if ( _State == INV_STATE_DISCONNECTED ) {
        pMessage->nMessageID = EVENT_TYPE_SESSION_ESTABLISHED;
        pEvent->nCallId = _nCallId;
    } else {
    }
    pMessage->pMessage  = (void *)pEvent;
    SendMessage( pUA->pQueue, pMessage );
}

