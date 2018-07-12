// Last Update:2018-06-12 19:14:20
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
#include "uaMgr.h"
#include "callMgr.h"
#include <sys/types.h>

UAManager gUAManager;
UAManager *pUAManager = &gUAManager;
static int nSdkCallId = 0;

static UA* FindUA(UAManager* _pUAManager, AccountID _nAccountId, struct list_head **po)
{
        UA* pUA;
        struct list_head *q, *pos;
        //DBG_LOG("FindUA in %p %p %p AccountID %d \n", &_pUAManager->UAList.list, pos, q, _nAccountId);

        list_for_each_safe(pos, q, &_pUAManager->UAList.list) {
                pUA = list_entry(pos, UA, list);
                //DBG_LOG("FindUA pos %p id %d\n", pos, pUA->id);
                if (pUA->id == _nAccountId) {
                        *po = pos;
                        return pUA;
                }
        }
        return NULL;
}

static Codec ConversionFormatToUser(CodecType _nCodec)
{
        switch (_nCodec) {
                case MEDIA_FORMAT_H264:
                        return CODEC_H264;
                case MEDIA_FORMAT_PCMA:
                        return CODEC_G711A;
                case MEDIA_FORMAT_PCMU:
                        return CODEC_G711U;
                default:
                        return MEDIA_FORMAT_H264;
        }
        return MEDIA_FORMAT_H264;
}

void OnIncomingCall(const const int _nAccountId, const int _nCallId, const const char *_pFrom)
{
        pthread_mutex_lock(&pUAManager->mutex);
        Message *pMessage = (Message *) malloc( sizeof(Message) );
        Event *pEvent = (Event *) malloc( sizeof(Event) );
        CallEvent *pCallEvent = NULL;
        if ( !pMessage || !pEvent ) {
                if (pMessage) free(pMessage);
                if (pEvent) free(pEvent);
                DBG_ERROR("malloc error\n");
                pthread_mutex_unlock(&pUAManager->mutex);
                return;
        }
        struct list_head *pos;
        UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
        if (pUA == NULL) {
                free(pMessage);
                free(pEvent);
                pthread_mutex_unlock(&pUAManager->mutex);
                return;
        }

        DBG_LOG("incoming call From %s to %d  _nCallId %d\n", _pFrom, _nAccountId, _nCallId);

        memset( pMessage, 0, sizeof(Message) );
        memset( pEvent, 0, sizeof(Event) );
        pMessage->nMessageID = EVENT_CALL;
        pCallEvent = &pEvent->body.callEvent;
        pCallEvent->callID = _nCallId;
        pCallEvent->status = CALL_STATUS_INCOMING;
        if (_pFrom && strlen(_pFrom) > 0) {
                char* pFromAccount = (char *) malloc (strlen(_pFrom) + 1);
                memset(pFromAccount, 0, strlen(_pFrom) + 1);
                memcpy(pFromAccount, _pFrom, strlen(_pFrom));
                pCallEvent->context = pFromAccount;
        }

        pMessage->pMessage = pEvent;
        if (pUA) {
                SendMessage(pUA->pQueue, pMessage);
        } else {
                DBG_ERROR("pUA is NULL\n");
        }
        DBG_LOG("incoming call end\n");
        pthread_mutex_unlock(&pUAManager->mutex);
        return;
}

