// Last Update:2018-06-08 18:41:37
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
#include "list.h"
#include "framework.h"
#include "uaMgr.h"

UAManager gUAManager;
UAManager *pUAManager = &gUAManager;

static UA* FindUA(UAManager* _pUAManager, AccountID _nAccountId, struct list_head **po)
{
        UA* pUA;
        struct list_head *q, *pos;
        DBG_LOG("FindUA in %p %p %p\n", &_pUAManager->UAList.list, pos, q);
        list_for_each_safe(pos, q, &_pUAManager->UAList.list) {
                DBG_LOG("FindUA pos %p\n", pos);
                pUA = list_entry(pos, UA, list);
                if (pUA->id == _nAccountId) {
                        *po = pos;
                        return pUA;
                }
        }
        return NULL;
}

void OnMessage(IN const void* _pInstance, IN int _nAccountId, IN const char* _pTopic, IN const char* _pMessage, IN size_t nLength)
{
        DBG_LOG("%p topic %s message %s nAccountId %d \n", _pInstance, _pTopic, _pMessage, _nAccountId);
}

void OnEvent(IN const void* _pInstance, IN int _nAccountId, IN int _nId,  IN const char* _pReason)
{       
        DBG_LOG("%p id %d, account id %d, reason  %s \n",_pInstance, _nAccountId, _nId, _pReason);
        // TODO call back to user.
        Message *pMessage = (Message *) malloc ( sizeof(Message) );
        Event *pEvent = (Event *) malloc( sizeof(Event) );
        MessageEvent *pMessageEvent = NULL;
        
        if ( !pMessage || !pEvent ) {
                DBG_ERROR("malloc error\n");
                return;
        }
        struct list_head *pos;

        UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
        if (pUA == NULL) {
                DBG_ERROR("PUA is NULL\n");
                return;
        }
        memset( pMessage, 0, sizeof(Message) );
        memset( pEvent, 0, sizeof(Event) );
        pMessage->nMessageID = EVENT_MESSAGE;
        pMessageEvent = &pEvent->body.messageEvent;
        pMessageEvent->status = _nId;
        pMessageEvent->message = _pReason;
        pMessage->pMessage  = (void *)pEvent;
        SendMessage(pUA->pQueue, pMessage);
}

void InitMqtt(struct MqttOptions* options, const char* _pId, const char* _pPassword, const char* _pImHost)
{
//Init option.
        options->pId = (char*)(_pId);
        options->bCleanSession = false;
        options->primaryUserInfo.nAuthenicatinMode = MQTT_AUTHENTICATION_USER;
        options->primaryUserInfo.pHostname = (char*)(_pImHost);
        //strcpy(options.bindaddress, "172.17.0.2");
        options->secondaryUserInfo.pHostname = NULL;
        //strcpy(options.secondBindaddress, "172.17.0.2`");
        options->primaryUserInfo.pUsername = "root";//(char*)(_pId);
        options->primaryUserInfo.pPassword = "root";//(char*)(_pPassword);
        options->secondaryUserInfo.pUsername = NULL;
        options->secondaryUserInfo.pPassword = NULL;
        options->secondaryUserInfo.nPort = 0;
        options->primaryUserInfo.nPort = 1883;
        options->primaryUserInfo.pCafile = NULL;
        options->primaryUserInfo.pCertfile = NULL;
        options->primaryUserInfo.pKeyfile = NULL;
        options->secondaryUserInfo.pCafile = NULL;
        options->secondaryUserInfo.pCertfile = NULL;
        options->secondaryUserInfo.pKeyfile = NULL;
        options->nKeepalive = 10;
        options->nQos = 0;
        options->bRetain = false;
        options->callbacks.OnMessage = &OnMessage;
        options->callbacks.OnEvent = &OnEvent;

}

static CodecType ConversionFormat(Codec _nCodec)
{
        switch (_nCodec) {
                case CODEC_H264:
                        return MEDIA_FORMAT_H264;
                case CODEC_G711A:
                        return MEDIA_FORMAT_PCMA;
                case CODEC_G711U:
                        return MEDIA_FORMAT_PCMU;
                default:
                        return MEDIA_FORMAT_H264;
        }
        return MEDIA_FORMAT_H264;
}


