// Last Update:2018-05-31 18:15:34
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

UA UaList;

ErrorID LibraryInit()
{
    SipCallBack cb;

    memset( &UaHead, 0, sizeof(UA) );

    cb.OnIncomingCall  = &cbOnIncomingCall;
    cb.OnCallStateChange = &cbOnCallStateChange;
    cb.OnRegStatusChange = &cbOnRegStatusChange;
    SipCreateInstance(&cb);

    return RET_OK;
}

ErrorID Register( IN char* id, IN char* host, IN char* password, int _bDeReg, OUT int *_nAccountId )
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
    nAccountId = SipAddNewAccount( id, password, host );
    SipRegAccount( nAccountId, _bDeReg );
    *_nAccountId = nAccountId;

    return RET_OK;
}

ErrorID UnRegister( AccountId _nAccountId )
{
    struct list_head *pos, q;
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

ErrorID MakeCall( int _nAccountId, IN char* _pDestUri, OUT int *_pCallId )
{
    int nCallId = 0;

    if ( !_pDestUri || !_pCallId )
        return RET_PARAM_ERROR;

    // libsip need to tell CallId when cb function been called
    *_pCallId = SipMakeNewCall( _nAccountId, _pDestUri );

    return RET_OK;
}

ErrorID PollEvents( AccountID id, OUT int* eventID, OUT EventData** data int nTimeOut )
{
    Message *pMessage = NULL;
    static Message *pLastMessage = NULL;
    struct list_head *pos, q;
    UA *tmp = NULL;

    if (!eventID || !event ) {
        return RET_PARAM_ERROR;
    }

    list_for_each_safe(pos, q, &UaList.list){
        tmp = list_entry(pos, UA, list);
        if ( tmp->id == _nAccountId ) {
            break;
        }
    }

    if ( !tmp ) {
        DBG_ERROR("account id not exist, id = %d\n", id );
        return RET_ACCOUNT_NOT_EXIST;
    }

    // pLastMessage use to free last message
    if ( tmp->pLastMessage ) {
        if ( tmp->pLastMessage->stream.packet ) {
            free( tmp->pLastMessage->stream.packet );
        }
        free( tmp->pLastMessage );
    }

    if ( nTimeOut ) {
        pMessage = ReceiveMessageTimeout( pUA->pQueue, nTimeOut );
    } else {
        pMessage = ReceiveMessage( pUA->pQueue );
    }

    if ( !pMessage ) {
        return RET_RETRY;
    }

    *eventID = pMessage->nMessageID;
    if ( pMessage->pMessage ) {
        *data = (EventData *)pMessage->pMessage;
        // save the pointer of current message
        // so next time we received message
        // we can free the last one
        tmp->pLastMessage = pMessage;
    }

    return RET_OK;
}

ErrorID AnswerCall( AccountId id, int _nCallId )
{
    (void)id;

    SipAnswerCall( _nCallId, OK );

    return RET_OK;
}

ErrorID RejectCall( AccountId id, int _nCallId )
{
    (void)id;

    SipAnswerCall( _nCallIndex, BUSY_HERE );

    return RET_OK;
}

ErrorID HangupCall( AccountId id, int _nCallId )
{
    (void)id;

    SipHangUp( _nCallId );

    return RET_OK;
}

ErrorID SendPacket( AccountId id , int nCallId, int streamIndex, IN char* buffer, int size)
{
}

ErrorID AddCodec( AccountID _id, Codec _codecs[], int _size, int _nSamplerate, int _channels);
{
    UA *pUA = NULL;

    list_for_each_entry( pUA, &UaList.list, list ){
        if ( pUA->id == _id ) {
            memcpy( pUA->streamInfo.codecs, _codecs, _size );
            pUA->streamInfo.samplerate = _nSamplerate;
            pUA->channels = _channels;
            return RET_OK;
        }
    }
    
    DBG_ERROR("account not found, id = %d\n", _id );
    return RET_ACCOUNT_NOT_EXIST;
}

int Report( int fd, const char* message, size_t length)
{
}