// Todo send to message queue.
static void OnRxRtp(void *_pUserData, CallbackType _type, void *_pCbData)
{
        //pthread_mutex_lock(&pUAManager->mutex);
        Call* pCall = (Call*)(_pUserData);
        if (pCall == NULL) {
                //pthread_mutex_unlock(&pUAManager->mutex);
                DBG_ERROR("OnRxRtp _pUserData is invaild******\n");
                return;
        }
        pthread_mutex_lock(&pUAManager->mutex);
        Message *pMessage = (Message *) malloc (sizeof(Message));
        Event *pEvent = (Event *) malloc(sizeof(Event));
        if ( !pMessage || !pEvent ) {
                if (pMessage) free(pMessage);
                if (pEvent) free(pEvent);
                DBG_ERROR("OnRxRtp malloc error***************\n");
                return;
        }
        struct list_head *pos;
        
        UA *pUA = FindUA(pUAManager, pCall->nAccountId, &pos);
        if (pUA == NULL) {
                DBG_ERROR("UA is NULL\n");
                free(pMessage);
                free(pEvent);
                pthread_mutex_unlock(&pUAManager->mutex);
                return;
        }

        switch (_type) {
                case CALLBACK_ICE: {
                        IceNegInfo *pInfo = (IceNegInfo *)_pCbData;
                        switch (pInfo->state) {
                                 case ICE_STATE_GATHERING_OK:
                                 {
                                         free(pMessage);
                                         free(pEvent);
                                         pthread_mutex_unlock(&pUAManager->mutex);
                                         DBG_LOG("=****=========>callback_ice: state: %d callStatus %d\n", pInfo->state, pCall->callStatus);
                                         pCall->pLocal = (pjmedia_sdp_session *)(pInfo->pData);
                                         if (pCall->callStatus == INV_STATE_EARLY || pCall->callStatus == INV_STATE_INCOMING) {
                                                 OnIncomingCall(pCall->nAccountId,pCall->id, pCall->from);
                                         }
                                         else if (pCall->callStatus == INV_STATE_CALLING) {
                                                 CALLMakeNewCall(pCall);
                                         }
                                         return;
                                 }

                                 case  ICE_STATE_NEGOTIATION_OK:
                                 {
                                         DBG_LOG("==========>callback_ice: state: %d\n", pInfo->state);
                                         CallEvent* event = &pEvent->body.callEvent;
                                         pMessage->nMessageID = EVENT_CALL;
                                         if (pInfo->nCount < 1) {
                                                 pCall->error = REASON_CALL_INVAILD_MEDIA_INFO;
                                                 SipHangUp(pCall->id);
                                                 free(pMessage);
                                                 free(pEvent);
                                                 pthread_mutex_unlock(&pUAManager->mutex);
                                                 return;
                                         }
                                         MediaInfo *pMedia = (MediaInfo *) malloc (sizeof(MediaInfo));
                                         for (int i = 0; i < pInfo->nCount; i++) {
                                                 DBG_LOG(" codec type: %d\n", pInfo->configs[i]->codecType);
                                                 pMedia->media[i].codecType = ConversionFormatToUser(pInfo->configs[i]->codecType);
                                                 if (pInfo->configs[i]->streamType == RTP_STREAM_VIDEO) {
                                                         pMedia->media[i].streamType = STREAM_VIDEO;
                                                 } else {
                                                         pMedia->media[i].streamType = STREAM_AUDIO;
                                                 }
                                                 pMedia->media[i].sampleRate = pInfo->configs[i]->nSampleOrClockRate;
                                                 pMedia->media[i].channels = pInfo->configs[i]->nChannel;
                                         }
                                         pMedia->nCount = pInfo->nCount;
                                         memcpy(&pCall->mediaInfo, event, sizeof(MediaInfo));
                                         event->callID = pCall->id;
                                         event->status = CALL_STATUS_ESTABLISHED;
                                         event->context = pMedia;
                                         event->reasonCode = REASON_OK;
                                         break;
                                 }
                                 case ICE_STATE_NEGOTIATION_FAIL:
                                 {
                                         DBG_ERROR("==========>callback_ice: state: %d\n", pInfo->state);
                                         pCall->error = REASON_CALL_NEGOTIATION_FAIL;
                                         SipHangUp(pCall->id);
                                         free(pMessage);
                                         free(pEvent);
                                         pthread_mutex_unlock(&pUAManager->mutex);
                                         return;
                                 }
                                 default :
                                 {
                                         CallEvent* event = &pEvent->body.callEvent;
                                         pMessage->nMessageID = EVENT_CALL;
                                         event->callID = pCall->id;
                                         event->status = CALL_STATUS_ERROR;
                                         UADeleteCall(pUA, pCall->id);
                                         event->context = NULL;
                                         event->reasonCode = REASON_CALL_INVAILD_SDP;
                                         DBG_ERROR("==========>callback_ice: state: %d\n", pInfo->state);
                                 }
                        }
                }
                        break;
                case CALLBACK_RTP:{
                        pMessage->nMessageID = EVENT_DATA;
                        RtpPacket *pPkt = (RtpPacket *)_pCbData;
                        pj_ssize_t nLen = pPkt->nDataLen;
                        DataEvent* event = &pEvent->body.dataEvent;
                        if (pPkt->type == RTP_STREAM_AUDIO && nLen == 160) {
                                event->stream = STREAM_AUDIO;
                        } else if (pPkt->type == RTP_STREAM_VIDEO) {
                                event->stream = STREAM_VIDEO;
                        }
                        //DBG_LOG("==========>callback_rtp nTimestamp %lld\n", pPkt->nTimestamp);
                        event->pts = pPkt->nTimestamp;
                        event->callID = pCall->id;
                        event->size = nLen;
                        event->codec = ConversionFormatToUser(pPkt->format);
                        event->data = (uint8_t*)malloc(nLen);
                        if (event->data == NULL) {
                                DBG_ERROR("OnRxRtp data malloc error***************\n");
                                pthread_mutex_unlock(&pUAManager->mutex);
                                return;
                        }
                        memcpy(event->data, pPkt->pData, nLen);
                }
                        break;
                case CALLBACK_RTCP:
                        DBG_LOG("==========>callback_rtcp\n");
                        free(pMessage);
                        free(pEvent);
                        pthread_mutex_unlock(&pUAManager->mutex);
                        return;
                case CALLBACK_SEND_RESULT:
                        //DBG_LOG("==========>callback_send result\n");
                        free(pMessage);
                        free(pEvent);
                        pthread_mutex_unlock(&pUAManager->mutex);
                        return;
        }
        pMessage->pMessage  = (void *)pEvent;
        SendMessage(pUA->pQueue, pMessage);
        pthread_mutex_unlock(&pUAManager->mutex);
}