ErrorID InitSDK( Media* _pMediaConfigs, int _nSize)
{
       SipInstanceConfig config;
       pUAManager->videoConfigs.nCount = 0;
       pUAManager->audioConfigs.nCount = 0;
       for (int count = 0; count < _nSize; ++count) {
               if (_pMediaConfigs[count].streamType == STREAM_VIDEO) {
                       pUAManager->videoConfigs.configs[pUAManager->videoConfigs.nCount].streamType = RTP_STREAM_VIDEO;
                       pUAManager->videoConfigs.configs[pUAManager->videoConfigs.nCount].codecType = ConversionFormat(_pMediaConfigs[count].codecType);
                       pUAManager->videoConfigs.configs[pUAManager->videoConfigs.nCount].nSampleOrClockRate = _pMediaConfigs[count].sampleRate;
                       pUAManager->videoConfigs.configs[pUAManager->videoConfigs.nCount].nChannel = _pMediaConfigs[count].channels;
                       ++pUAManager->videoConfigs.nCount;
               }
               else if (_pMediaConfigs[count].streamType == STREAM_AUDIO) {
                       pUAManager->audioConfigs.configs[pUAManager->audioConfigs.nCount].streamType = RTP_STREAM_AUDIO;
                       pUAManager->audioConfigs.configs[pUAManager->audioConfigs.nCount].codecType = ConversionFormat(_pMediaConfigs[count].codecType);
                       pUAManager->audioConfigs.configs[pUAManager->audioConfigs.nCount].nSampleOrClockRate = _pMediaConfigs[count].sampleRate;
                       pUAManager->audioConfigs.configs[pUAManager->audioConfigs.nCount].nChannel = _pMediaConfigs[count].channels;
                       ++pUAManager->audioConfigs.nCount;
               }
        }
        config.Cb.OnIncomingCall  = &cbOnIncomingCall;
        config.Cb.OnCallStateChange = &cbOnCallStateChange;
        config.Cb.OnRegStatusChange = &cbOnRegStatusChange;
        config.nMaxCall = 10;
        config.nMaxAccount = 10;
        // debug code.
        SipSetLogLevel(1);
        SipCreateInstance(&config);
        INIT_LIST_HEAD(&pUAManager->UAList.list);
        pUAManager->bInitSdk = true;
        return RET_OK;
}

ErrorID UninitSDK()
{
        struct list_head *pos, *q;
        UA *pUA;
        if (!pUAManager->bInitSdk) {
                DBG_ERROR("not init sdk\n");
                return RET_INIT_ERROR;
        }
        list_for_each_safe(pos, q, &pUAManager->UAList.list){
                pUA = list_entry(pos, UA, list);
                list_del(pos);
                UAUnRegister(pUA);
        }
        pUAManager->bInitSdk = false;
        memset(&pUAManager->videoConfigs, 0, sizeof(MediaConfig));
        memset(&pUAManager->audioConfigs, 0, sizeof(MediaConfig));

        return RET_OK;
}

AccountID Register(const char* _id, const char* _password, const char* _pSigHost,
                   const char* _pMediaHost, const char* _pImHost)
{
    int nAccountId = 0;
    struct MqttOptions options;
    InitMqtt(&options, _id, _password, _pImHost);
    UA *pUA = UARegister(_id, _password, _pSigHost, _pMediaHost, &options, &pUAManager->videoConfigs, &pUAManager->audioConfigs);
    int nReason = 0;

    if (!pUAManager->bInitSdk) {
        DBG_ERROR("not init sdk\n");
        return RET_INIT_ERROR;
    }
    if (pUA == NULL) {
        DBG_ERROR("malloc error\n");
        return RET_MEM_ERROR;
    }
    list_add(&(pUA->list), &(pUAManager->UAList.list));
    return pUA->id;
}

