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

static Call* FindCall(UA* _pUa, int _nCallId, struct list_head *pos)
{
        Call* pCall;
        struct list_head *q;
        list_for_each_safe(pos, q, &_pUa->callList.list) {
                pCall = list_entry(pos, Call, list);
                if (pCall->id == _nCallId) {
                        return pCall;
                }
        }
        return NULL;
}

void InitMqtt(struct MqttOptions option)
{
//Init option.
}

// register a account
// @return UA struct point. If return NULL, error.
UA* UARegister(const char* _pId, const char* _pPassword, const char* _pSigHost,
               const char* _pMediaHost, const char* _pImHost, int _nTimeOut)
{
        UA *pUA = ( UA *) malloc ( sizeof(UA) );
        int nReason = 0;

        if (!pUA) {
                DBG_ERROR("malloc error\n");
                return NULL;
        }
        memset( pUA, 0, sizeof(UA) );

        pUA->pQueue = CreateMessageQueue(MESSAGE_QUEUE_MAX);
        if (!pUA->pQueue) {
                DBG_ERROR("queue malloc fail\n");
                free(pUA);
                return NULL;
        }
        list_add( &(pUA->list), &(pUAManager->UAList.list) );
        int nAccountId = SipAddNewAccount( _pId, _pPassword, _pSigHost, (void *)pUA );
        SipRegAccount( nAccountId, 0 );

        //mqtt create instance.
        struct MqttOptions option;
        InitMqtt(option);
        pUA->pMqttInstance = MqttCreateInstance(&option);
        pUA->id = nAccountId;
        return pUA;
}

ErrorID UAUnRegister(UA* _pUa)
{
        SipRegAccount(_pUa->id, 1 );
        MqttDestroy(_pUa->pMqttInstance);
        _pUa->pMqttInstance = NULL;
        free(_pUa);
}

// make a call, user need to save call id
ErrorID UAMakeCall(UA* _pUa, const char* id, const char* host, OUT int* callID)
{
        Call* call = CALLMakeCall(_pUa->id, id, host, callID);
        list_add(&(call->list), &(_pUa->callList.list));
        return RET_OK;
}

ErrorID UAAnswerCall(UA* _pUa, int nCallId)
{
        struct list_head *pos;
        Call* call = FindCall(_pUa, nCallId, pos);
        if (call) {
                return CALLAnswerCall(call);
        }
        else {
                return RET_CALL_NOT_EXIST;
        }
}

ErrorID UARejectCall(UA* _pUa, int nCallId)
{
        struct list_head *pos;
        Call* call = FindCall(_pUa, nCallId, pos);
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
        struct list_head *pos;
        Call* call = FindCall(_pUa, nCallId, pos);
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
        Call* call = FindCall(_pUa, nCallId, pos);
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
        Call* call = FindCall(_pUa, nCallId, pos);
        if (call) {
                return CALLPollEvent(call, _pType, _pEvent, _pTimeOut);
        }
        else {
                fprintf(stderr, "Call is not exist\n");
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
        SipAnswerCode code = CallOnIncomingCall(call, _nCallId, pFrom, pMedia);
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
        struct list_head *pos;
        Call* call = FindCall(_pUa, nCallId, pos);
        if (call) {
                CALLOnCallStateChange(call, State, StatusCode, pMedia);
                if (StatusCode >= 400 || State == INV_STATE_DISCONNECTED) {
                                list_del(pos);
                }
        }
}