void OnMessage(const void* _pInstance, int _nAccountId, const char* _pTopic, const char* _pMessage, size_t nLength)
{
        //DBG_LOG("%p topic %s message %s nAccountId %d length %d\n", _pInstance, _pTopic, _pMessage, _nAccountId, nLength);
        Message *pMessage = (Message *) malloc ( sizeof(Message) );
        Event *pEvent = (Event *) malloc( sizeof(Event) );
        MessageEvent *pMessageEvent = NULL;
        if (!pMessage || !pEvent) {
                DBG_ERROR("malloc error\n");
                if (pMessage) free(pMessage);
                if (pEvent) free(pEvent);
                return;
        }
        struct list_head *pos;
        pthread_mutex_lock(&pUAManager->mutex);
        UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
        if (pUA == NULL) {
                DBG_ERROR("UA is NULL\n");
                if (pMessage) free(pMessage);
                if (pEvent) free(pEvent);
                pthread_mutex_unlock(&pUAManager->mutex);
                return;
        }
        memset( pMessage, 0, sizeof(Message) );
        memset( pEvent, 0, sizeof(Event) );
        pMessage->nMessageID = EVENT_MESSAGE;
        pMessageEvent = &pEvent->body.messageEvent;
        pMessageEvent->status = MESSAGE_STATUS_DATA;
        char *message = (char *) malloc (nLength) ;
        strncpy(message, _pMessage, nLength);
        message[nLength] = 0;
        pMessageEvent->message = message;
        char *topic = (char *) malloc (strlen(_pTopic) + 1);
        strncpy(topic, _pTopic, strlen(_pTopic));
        topic[strlen(_pTopic)] = 0;
        pMessageEvent->topic = topic;
        //DBG_LOG("message %p  %s\n", pMessageEvent->message, pMessageEvent->message);
        pMessage->pMessage  = (void *)pEvent;
        SendMessage(pUA->pQueue, pMessage);
        pthread_mutex_unlock(&pUAManager->mutex);
}

