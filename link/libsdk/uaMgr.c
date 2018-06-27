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

static Call* FindCall(UA* _pUa, int _nCallId, struct list_head **pos)
{
        Call* pCall;
        struct list_head *q;
        struct list_head *po;
        //DBG_LOG("Findcall in %p %p %p\n", &_pUa->callList.list, *pos, q);
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

static Call* FindCallByActualId(UA* _pUa, int _nCallId, struct list_head **pos)
{       
        Call* pCall;
        struct list_head *q;
        struct list_head *po;
        list_for_each_safe(po, q, &_pUa->callList.list) {
                pCall = list_entry(po, Call, list);
                if (pCall->nActualId == _nCallId) {
                        *pos = po;
                        DBG_LOG("Findcall out %p %p\n", pCall, *pos);
                        return pCall;
                }
        }
        return NULL;
}
// register a account
// @return UA struct point. If return NULL, error.
UA* UARegister(const char* _pId, const char* _pPassword, const char* _pSigHost,
               const char* _pMediaHost, MqttOptions* _pOptions,
               UAConfig* _pConfig)
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
        sipConfig.nMaxOngoingCall = 10;
        int nAccountId = 0;
        DBG_LOG("UARegister %s %s %s %p ongoing call %d\n",
                sipConfig.pUserName, sipConfig.pPassWord, sipConfig.pDomain, sipConfig.pUserData, sipConfig.nMaxOngoingCall);
        if (SipIsUserAlreadyExist(&sipConfig)) {
                DBG_ERROR("user Already Exist\n");
                free(pUA);
                return NULL;
        }
        SipAnswerCode Ret = SipAddNewAccount(&sipConfig, &nAccountId);
        if (Ret != SIP_SUCCESS) {
                DBG_ERROR("Add Account Error, Ret = %d\n", Ret);
                free(pUA);
                return NULL;
        }
        Ret = SipRegAccount(nAccountId, 1);
        if (Ret != SIP_SUCCESS) {
                SipDeleteAccount(nAccountId);
                DBG_ERROR("Register Account Error, Ret = %d\n", Ret);
                free(pUA);
                return NULL;
        }
        pUA->regStatus == TRYING;
        //mqtt create instance.
        _pOptions->nAccountId = nAccountId;
        pUA->pMqttInstance = MqttCreateInstance(_pOptions);
        pUA->id = nAccountId;
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
        SipAnswerCode code = SipRegAccount(_pUa->id, 0);
        MqttDestroy(_pUa->pMqttInstance);
        DestroyMessageQueue(&_pUa->pQueue);
        _pUa->pMqttInstance = NULL;
        free(_pUa);
        if (code !=OK) {
                return RET_OK;
        }
        else {
                return RET_FAIL;
        }
}

// make a call, user need to save call id
ErrorID UAMakeCall(UA* _pUa, const char* id, const char* host, OUT int* callID)
{
        if (_pUa->regStatus == OK) {
                Call* call = CALLMakeCall(_pUa->id, id, host, callID, &_pUa->config);
                if (call == NULL) {
                        return RET_MEM_ERROR;
                }
                DBG_LOG("UAMakeCall in call %p call list %p\n",call, &(call->list));
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
        DBG_LOG("UAAnswerCall in call id %d ua %p\n",nCallId, _pUa);
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
                //list_del(pos);
                ErrorID id = CALLRejectCall(call);
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
                return id;
        }
        else {
                return RET_CALL_NOT_EXIST;
        }
}

// send a packet
ErrorID UASendPacket(UA* _pUa, int nCallId, Stream streamID, const uint8_t* buffer, int size, int64_t nTimestamp)
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
ErrorID UAReport(UA* _pUa,  const char* message, int length)
{
        return MqttPublish(_pUa->pMqttInstance, "/test/test", length, message);
}

SipAnswerCode UAOnIncomingCall(UA* _pUa, const int _nCallId, const char *pFrom, const void *pMedia)
{
        struct list_head *pos;
        Call* call;
        DBG_LOG("UAOnIncomingCall \n");
        SipAnswerCode code = CALLOnIncomingCall(&call, _pUa->id, _nCallId, pFrom, pMedia, &_pUa->config);
        list_add(&(call->list), &(_pUa->callList.list));
        return code;
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

void UAOnCallStateChange(UA* _pUa, const int nCallId, const SipInviteState State, const SipAnswerCode StatusCode, const void *pMedia, int* pId)
{
        struct list_head *pos = NULL;
        DBG_LOG("UA call statue change \n");
        Call* call = FindCallByActualId(_pUa, nCallId, &pos);
        if (call) {
                *pId = call->id;
                DBG_LOG("call %p\n", call);
                //CALLOnCallStateChange(&call, State, StatusCode, pMedia);
                if (State == INV_STATE_DISCONNECTED) {
                                DBG_LOG("*******UAOnCallStateChange del %p \n", pos);
                                list_del(pos);
                }
                CALLOnCallStateChange(&call, State, StatusCode, pMedia);
                DBG_LOG("call change end \n");
        }
}

void UADeleteCall(UA* _pUa, const int nCallId)
{
        struct list_head *pos = NULL;
        DBG_LOG("UA Delete call \n");
        Call* pCall = FindCall(_pUa, nCallId, &pos);
        if (pCall) {
                DBG_LOG("call %p\n", pCall);
                list_del(pos);
                //todo call releasecall
                free(pCall);
                DBG_LOG("call delete call end \n");
        }
}
