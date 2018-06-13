// Last Update:2018-06-04 14:08:15
/**
 * @file framework.c
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-25
 */
#include <string.h>
#include "queue.h"
#include "sdk_interface.h"
#include "sdk_local.h"
#include "sip.h"
#include "dbg.h"

SipAnswerCode cbOnIncomingCall(IN const int _nAccountId, IN const int _nCallId,
                               IN const char *_pFrom, IN const void *_pUser, IN const void *_pMedia)
{
    Message *pMessage = (Message *) malloc( sizeof(Message) );
    Event *pEvent = (Event *) malloc( sizeof(Event) );
    CallEvent *pCallEvent = NULL;
    const UA *pUA = _pUser;

    DBG_LOG("incoming call From %s to %d\n", _pFrom, _nAccountId);

    if ( !pMessage || !pEvent ) {
        DBG_ERROR("malloc error\n");
        return 0;
    }

    memset( pMessage, 0, sizeof(Message) );
    memset( pEvent, 0, sizeof(Event) );
    pMessage->nMessageID = EVENT_CALL;
    pCallEvent = &pEvent->body.callEvent;
    pCallEvent->callID = _nCallId;
    pCallEvent->status = CALL_STATUS_INCOMING;
    if ( _pFrom ) {
        pCallEvent->pFromAccount = (char *) malloc ( strlen(_pFrom) + 1);
        memset( pCallEvent->pFromAccount, 0, strlen(_pFrom) + 1 );
        memcpy( pCallEvent->pFromAccount, _pFrom, strlen(_pFrom) );
    }

    pMessage->pMessage = pEvent;
    if ( pUA )
        SendMessage( pUA->pQueue, pMessage );
    else {
        DBG_ERROR("pUA is NULL\n");
    }

	return OK;
}

void cbOnRegStatusChange( IN const int _nAccountId, IN const SipAnswerCode _regStatusCode, IN const void *_pUser )
{
    UA *pUA = ( UA *)_pUser;

    DBG_LOG("reg status = %d\n", _regStatusCode);
    if ( pUA ) {
        if ( _regStatusCode == OK ||
             _regStatusCode == UNAUTHORIZED ||
             _regStatusCode == REQUEST_TIMEOUT ) {
            pUA->regStatus = _regStatusCode;
            pthread_cond_signal( &pUA->registerCond );
        }
    } else {
        DBG_ERROR("pUser is NULL\n");
    }
}

void cbOnCallStateChange(IN const int _nCallId, IN const int _nAccountId, IN const SipInviteState _State,
                         IN const SipAnswerCode _StatusCode, IN const void *pUser, IN const void *pMedia)
{
    Message *pMessage = (Message *) malloc ( sizeof(Message) );
    Event *pEvent = (Event *) malloc( sizeof(Event) );
    CallEvent *pCallEvent = NULL;
    const UA *pUA = pUser;

    DBG_LOG("state = %d, status code = %d\n", _State, _StatusCode);

    if ( !pMessage || !pEvent ) {
        DBG_ERROR("malloc error\n");
        return;
    }

    memset( pMessage, 0, sizeof(Message) );
    memset( pEvent, 0, sizeof(Event) );
    pMessage->nMessageID = EVENT_CALL;
    pCallEvent = &pEvent->body.callEvent;
    pCallEvent->callID = _nCallId;
    if ( _State == INV_STATE_CONFIRMED ) {
        pCallEvent->status = CALL_STATUS_ESTABLISHED;
    } else if ( _State == INV_STATE_DISCONNECTED ) {
        pCallEvent->status = CALL_STATUS_HANGUP;
    } else {
    }
    pMessage->pMessage  = (void *)pEvent;
    if ( pUA )
        SendMessage( pUA->pQueue, pMessage );
    else {
        DBG_ERROR("pUA is NULL\n");
    }
}

