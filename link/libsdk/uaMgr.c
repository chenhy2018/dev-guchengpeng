#include <string.h>
#include "sip.h"
#include "dbg.h"
#include "queue.h"
#ifdef WITH_P2P
#include "sdk_interface_p2p.h"
#else
#include "sdk_interface.h"
#endif
#include "mqtt.h"
#include "sdk_local.h"
#include "list.h"
#include "callMgr.h"
#include "qrtc.h"

static int nSdkAccountId = 0;
static Call* FindCall(UA* _pUa, int _nCallId, struct list_head **pos)
{
        Call* pCall;
        struct list_head *q;
        struct list_head *po;
        //        DBG_LOG("Findcall in %p %p %p\n", &_pUa->callList.list, *pos, q);
        list_for_each_safe(po, q, &_pUa->callList.list) {
                pCall = list_entry(po, Call, list);
                if (pCall->id == _nCallId) {
                        *pos = po;
                        //DBG_LOG("Findcall out %p %p\n", pCall, *pos);
                        return pCall;
                }
        }
        return NULL;
}

// register a account
// @return UA struct point. If return NULL, error.
#ifdef WITH_P2P
UA* UARegister(const char* _pId, const char* _pPassword, const char* _pSigHost,
               const char* _pMediaHost, MqttOptions* _pOptions,
               UAConfig* _pConfig)
#else
UA* UARegister(const char* _pId, const char* _pPassword, const char* _pSigHost,
               MqttOptions* _pOptions, UAConfig* _pConfig)
#endif
{

        UA *pUA = (UA *) malloc (sizeof(UA));
        int nReason = 0;

        if (!pUA) {
                DBG_ERROR("malloc error\n");
                return NULL;
        }
        memset( pUA, 0, sizeof(UA) );

        SipAccountConfig sipConfig;
        sipConfig.pUserName = (char*)_pId;
        sipConfig.pPassWord = (char*)_pPassword;
        sipConfig.pDomain = (char*)_pSigHost;
        sipConfig.pUserData = (void *)pUA;
        sipConfig.nMaxOngoingCall = MAX_ONGOING_CALL_COUNT;
        DBG_LOG("UARegister %s %s %s %p ongoing call %d\n",
                sipConfig.pUserName, sipConfig.pPassWord, sipConfig.pDomain, sipConfig.pUserData, sipConfig.nMaxOngoingCall);
        if (_pSigHost != NULL) {        
                SipAnswerCode Ret = SipRegAccount(&sipConfig, nSdkAccountId);
                if (Ret != SIP_SUCCESS) {
                        DBG_ERROR("Register Account Error, Ret = %d\n", Ret);
                        free(pUA);
                        return NULL;
                }
                pUA->regStatus = TRYING;
        } else {
                pUA->regStatus = NOT_FOUND;
        }
        //mqtt create instance.
        _pOptions->nAccountId = nSdkAccountId;
        if (_pOptions->userInfo.pHostname) {
                pUA->pMqttInstance = MqttCreateInstance(_pOptions);
        }
        pUA->id = nSdkAccountId;
        strncpy(pUA->userId, _pId, MAX_USER_ID_SIZE - 1);
#ifdef WITH_P2P
        pUA->config.pVideoConfigs = &_pConfig->videoConfigs;
        pUA->config.pAudioConfigs = &_pConfig->audioConfigs;
        pUA->config.pCallback = &_pConfig->callback;
        if (_pMediaHost) {
                strncpy(pUA->config.turnHost, _pMediaHost, MAX_TURN_HOST_SIZE - 1);
        }
        if (_pId) {
                strncpy(pUA->config.turnUsername, _pId, MAX_TURN_USR_SIZE -1);
        }
        if (_pPassword) {
                strncpy(pUA->config.turnPassword, _pPassword, MAX_TURN_PWD_SIZE -1);
        }
#else
        if (_pSigHost != NULL) {
                //It may cause crash because not call pj_thread_register
                CreateTmpSDP(&pUA->config.pSdp);
        }
#endif
        nSdkAccountId++;
        INIT_LIST_HEAD(&pUA->callList.list);
        pUA->pQueue = CreateMessageQueue(MESSAGE_QUEUE_MAX);
        if (!pUA->pQueue) {
                DBG_ERROR("queue malloc fail\n");
                free(pUA);
                return NULL;
        }
        return pUA;
}