void OnEvent(IN const void* _pInstance, IN int _nAccountId, IN int _nId,  IN const char* _pReason)
{       
        DBG_LOG("%p account id %d, id %d, reason  %s \n",_pInstance, _nAccountId, _nId, _pReason);
        Message *pMessage = (Message *) malloc ( sizeof(Message) );
        Event *pEvent = (Event *) malloc( sizeof(Event) );
        MessageEvent *pMessageEvent = NULL;
        
        if (!pMessage || !pEvent) {
                DBG_ERROR("malloc error\n");
                if (pMessage) free(pMessage);
                if (pEvent) free(pEvent);
                return;
        }
        struct list_head *pos;

        pthread_mutex_lock(&pUAManager->mutex);

        UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
        if (pUA == NULL) {
                DBG_ERROR("UA is NULL\n");
                if (pMessage) free(pMessage);
                if (pEvent) free(pEvent);
                pthread_mutex_unlock(&pUAManager->mutex);
                return;
        }

        memset( pMessage, 0, sizeof(Message) );
        memset( pEvent, 0, sizeof(Event) );
        pMessage->nMessageID = EVENT_MESSAGE;
        pMessageEvent = &pEvent->body.messageEvent;

        if (_nId == MQTT_CONNECT_SUCCESS) {
                pMessageEvent->status = MESSAGE_STATUS_CONNECT;
        } else if (_nId == MQTT_DISCONNECT_SUCCESS) {
                pMessageEvent->status = MESSAGE_STATUS_DISCONNECT;
        } else {
                pMessageEvent->status = MESSAGE_STATUS_ERROR;
        }

        char *message = (char *) malloc (strlen(_pReason) + 1) ;
        strncpy(message, _pReason, strlen(_pReason));
        message[strlen(_pReason)] = 0;
        pMessageEvent->message = message;//_pReason;
        pMessageEvent->topic = NULL;
        DBG_LOG("message %p  %s\n", pMessageEvent->message, pMessageEvent->message);
        pMessage->pMessage  = (void *)pEvent;
        SendMessage(pUA->pQueue, pMessage);
        pthread_mutex_unlock(&pUAManager->mutex);
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
        options->primaryUserInfo.pUsername = (char*)(_pId);
        options->primaryUserInfo.pPassword = (char*)(_pPassword);
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

       if ( pUAManager->bInitSdk ) {
           return RET_SDK_ALREADY_INITED;
       }

       pUAManager->config.videoConfigs.nCount = 0;
       pUAManager->config.audioConfigs.nCount = 0;
       for (int count = 0; count < _nSize; ++count) {
               if (_pMediaConfigs[count].streamType == STREAM_VIDEO) {
                       pUAManager->config.videoConfigs.configs[pUAManager->config.videoConfigs.nCount].streamType
                         = RTP_STREAM_VIDEO;
                       pUAManager->config.videoConfigs.configs[pUAManager->config.videoConfigs.nCount].codecType
                         = ConversionFormat(_pMediaConfigs[count].codecType);
                       pUAManager->config.videoConfigs.configs[pUAManager->config.videoConfigs.nCount].nSampleOrClockRate
                         = _pMediaConfigs[count].sampleRate;
                       pUAManager->config.videoConfigs.configs[pUAManager->config.videoConfigs.nCount].nChannel
                         = _pMediaConfigs[count].channels;
                       ++pUAManager->config.videoConfigs.nCount;
               } else if (_pMediaConfigs[count].streamType == STREAM_AUDIO) {
                       pUAManager->config.audioConfigs.configs[pUAManager->config.audioConfigs.nCount].streamType
                         = RTP_STREAM_AUDIO;
                       pUAManager->config.audioConfigs.configs[pUAManager->config.audioConfigs.nCount].codecType
                         = ConversionFormat(_pMediaConfigs[count].codecType);
                       pUAManager->config.audioConfigs.configs[pUAManager->config.audioConfigs.nCount].nSampleOrClockRate
                          = _pMediaConfigs[count].sampleRate;
                       pUAManager->config.audioConfigs.configs[pUAManager->config.audioConfigs.nCount].nChannel
                         = _pMediaConfigs[count].channels;
                       ++pUAManager->config.audioConfigs.nCount;
               }
        }
        config.Cb.OnIncomingCall  = &cbOnIncomingCall;
        config.Cb.OnCallStateChange = &cbOnCallStateChange;
        config.Cb.OnRegStatusChange = &cbOnRegStatusChange;
        config.nMaxCall = MAX_CALL_COUNT;
        config.nMaxAccount = MAX_ACCOUNT;
        pUAManager->config.callback.OnRxRtp = &OnRxRtp;
        // debug code.
        SetLogLevel(LOG_VERBOSE);
        SipCreateInstance(&config);
        INIT_LIST_HEAD(&pUAManager->UAList.list);
        pUAManager->bInitSdk = true;
        pthread_mutex_init(&pUAManager->mutex, NULL);

        InitialiseRtp();
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
        pthread_mutex_lock(&pUAManager->mutex);
        list_for_each_safe(pos, q, &pUAManager->UAList.list){
                pUA = list_entry(pos, UA, list);
                list_del(pos);
                UAUnRegister(pUA);
        }
        pthread_mutex_unlock(&pUAManager->mutex);
        pthread_mutex_destroy(&pUAManager->mutex);
        UninitialiseRtp();
        pUAManager->bInitSdk = false;
        memset(&pUAManager->config.videoConfigs, 0, sizeof(MediaConfigSet));
        memset(&pUAManager->config.audioConfigs, 0, sizeof(MediaConfigSet));

        return RET_OK;
}

