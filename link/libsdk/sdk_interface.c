// Last Update:2018-06-04 14:18:25
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

UAManager gUAManager;
UAManager *pUAManager = &gUAManager;

ErrorID InitSDK( Media* _pMediaConfigs, int _nSize)
{
    SipCallBack cb;

    memset( pUAManager, 0, sizeof(UAManager) );
    
    memcpy( pUAManager->mediaConfigs, _pMediaConfigs, _nSize );
    cb.OnIncomingCall  = &cbOnIncomingCall;
    cb.OnCallStateChange = &cbOnCallStateChange;
    cb.OnRegStatusChange = &cbOnRegStatusChange;
    SipCreateInstance(&cb);
	pthread_mutex_init( &pUAManager->mutex, NULL );

    return RET_OK;
}

ErrorID UninitSDK()
{
    struct list_head *pos, *q;
    UA *pUA;

    list_for_each_safe(pos, q, &pUAManager->UAList.list){
        pUA = list_entry(pos, UA, list);
        pthread_cond_destroy( &pUA->registerCond );
        list_del(pos);
        free(pUA);
    }
	pthread_mutex_destroy( &pUAManager->mutex );

    return RET_OK;
}

AccountID Register( IN  char* _id, IN char* _password, IN char* _pSigHost,
                   IN char* _pMediaHost, IN char* _pImHost, int _nDeReg, int _nTimeOut )
{
    int nAccountId = 0;
    UA *pUA = ( UA *) malloc ( sizeof(UA) );
	struct timeval now;
	struct timespec waitTime;
    int nReason = 0;

    if ( !pUA ) {
        DBG_ERROR("malloc error\n");
        return RET_MEM_ERROR;
    }
    memset( pUA, 0, sizeof(UA) );
	pthread_cond_init( &pUA->registerCond, NULL);
    pUA->pQueue = CreateMessageQueue( MESSAGE_QUEUE_MAX );
    if ( !pUA->pQueue ) {
        DBG_ERROR("queue malloc fail\n");
        return RET_MEM_ERROR;
    }
	pthread_mutex_lock( &pUAManager->mutex );
    list_add( &(pUA->list), &(pUAManager->UAList.list) );
    nAccountId = SipAddNewAccount( _id, _password, _pSigHost, (void *)pUA );
    SipRegAccount( nAccountId, 0 );
    gettimeofday(&now, NULL);
    waitTime.tv_sec = now.tv_sec;
    waitTime.tv_nsec = now.tv_usec * 1000 + _nTimeOut * 1000 * 1000;
    nReason = pthread_cond_timedwait( &pUA->registerCond, &pUAManager->mutex, &waitTime );
    if (nReason == ETIMEDOUT) {
        DBG_ERROR("register time out\n");
        return RET_REGISTER_TIMEOUT;
    }
	pthread_mutex_unlock( &pUAManager->mutex );

    if ( pUA->regStatus == REQUEST_TIMEOUT ) {
        DBG_ERROR("register server return timeout\n");
        return RET_TIMEOUT_FROM_SERVER;
    }

    if ( pUA->regStatus == UNAUTHORIZED ) {
        DBG_ERROR("user unauthorized\n");
        return RET_USER_UNAUTHORIZED;
    }

    return nAccountId;
}

ErrorID UnRegister( AccountID _nAccountId )
{
    struct list_head *pos, *q;
    UA *pUA = NULL;

    SipRegAccount( _nAccountId, 1 );
    list_for_each_safe(pos, q, &pUAManager->UAList.list){
        pUA = list_entry(pos, UA, list);
        if ( pUA->id == _nAccountId ) {
            pthread_cond_destroy( &pUA->registerCond );
            list_del(pos);
            free(pUA);
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

    list_for_each_safe(pos, q, &pUAManager->UAList.list){
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

ErrorID RegisterTopic(AccountID id, const char* topic)
{
}

ErrorID UnregisterTopic(AccountID id, const char* topic)
{
}