ErrorID UAUnRegister(UA* _pUa)
{
        SipAnswerCode code = OK;
        if (_pUa->regStatus != NOT_FOUND) {
                code = SipUnRegAccount(_pUa->id);
        }
        if (_pUa->pMqttInstance) {
                MqttDestroy(_pUa->pMqttInstance);
                _pUa->pMqttInstance = NULL;
        }
        DestroyMessageQueue(&_pUa->pQueue);
        free(_pUa);
        if (code !=OK) {
                return RET_OK;
        }
        else {
                return RET_FAIL;
        }
}

// make a call, user need to save call id
ErrorID UAMakeCall(UA* _pUa, const char* _pId, const char* _pHost, int _nCallId)
{
        if (_pUa->regStatus == OK) {
                Call* pCall = CALLMakeCall(_pUa->id, _pId, _pHost, _nCallId, &_pUa->config);
                if (pCall == NULL) {
                        return RET_MEM_ERROR;
                }
                DBG_LOG("UAMakeCall in call %p call list %p\n",pCall, &(pCall->list));
                list_add(&(pCall->list), &(_pUa->callList.list));
                return RET_OK;
        }
        else if (_pUa->regStatus == UNAUTHORIZED) {
                return RET_USER_UNAUTHORIZED;
        }
        else if (_pUa->regStatus == REQUEST_TIMEOUT) {
                return RET_REGISTER_TIMEOUT;
        }
        else if (_pUa->regStatus == TRYING) {
                return RET_REGISTERING;
        }
        else {
                return RET_FAIL;
        }
}

ErrorID UAAnswerCall(UA* _pUa, int _nCallId)
{
        struct list_head *pos = NULL;
        DBG_LOG("UAAnswerCall in call id %d ua %p\n",_nCallId, _pUa);
        Call* pCall = FindCall(_pUa, _nCallId, &pos);
        if (pCall) {
                return CALLAnswerCall(pCall);
        }
        else {
                return RET_CALL_NOT_EXIST;
        }
}

ErrorID UARejectCall(UA* _pUa, int _nCallId)
{
        struct list_head *pos = NULL;
        Call* pCall = FindCall(_pUa, _nCallId, &pos);
        if (pCall) {
                //list_del(pos);
                ErrorID id = CALLRejectCall(pCall);
                return id;
        }
        else {
                return RET_CALL_NOT_EXIST;
        }
}

// hangup a call
ErrorID UAHangupCall(UA* _pUa, int _nCallId)
{
        struct list_head *pos = NULL;
        Call* pCall = FindCall(_pUa, _nCallId, &pos);
        if (pCall) {
                ErrorID id =  CALLHangupCall(pCall);
                return id;
        }
        else {
                return RET_CALL_NOT_EXIST;
        }
}

#ifdef WITH_P2P
// send a packet
ErrorID UASendPacket(UA* _pUa, int _nCallId, Stream _nStreamID, const uint8_t * _pBuffer, int _nSize, int64_t _nTimestamp)
{
        struct list_head *pos;
        Call* pCall = FindCall(_pUa, _nCallId, &pos);
        if (pCall) {
                return CALLSendPacket(pCall, _nStreamID, _pBuffer, _nSize, _nTimestamp);
        }
        else {
                return RET_CALL_NOT_EXIST;
        }
}
#endif