AccountID Register(const char* _id, const char* _password, const char* _pSigHost,
                   const char* _pMediaHost, const char* _pImHost)
{
        int nAccountId = 0;
        struct MqttOptions options;
        if (!pUAManager->bInitSdk) {
                DBG_ERROR("not init sdk\n");
                return RET_INIT_ERROR;
        }
        InitMqtt(&options, _id, _password, _pImHost);
        pthread_mutex_lock(&pUAManager->mutex);
        UA *pUA = UARegister(_id, _password, _pSigHost, _pMediaHost, &options, &pUAManager->config);
        int nReason = 0;
    
        if (pUA == NULL) {
                DBG_ERROR("malloc error\n");
                pthread_mutex_unlock(&pUAManager->mutex);
                return RET_MEM_ERROR;
        }
        list_add(&(pUA->list), &(pUAManager->UAList.list));
        pthread_mutex_unlock(&pUAManager->mutex);
        return pUA->id;
}

ErrorID UnRegister(AccountID _nAccountId)
{
        struct list_head *pos;
        pthread_mutex_lock(&pUAManager->mutex);
        DBG_LOG("UnRegister account id %d\n", _nAccountId);
        UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
        if (pUA != NULL) {
                list_del(pos);
                pthread_mutex_unlock(&pUAManager->mutex);
                UAUnRegister(pUA);
                return RET_OK;
        }
        pthread_mutex_unlock(&pUAManager->mutex);
        return RET_ACCOUNT_NOT_EXIST;
}

ErrorID MakeCall(AccountID _nAccountId, const char* id, const char* _pDestUri, OUT int* _pCallId)
{
        struct list_head *pos;
        ErrorID res = RET_ACCOUNT_NOT_EXIST;

        if (!_pDestUri || !_pCallId || !id)
            return RET_PARAM_ERROR;

        pthread_mutex_lock(&pUAManager->mutex);
        *_pCallId = nSdkCallId++;
        UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
        if (pUA != NULL) {
                res = UAMakeCall(pUA, id, _pDestUri, *_pCallId);
        }
        pthread_mutex_unlock(&pUAManager->mutex);

        return res;
}

