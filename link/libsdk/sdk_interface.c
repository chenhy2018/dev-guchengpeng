// Last Update:2018-06-03 20:04:08
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
#include "mqtt.h"
#include "sdk_local.h"
#include "framework.h"
#include "list.h"

UA UaList;

ErrorID InitSDK( Media* _pMediaConfigs, int _nSize)
{
    SipCallBack cb;

    memset( &UaList, 0, sizeof(UA) );

    cb.OnIncomingCall  = &cbOnIncomingCall;
    cb.OnCallStateChange = &cbOnCallStateChange;
    cb.OnRegStatusChange = &cbOnRegStatusChange;
    SipCreateInstance(&cb);

    return RET_OK;
}

ErrorID UninitSDK()
{
}

AccountID Register( IN  char* id, IN char* password, IN char* sigHost,
                   IN char* mediaHost, IN char* imHost, int _nDeReg )
{
    int nAccountId = 0;
    UA *pUA = ( UA *) malloc ( sizeof(UA) );

    if ( !pUA ) {
        DBG_ERROR("malloc error\n");
        return RET_MEM_ERROR;
    }
    memset( pUA, 0, sizeof(UA) );
    pUA->pQueue = CreateMessageQueue( MESSAGE_QUEUE_MAX );
    if ( !pUA->pQueue ) {
        DBG_ERROR("queue malloc fail\n");
        return RET_MEM_ERROR;
    }
    list_add( &(pUA->list), &(UaList.list) );
    nAccountId = SipAddNewAccount( id, password, sigHost, (void *)pUA );
    SipRegAccount( nAccountId, _nDeReg );

    return nAccountId;
}

ErrorID UnRegister( AccountID _nAccountId )
{
    struct list_head *pos, *q;
    UA *tmp;

    list_for_each_safe(pos, q, &UaList.list){
        tmp = list_entry(pos, UA, list);
        if ( tmp->id == _nAccountId ) {
            list_del(pos);
            free(tmp);
            return RET_OK;
        }
    }

    return RET_ACCOUNT_NOT_EXIST;
}

ErrorID MakeCall(AccountID _nAccountId, const char* id, const char* _pDestUri, OUT int* _pCallId )
{
    int nCallId = 0;
    void *pMedia = NULL;

    if ( !_pDestUri || !_pCallId )
        return RET_PARAM_ERROR;

    *_pCallId = SipMakeNewCall( _nAccountId, _pDestUri, pMedia );

    return RET_OK;
}

ErrorID PollEvent(AccountID _nAccountID, EventType* _pType, Event* _pEvent, int _nTimeOut )
{
    Message *pMessage = NULL;
    struct list_head *pos, *q;
    UA *pUA = NULL;

    if (!_pType || !_pEvent ) {
        return RET_PARAM_ERROR;
    }

    list_for_each_safe(pos, q, &UaList.list){
        pUA = list_entry(pos, UA, list);
        if ( pUA->id == _nAccountID ) {
            break;
        }
    }

    if ( !pUA ) {
        DBG_ERROR("account id not exist, id = %d\n", _nAccountID );
        return RET_ACCOUNT_NOT_EXIST;
    }

    // pLastMessage use to free last message
    if ( pUA->pLastMessage ) {
        Event *pEvent = (Event *) pUA->pLastMessage->pMessage;
        if ( pEvent->body.dataEvent.data ) {
            free( pEvent->body.dataEvent.data );
            pEvent->body.dataEvent.data = NULL;
        }
        free( pEvent );
        pEvent = NULL;
    }

    if ( _nTimeOut ) {
        pMessage = ReceiveMessageTimeout( pUA->pQueue, _nTimeOut );
    } else {
        pMessage = ReceiveMessage( pUA->pQueue );
    }

    if ( !pMessage ) {
        return RET_RETRY;
    }

    *_pType = pMessage->nMessageID;
    if ( pMessage->pMessage ) {
        // save the pointer of current message
        // so next time we received message
        // we can free the last one
        pUA->pLastMessage = pMessage;
    }

    return RET_OK;
}

ErrorID AnswerCall( AccountID id, int _nCallId )
{
    int ret = 0;
    char *pReason = NULL;
    void *pMedia = NULL;

    (void)id;

    ret = SipAnswerCall( _nCallId, OK, pReason,  pMedia );

    return RET_OK;
}

ErrorID RejectCall( AccountID id, int _nCallId )
{
    int ret = 0;
    char *pReason = NULL;
    void *pMedia = NULL;

    (void)id;

    ret = SipAnswerCall( _nCallId, BUSY_HERE, pReason,  pMedia );

    return RET_OK;
}

ErrorID HangupCall( AccountID id, int _nCallId )
{
    (void)id;

    SipHangUp( _nCallId );

    return RET_OK;
}

ErrorID SendPacket(AccountID id, int callID, Stream streamID, const char* buffer, int size)
{
}

ErrorID Report(AccountID id, const char* topic, const char* message, int length)
{
}

ErrorID RegisterTopic(AccountID id, const char* topic);
ErrorID UnregisterTopic(AccountID id, const char* topic);

