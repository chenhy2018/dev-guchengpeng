#include <string.h>
#include "sip.h"
#include "dbg.h"
#include "queue.h"
#include "sdk_interface.h"
#include "mqtt.h"
#include "sdk_local.h"
#include "framework.h"
#include "list.h"
#include "callMgr.h"
#include "qrtc.h"

void OnMessage(IN const void* _pInstance, IN const char* _pTopic, IN const char* _pMessage, IN size_t nLength)
{
        DBG_LOG("%p topic %s message %s \n", _pInstance, _pTopic, _pMessage);
}

void OnEvent(IN const void* _pInstance, IN int _nId,  IN const char* _pReason)
{
        DBG_LOG("%p id %d, reason  %s \n",_pInstance, _nId, _pReason);
        // TODO call back to user.
}

static Call* FindCall(UA* _pUa, int _nCallId, struct list_head **pos)
{
        Call* pCall;
        struct list_head *q;
        struct list_head *po;
        DBG_LOG("Findcall in %p %p %p\n", &_pUa->callList.list, *pos, q);
        list_for_each_safe(po, q, &_pUa->callList.list) {
                pCall = list_entry(po, Call, list);
                if (pCall->id == _nCallId) {
                        *pos = po;
                        DBG_LOG("Findcall out %p %p\n", pCall, *pos);
                        return pCall;
                }
        }
        return NULL;
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

// register a account
// @return UA struct point. If return NULL, error.
UA* UARegister(const char* _pId, const char* _pPassword, const char* _pSigHost,
               const char* _pMediaHost, const char* _pImHost, int _nTimeOut,
               MediaConfig* _pVideo, MediaConfig* _pAudio)
{

        UA *pUA = (UA *) malloc (sizeof(UA));
        int nReason = 0;

        if (!pUA) {
                DBG_ERROR("malloc error\n");
                return NULL;
        }
        memset( pUA, 0, sizeof(UA) );
        INIT_LIST_HEAD(&pUA->callList.list);
        pUA->pQueue = CreateMessageQueue(MESSAGE_QUEUE_MAX);
        if (!pUA->pQueue) {
                DBG_ERROR("queue malloc fail\n");
                free(pUA);
                return NULL;
        }
        SipAccountConfig sipConfig;
        sipConfig.pUserName = _pId;
        sipConfig.pPassWord = _pPassword;
        sipConfig.pDomain = _pSigHost;
        sipConfig.pUserData = (void *)pUA;
        sipConfig.nMaxOngoingCall = 10;
        int nAccountId = 0;
        DBG_LOG("UARegister %s %s %s %p ongoing call %d\n", sipConfig.pUserName, sipConfig.pPassWord, sipConfig.pDomain, sipConfig.pUserData, sipConfig.nMaxOngoingCall);
        SipAddNewAccount(&sipConfig, &nAccountId);
        SipRegAccount(nAccountId, 1);
        pUA->regStatus == TRYING;
        //mqtt create instance.
        struct MqttOptions option;
        InitMqtt(&option, _pId, _pPassword, _pImHost);
        DBG_LOG("NOT MQTT");
        pUA->pMqttInstance = MqttCreateInstance(&option);
        pUA->id = nAccountId;
        pUA->pVideoConfigs = _pVideo;
        pUA->pAudioConfigs = _pAudio;
        strncpy(pUA->turnHost, _pMediaHost, MAX_TURN_HOST_SIZE - 1);
        strncpy(pUA->turnUsername, _pId, MAX_TURN_USR_SIZE -1);
        strncpy(pUA->turnPassword, _pPassword, MAX_TURN_PWD_SIZE -1);
        return pUA;
}

ErrorID UAUnRegister(UA* _pUa)
{
        SipRegAccount(_pUa->id, 0);
        MqttDestroy(_pUa->pMqttInstance);
        _pUa->pMqttInstance = NULL;
        free(_pUa);
}

// make a call, user need to save call id
ErrorID UAMakeCall(UA* _pUa, const char* id, const char* host, OUT int* callID)
{
        if (_pUa->regStatus == OK) {
                Call* call = CALLMakeCall(_pUa->id, id, host, callID, _pUa->pVideoConfigs, _pUa->pAudioConfigs,
                                          _pUa->turnHost, _pUa->turnUsername, _pUa->turnPassword);
                if (call == NULL) {
                        return RET_MEM_ERROR;
                }
                list_add(&(call->list), &(_pUa->callList.list));
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

ErrorID UAAnswerCall(UA* _pUa, int nCallId)
{
        struct list_head *pos = NULL;
        Call* call = FindCall(_pUa, nCallId, &pos);
        if (call) {
                return CALLAnswerCall(call);
        }
        else {
                return RET_CALL_NOT_EXIST;
        }
}

ErrorID UARejectCall(UA* _pUa, int nCallId)
{
        struct list_head *pos = NULL;
        Call* call = FindCall(_pUa, nCallId, &pos);
        if (call) {
                ErrorID id = CALLRejectCall(call);
                list_del(pos);
                return id;
        }
        else {
                return RET_CALL_NOT_EXIST;
        }
}

// hangup a call
ErrorID UAHangupCall(UA* _pUa, int nCallId)
{
        struct list_head *pos = NULL;
        Call* call = FindCall(_pUa, nCallId, &pos);
        if (call) {
                ErrorID id =  CALLHangupCall(call);
                list_del(pos);
                return id;
        }
        else {
                return RET_CALL_NOT_EXIST;
        }
}

// send a packet
ErrorID UASendPacket(UA* _pUa, int nCallId, Stream streamID, const char* buffer, int size, int64_t nTimestamp)
{
        struct list_head *pos;
        Call* call = FindCall(_pUa, nCallId, &pos);
        if (call) {
                return CALLSendPacket(call, streamID, buffer, size, nTimestamp);
        }
        else {
                return RET_CALL_NOT_EXIST;
        }
}
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
        else if (_pEvent->type == EVENT_DATA) {
                DataEvent* event = (DataEvent*)(&_pEvent->body.dataEvent);
                nCallId = event->callID;
        }
        else {
                //To do error event and another event.
        }
        struct list_head *pos;
        Call* call = FindCall(_pUa, nCallId, &pos);
        if (call) {
                return CALLPollEvent(call, _pType, _pEvent, _pTimeOut);
        }
        else {
                DBG_LOG("Call is not exist\n");
                return RET_CALL_NOT_EXIST;
        }
}

// mqtt report
ErrorID UAReport(UA* _pUa, const char* topic, const char* message, int length)
{
        return MqttPublish(_pUa->pMqttInstance, topic, length, message);
}

SipAnswerCode UAOnIncomingCall(UA* _pUa, const int _nCallId, const char *pFrom, const void *pMedia)
{
        struct list_head *pos;
        Call** call;
        SipAnswerCode code = CALLOnIncomingCall(call, _nCallId, pFrom, pMedia, _pUa->pVideoConfigs, _pUa->pAudioConfigs,
                                                _pUa->turnHost, _pUa->turnUsername, _pUa->turnPassword);
        list_add(&((*call)->list), &(_pUa->callList.list));
}

void UAOnRegStatusChange(UA* _pUa, const SipAnswerCode _nRegStatusCode)
{
        if ( _nRegStatusCode == OK ||
             _nRegStatusCode == UNAUTHORIZED || 
             _nRegStatusCode == REQUEST_TIMEOUT ) {
            _pUa->regStatus = _nRegStatusCode;
        }
        else {
            _pUa->regStatus = DECLINE;
        }
}

void UAOnCallStateChange(UA* _pUa, const int nCallId, const SipInviteState State, const SipAnswerCode StatusCode, const void *pMedia)
{
        struct list_head *pos = NULL;
        DBG_LOG("UA call statue change");
        Call* call = FindCall(_pUa, nCallId, &pos);
        if (call) {
                DBG_LOG("call %p", call);
                CALLOnCallStateChange(&call, State, StatusCode, pMedia);
                DBG_LOG("call change end \n");
                if (StatusCode >= 400 && State == INV_STATE_DISCONNECTED) {
                                DBG_LOG("Findcall out %p \n", pos);
                                list_del(pos);
                                DBG_LOG("Findcall out %p \n", pos);
                }
        }
}