ErrorID UnRegister(AccountID _nAccountId)
{
    struct list_head *pos;
    UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
    if (pUA != NULL) {
            list_del(pos);
            UAUnRegister(pUA);
            return RET_OK;
    }

    return RET_ACCOUNT_NOT_EXIST;
}

ErrorID MakeCall(AccountID _nAccountId, const char* id, const char* _pDestUri, OUT int* _pCallId)
{
    struct list_head *pos;
    if ( !_pDestUri || !_pCallId )
        return RET_PARAM_ERROR;

    UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
    if (pUA != NULL) {
            return UAMakeCall(pUA, id, _pDestUri, _pCallId);
    }

    return RET_ACCOUNT_NOT_EXIST;
}

ErrorID PollEvent(AccountID _nAccountID, EventType* _pType, Event** _pEvent, int _nTimeOut )
{
    Message *pMessage = NULL;
    struct list_head *pos, *q;
    UA *pUA = NULL;

    if (!_pType || !_pEvent ) {
        return RET_PARAM_ERROR;
    }


    pUA = FindUA(pUAManager, _nAccountID, &pos);
    if (pUA == NULL) {
            DBG_ERROR( "RET_ACCOUNT_NOT_EXIST\n");
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
    DBG_LOG("wait for event, pUA = 0x%x\n", pUA );
    if (_nTimeOut) {
        pMessage = ReceiveMessageTimeout( pUA->pQueue, _nTimeOut );
    } else {
        pMessage = ReceiveMessage( pUA->pQueue );
    }

    DBG_LOG("[ LIBSDK ]get one event\n");
    if (!pMessage) {
        return RET_RETRY;
    }

    *_pType = pMessage->nMessageID;
    if ( pMessage->pMessage ) {
        // save the pointer of current message
        // so next time we received message
        // we can free the last one
        pUA->pLastMessage = pMessage;
    }
    
    *_pEvent = (Event *)pMessage->pMessage;

#if 0
    if (UAPollEvent(pUA, _pType, _pEvent, _nTimeOut) == RET_CALL_NOT_EXIST) {
        fprintf(stderr, "Call is not exist, poll next event\n");
        return PollEvent(_nAccountID, _pType,  _pEvent, _nTimeOut);
    }
#endif
    return RET_OK;
}

ErrorID AnswerCall(AccountID id, int _nCallId)
{
    struct list_head *pos;
    
    UA *pUA = FindUA(pUAManager, id, &pos);
    if (pUA != NULL) {
            return UAAnswerCall(pUA, _nCallId);
    }

    return RET_ACCOUNT_NOT_EXIST;
}

ErrorID RejectCall( AccountID id, int _nCallId )
{
    struct list_head *pos;

    UA *pUA = FindUA(pUAManager, id, &pos);
    if (pUA != NULL) {
            return UAAnswerCall(pUA, _nCallId);
    }

    return RET_ACCOUNT_NOT_EXIST;
}

ErrorID HangupCall( AccountID id, int _nCallId )
{
    struct list_head *pos;

    UA *pUA = FindUA(pUAManager, id, &pos);
    if (pUA != NULL) {
            return UAHangupCall(pUA, _nCallId);
    }

    return RET_ACCOUNT_NOT_EXIST;
}

ErrorID SendPacket(AccountID id, int _nCallId, Stream streamID, const uint8_t* buffer, int size, int64_t nTimestamp)
{
    struct list_head *pos;

    UA *pUA = FindUA(pUAManager, id, &pos);
    if (pUA != NULL) {
            return UASendPacket(pUA, _nCallId, streamID, buffer, size, nTimestamp);
    }

    return RET_ACCOUNT_NOT_EXIST;
}

ErrorID Report(AccountID id, const char* message, int length)
{
    struct list_head *pos;

    UA *pUA = FindUA(pUAManager, id, &pos);
    if (pUA != NULL) {
            return UAReport(pUA, message, length);
    }

    return RET_ACCOUNT_NOT_EXIST;
}

SipAnswerCode cbOnIncomingCall(const const int _nAccountId, const int _nCallId,
                               const const char *_pFrom, const void *_pUser, IN const void *_pMedia)
{   
    Message *pMessage = (Message *) malloc( sizeof(Message) );
    Event *pEvent = (Event *) malloc( sizeof(Event) );
    CallEvent *pCallEvent = NULL;
    const UA *_pUA = _pUser;
    struct list_head *pos;
    
    UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
    if (pUA == NULL) {
            return DOES_NOT_EXIST_ANYWHERE;
    }
    
    DBG_LOG("incoming call From %s to %d\n", _pFrom, _nAccountId);

    UAOnIncomingCall(pUA, _nCallId, _pFrom, _pMedia);
  
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

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _regStatusCode, const void *_pUser )
{
    Message *pMessage = (Message *) malloc( sizeof(Message) );
    Event *pEvent = (Event *) malloc( sizeof(Event) );
    CallEvent *pCallEvent = NULL;
    UA *_pUA = ( UA *)_pUser;
    struct list_head *pos;

    UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
    if (pUA == NULL) {
            DBG_ERROR("pUser is NULL %p\n", _pUA);
            return;
    }

    DBG_VAL(_nAccountId);
    DBG_LOG("pUA address is 0x%x, _regStatusCode = %d\n", pUA, _regStatusCode );
    memset( pMessage, 0, sizeof(Message) );
    memset( pEvent, 0, sizeof(Event) );
    pMessage->nMessageID = EVENT_CALL;
    pCallEvent = &pEvent->body.callEvent;
    pCallEvent->callID = 0;
    if ( _regStatusCode == OK ) {
        pCallEvent->status = CALL_STATUS_REGISTERED;
    } else {
        pCallEvent->status = CALL_STATUS_REGISTER_FAIL;
    }
    pCallEvent->pFromAccount = NULL;
    pMessage->pMessage = pEvent;
    if ( pUA ) {
        DBG_LOG("[ LIBSDK ] SendMessage\n");
        SendMessage( pUA->pQueue, pMessage );
    } else {
        DBG_ERROR("pUA is NULL\n");
        return;
    }

    DBG_LOG("reg status = %d\n", _regStatusCode);
    UAOnRegStatusChange(pUA, _regStatusCode);
    if ( pUA ) {
        if ( _regStatusCode == OK ||
             _regStatusCode == UNAUTHORIZED ||
             _regStatusCode == REQUEST_TIMEOUT ) {
            pUA->regStatus = _regStatusCode;
        }
    }
}

void cbOnCallStateChange(const int _nCallId, const int _nAccountId, const SipInviteState _State,
                         const SipAnswerCode _StatusCode, const void *pUser, const void *pMedia)
{
    Message *pMessage = (Message *) malloc ( sizeof(Message) );
    Event *pEvent = (Event *) malloc( sizeof(Event) );
    CallEvent *pCallEvent = NULL;
    const UA *_pUA = pUser;
    struct list_head *pos;

    DBG_LOG("state = %d, status code = %d\n", _State, _StatusCode);

    if ( !pMessage || !pEvent ) {
            DBG_ERROR("malloc error\n");
            return;
    }

    UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
    if (pUA == NULL) {
            DBG_ERROR("pUser is NULL\n");
            return;
    }

    memset( pMessage, 0, sizeof(Message) );
    memset( pEvent, 0, sizeof(Event) );
    pMessage->nMessageID = EVENT_CALL;
    pCallEvent = &pEvent->body.callEvent;
    pCallEvent->callID = _nCallId;
    DBG_ERROR("pUser is NULL\n");
    if ( _State == INV_STATE_CONFIRMED ) {
            pCallEvent->status = CALL_STATUS_ESTABLISHED;
    } else if ( _State == INV_STATE_DISCONNECTED ) {
            pCallEvent->status = CALL_STATUS_HANGUP;
    } else {
    }
    DBG_ERROR("pUser is NULL\n");
    pMessage->pMessage  = (void *)pEvent;
    SendMessage(pUA->pQueue, pMessage);
    UAOnCallStateChange(pUA, _nCallId, _State, _StatusCode, pMedia);
    DBG_LOG("cbOnCallStateChange end\n");
}