// poll a event
// if the EventData have video or audio data
// the user shuould copy the the packet data as soon as possible
ErrorID UAPollEvent(UA* _pUa, EventType* _pType, Event* _pEvent, int _pTimeOut)
{
        int nCallId = 0; //todo. need change interface.
        if (_pEvent->type == EVENT_CALL) {
                CallEvent* event = (CallEvent*)(&_pEvent->body.callEvent);
                nCallId = event->callID;
        }
#ifdef WITH_P2P
        else if (_pEvent->type == EVENT_DATA) {
                DataEvent* event = (DataEvent*)(&_pEvent->body.dataEvent);
                nCallId = event->callID;
        }
#endif
        else {
                //To do error event and another event.
        }
        struct list_head *pos;
        Call* pCall = FindCall(_pUa, nCallId, &pos);
        if (pCall) {
                return CALLPollEvent(pCall, _pType, _pEvent, _pTimeOut);
        }
        else {
                DBG_LOG("Call is not exist\n");
                return RET_CALL_NOT_EXIST;
        }
}

// mqtt report
ErrorID UAReport(UA* _pUa, const char* topic, const char* message, int length)
{
        if (_pUa->pMqttInstance == NULL) {
                return RET_PARAM_ERROR;
        }
        MQTT_ERR_STATUS res = MqttPublish(_pUa->pMqttInstance, topic, length, message);
        if (res == MQTT_SUCCESS) {
                return RET_OK;
        } else {
                return res;
        }
}

ErrorID UASubscribe(UA* _pUa, const char* topic)
{
        if (_pUa->pMqttInstance == NULL) {
                return RET_PARAM_ERROR;
        }
        MQTT_ERR_STATUS res = MqttSubscribe(_pUa->pMqttInstance, topic);
        if (res == MQTT_SUCCESS) {
                return RET_OK;
        } else {
                return res;
        }
}

ErrorID UAUnsubscribe(UA* _pUa, const char* topic)
{
        if (_pUa->pMqttInstance == NULL) {
                return RET_PARAM_ERROR;
        }
        MQTT_ERR_STATUS res = MqttUnsubscribe(_pUa->pMqttInstance, topic);
        if (res == MQTT_SUCCESS) {
                return RET_OK;
        } else {
                return res;
        }
}

SipAnswerCode UAOnIncomingCall(UA* _pUa, const int _nCallId, const char *_pFrom, const void *_pMedia)
{
        struct list_head *pos;
        Call* pCall;
        DBG_LOG("UAOnIncomingCall \n");
        SipAnswerCode code = CALLOnIncomingCall(&pCall, _pUa->id, _nCallId, _pFrom, _pMedia, &_pUa->config);
        list_add(&(pCall->list), &(_pUa->callList.list));
        return code;
}

void UAOnRegStatusChange(UA* _pUa, const SipAnswerCode _nRegStatusCode)
{
        if ( _nRegStatusCode == OK ||
             _nRegStatusCode == UNAUTHORIZED || 
             _nRegStatusCode == REQUEST_TIMEOUT ) {
            _pUa->regStatus = _nRegStatusCode;
        } else {
            _pUa->regStatus = DECLINE;
        }
}

void UAOnCallStateChange(UA* _pUa, const int _nCallId, const SipInviteState _nState, const SipAnswerCode _nStatusCode, const void *_pMedia, int *_pId, ReasonCode *_pReasonCode)
{
        struct list_head *pos = NULL;
        DBG_LOG("UA call statue change \n");
        Call* pCall = FindCall(_pUa, _nCallId, &pos);
        if (pCall) {
                *_pId = pCall->id;
                *_pReasonCode = pCall->error;
                DBG_LOG("call %p\n", pCall);
                if (_nState == INV_STATE_DISCONNECTED) {
                                DBG_LOG("*******UAOnCallStateChange del %p \n", pos);
                                list_del(pos);
                }
                CALLOnCallStateChange(&pCall, _nState, _nStatusCode, _pMedia);
                DBG_LOG("call change end \n");
        }
}

void UADeleteCall(UA* _pUa, const int _nCallId)
{
        struct list_head *pos = NULL;
        DBG_LOG("UA Delete call \n");
        Call* pCall = FindCall(_pUa, _nCallId, &pos);
        if (pCall) {
                DBG_LOG("call %p\n", pCall);
                list_del(pos);
                free(pCall);
                DBG_LOG("call delete call end \n");
        }
}