ErrorID PollEvent(AccountID _nAccountID, EventType* _pType, Event** _pEvent, int _nTimeOut )
{
        Message *pMessage = NULL;
        struct list_head *pos, *q;
        UA *pUA = NULL;

        if (!_pType || !_pEvent ) {
                return RET_PARAM_ERROR;
        }

        pthread_mutex_lock(&pUAManager->mutex);
        pUA = FindUA(pUAManager, _nAccountID, &pos);
        if (pUA == NULL) {
                DBG_ERROR( "RET_ACCOUNT_NOT_EXIST\n");
                pthread_mutex_unlock(&pUAManager->mutex);
                return RET_ACCOUNT_NOT_EXIST;
        }
        // pLastMessage use to free last message
        if ( pUA->pLastMessage ) {
                Event *pEvent = (Event *) pUA->pLastMessage->pMessage;
                if (pUA->pLastMessage->nMessageID == EVENT_DATA) {
                        if (pEvent->body.dataEvent.data) {
                                free(pEvent->body.dataEvent.data);
                                pEvent->body.dataEvent.data = NULL;
                        }
                }
                if (pUA->pLastMessage->nMessageID == EVENT_MESSAGE) {
                        if (pEvent->body.messageEvent.message) {
                                free(pEvent->body.messageEvent.message);
                                pEvent->body.messageEvent.message = NULL;
                        }
                        if (pEvent->body.messageEvent.topic) {
                                free(pEvent->body.messageEvent.topic);
                                pEvent->body.messageEvent.topic = NULL;
                        }
                }
                if (pUA->pLastMessage->nMessageID == EVENT_CALL) {
                        if (pEvent->body.callEvent.context) {
                                free(pEvent->body.callEvent.context);
                                pEvent->body.callEvent.context = NULL;
                        }
                }
                free(pEvent);
                pEvent = NULL;
                pUA->pLastMessage = NULL;
        }
        pthread_mutex_unlock(&pUAManager->mutex);
        //DBG_LOG("wait for event, pUA = %p\n", pUA );

        if (_nTimeOut) {
                pMessage = ReceiveMessageTimeout( pUA->pQueue, _nTimeOut );
        } else {
                pMessage = ReceiveMessage( pUA->pQueue );
        }

        pthread_mutex_lock(&pUAManager->mutex);
        pUA = FindUA(pUAManager, _nAccountID, &pos);
        if (pUA == NULL) {
                DBG_ERROR( "RET_ACCOUNT_NOT_EXIST\n");
                pthread_mutex_unlock(&pUAManager->mutex);
                return RET_ACCOUNT_NOT_EXIST;
        }

        //DBG_LOG("[ LIBSDK ]get one event\n");
        if (!pMessage) {
                pthread_mutex_unlock(&pUAManager->mutex);
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
        pthread_mutex_unlock(&pUAManager->mutex);
        return RET_OK;
}

ErrorID AnswerCall(AccountID id, int _nCallId)
{
        struct list_head *pos;
        ErrorID error = RET_ACCOUNT_NOT_EXIST;
        pthread_mutex_lock(&pUAManager->mutex);
        UA *pUA = FindUA(pUAManager, id, &pos);
        if (pUA != NULL) {
                error = UAAnswerCall(pUA, _nCallId);
        }
        pthread_mutex_unlock(&pUAManager->mutex);
        return error;
}

ErrorID RejectCall( AccountID id, int _nCallId )
{
        struct list_head *pos;
        ErrorID error = RET_ACCOUNT_NOT_EXIST;
        pthread_mutex_lock(&pUAManager->mutex);
        UA *pUA = FindUA(pUAManager, id, &pos);
        if (pUA != NULL) {
                error = UARejectCall(pUA, _nCallId);
        }
        pthread_mutex_unlock(&pUAManager->mutex);
        return error;
}

ErrorID HangupCall( AccountID id, int _nCallId )
{
        struct list_head *pos;
        ErrorID error = RET_ACCOUNT_NOT_EXIST;
        pthread_mutex_lock(&pUAManager->mutex);
        UA *pUA = FindUA(pUAManager, id, &pos);
        if (pUA != NULL) {
                error = UAHangupCall(pUA, _nCallId);
        }
        pthread_mutex_unlock(&pUAManager->mutex);
        return error;
}

ErrorID SendPacket(AccountID id, int _nCallId, Stream streamID, const uint8_t* buffer, int size, int64_t nTimestamp)
{
        struct list_head *pos;
        ErrorID error = RET_ACCOUNT_NOT_EXIST;
        if (streamID == STREAM_AUDIO && size > MAX_AUDIO_SIZE) {
                return RET_PARAM_ERROR;
        }
        pthread_mutex_lock(&pUAManager->mutex);
        UA *pUA = FindUA(pUAManager, id, &pos);
        if (pUA != NULL) {
                error = UASendPacket(pUA, _nCallId, streamID, buffer, size, nTimestamp);
        }
        pthread_mutex_unlock(&pUAManager->mutex);
        return error;
}

ErrorID Report(AccountID id, const char* topic, const char* message, int length)
{
        struct list_head *pos;
        ErrorID error = RET_ACCOUNT_NOT_EXIST;
        pthread_mutex_lock(&pUAManager->mutex);
        UA *pUA = FindUA(pUAManager, id, &pos);
        if (pUA != NULL) {
                error = UAReport(pUA, topic, message, length);
        }
        pthread_mutex_unlock(&pUAManager->mutex);
        return error;
}

ErrorID Subscribe(AccountID id, const char* topic)
{
        struct list_head *pos;
        ErrorID error = RET_ACCOUNT_NOT_EXIST;
        pthread_mutex_lock(&pUAManager->mutex);
        UA *pUA = FindUA(pUAManager, id, &pos);
        if (pUA != NULL) {
                error = UASubscribe(pUA, topic);
        }
        pthread_mutex_unlock(&pUAManager->mutex);
        return error;
}

ErrorID Unsubscribe(AccountID id, const char* topic)
{
        struct list_head *pos;
        ErrorID error = RET_ACCOUNT_NOT_EXIST;
        pthread_mutex_lock(&pUAManager->mutex);
        UA *pUA = FindUA(pUAManager, id, &pos);
        if (pUA != NULL) {
                error = UAUnsubscribe(pUA, topic);
        }
        pthread_mutex_unlock(&pUAManager->mutex);
        return error;
}

SipAnswerCode cbOnIncomingCall(const int _nAccountId, const const char *_pFrom, const void *_pUser, IN const void *_pMedia, OUT int *pCallId)
{   
        pthread_mutex_lock(&pUAManager->mutex);
        const UA *_pUA = _pUser;
        struct list_head *pos;
        UA *pUA = FindUA(pUAManager, _nAccountId, &pos);
        if (pUA == NULL) {
                pthread_mutex_unlock(&pUAManager->mutex);
                return DOES_NOT_EXIST_ANYWHERE;
        }
    
        DBG_LOG("incoming call From %s to %d %p call id %d\n", _pFrom, _nAccountId, _pMedia, nSdkCallId);
        UAOnIncomingCall(pUA, nSdkCallId, _pFrom, _pMedia);
        *pCallId = nSdkCallId++;
        pthread_mutex_unlock(&pUAManager->mutex);
        return OK;
}

void cbOnRegStatusChange(const int _nAccountId, const SipAnswerCode _regStatusCode, const void *_pUser )
{
        DBG_LOG("pUA address is, _regStatusCode = %d\n", _regStatusCode );
        Message *pMessage = (Message *) malloc( sizeof(Message) );
        Event *pEvent = (Event *) malloc( sizeof(Event) );
        if (!pMessage || !pEvent) {
                if (pMessage) free(pMessage);
                if (pEvent) free(pEvent);
                DBG_ERROR("malloc error\n");
                return;
        }
        CallEvent *pCallEvent = NULL;
        UA *_pUA = ( UA *)_pUser;
        struct list_head *pos;
        UA *pUA = FindUA(pUAManager, _nAccountId, &pos);

        if (pUA == NULL) {
                DBG_ERROR("pUser is NULL %p\n", _pUA);
                free(pMessage);
                free(pEvent);
                return;
        }
    
        DBG_VAL(_nAccountId);
        memset( pMessage, 0, sizeof(Message) );
        memset( pEvent, 0, sizeof(Event) );
        pMessage->nMessageID = EVENT_CALL;
        pCallEvent = &pEvent->body.callEvent;
        pCallEvent->callID = 0;
        if (_regStatusCode == OK) {
                pCallEvent->status = CALL_STATUS_REGISTERED;
                pCallEvent->reasonCode = REASON_OK;
        } else {
                pCallEvent->status = CALL_STATUS_ERROR;
                pCallEvent->reasonCode = REASON_REGISTERING_FAIL;
                UnRegister(_nAccountId);
        }
        pCallEvent->context = NULL;
        pMessage->pMessage = pEvent;
        if (pUA) {
                DBG_LOG("[ LIBSDK ] SendMessage\n");
                SendMessage( pUA->pQueue, pMessage );
        } else {
                DBG_ERROR("pUA is NULL\n");
                return;
        }

        DBG_LOG("reg status = %d\n", _regStatusCode);
        UAOnRegStatusChange(pUA, _regStatusCode);
        if (pUA) {
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

        DBG_LOG("state = %d, status code = %d callid %d accountid %d\n", _State, _StatusCode, _nCallId, _nAccountId);
        if ( !pMessage || !pEvent ) {
                DBG_ERROR("malloc error\n");
                if (pMessage) free(pMessage);
                if (pEvent) free(pEvent);
                return;
        }
        pthread_mutex_lock(&pUAManager->mutex);
        UA *pUA = FindUA(pUAManager, _nAccountId, &pos);

        if (pUA == NULL) {
                pthread_mutex_unlock(&pUAManager->mutex);
                DBG_ERROR("pUser is NULL\n");
                free(pMessage);
                free(pEvent);
                return;
        }

        memset(pMessage, 0, sizeof(Message));
        memset(pEvent, 0, sizeof(Event));
        pMessage->nMessageID = EVENT_CALL;
        pCallEvent = &pEvent->body.callEvent;
        UAOnCallStateChange(pUA, _nCallId, _State, _StatusCode, pMedia, &pCallEvent->callID, &pCallEvent->reasonCode);

        pUA = FindUA(pUAManager, _nAccountId, &pos);

        if (pUA == NULL) {
                pthread_mutex_unlock(&pUAManager->mutex);
                DBG_ERROR("pUser is NULL\n");
                free(pMessage);
                free(pEvent);
                return;
        }
        if (_State == INV_STATE_DISCONNECTED) {
                if (_StatusCode == BUSY_HERE) {
                        pCallEvent->status = CALL_STATUS_REJECT;
                        pCallEvent->reasonCode = REASON_OK;
                } else if (pCallEvent->reasonCode != REASON_OK) {
                        pCallEvent->status = CALL_STATUS_ERROR;
                } else if (_StatusCode == OK) {
                        pCallEvent->status = CALL_STATUS_HANGUP;
                } else {
                        DBG_ERROR("state = %d, status code = %d callid %d accountid %d\n", _State, _StatusCode, _nCallId, _nAccountId);
                        pCallEvent->reasonCode = _StatusCode;
                        pCallEvent->status = CALL_STATUS_ERROR;
                }
        } else if (_State == INV_STATE_CALLING) {
                pCallEvent->status = CALL_STATUS_RING;
        } else {
                DBG_ERROR("Not handle state %d\n", _State);
                free(pMessage);
                free(pEvent);
                pthread_mutex_unlock(&pUAManager->mutex);
                return;
        }
        pCallEvent->context = NULL;
        pMessage->pMessage  = (void *)pEvent;
        SendMessage(pUA->pQueue, pMessage);
        pthread_mutex_unlock(&pUAManager->mutex);
}

void SetLogLevel(int level) {
        if (pUAManager->bInitSdk) {
              SetDebugLogLevel(level);   
        } else {
              DBG_ERROR("SDK is not init");
        }
}
