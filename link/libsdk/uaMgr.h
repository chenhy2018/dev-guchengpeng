#ifndef UAMAR_H
#define UAMAR_H

#include "sdk_local.h"

// register a account
// @return UA struct point. If return NULL, error.
#ifdef WITH_P2P
UA* UARegister(const char* _pId, const char* _pPassword, const char* _pSigHost,
               const char* _pMediaHost, MqttOptions* _pOptions,
               UAConfig* _pConfig);
#else
UA* UARegister(const char* _pId, const char* _pPassword, const char* _pSigHost,
               MqttOptions* _pOptions, UAConfig* _pConfig);
#endif

ErrorID UAUnRegister(UA* _pUa);
// make a call, user need to save call id
ErrorID UAMakeCall(UA* _pUa, const char* _pId, const char* _pHost, IN int _nCallId);
ErrorID UAAnswerCall(UA* _pUa, int _nCallId);
ErrorID UARejectCall(UA* _pUa, int _nCallId);
// hangup a call
ErrorID UAHangupCall(UA* _pUa, int _nCallId);

#ifdef WITH_P2P
// send a packet
ErrorID UASendPacket(UA* _pUa, int _nCallID, Stream _nStreamID, const uint8_t* _pBuffer, int _nSize, int64_t _nTimestamp);
#endif

// poll a event
// if the EventData have video or audio data
// the user shuould copy the the packet data as soon as possible
ErrorID UAPollEvent(UA* _pUa, EventType* _pType, Event* _pEvent, int _nTimeOut);

// mqtt report
ErrorID UAReport(UA* _pUa, const char* topic, const char* _pMessage, int _nLength);
ErrorID UASubscribe(UA* _pUa, const char* topic);
ErrorID UAUnsubscribe(UA* _pUa, const char* topic);

void UADeleteCall(UA* _pUa, const int _nCallId);

SipAnswerCode UAOnIncomingCall(UA* _pUa, const int _nCallId, const char *_pFrom, const void *_pMedia);

void UAOnRegStatusChange(UA* _pUa, const SipAnswerCode _nRegStatusCode);

void UAOnCallStateChange(UA* _pUa, const int _nCallId, const SipInviteState _nState, const SipAnswerCode _nStatusCode, const void *_pMedia, int* _pId, ReasonCode *_pReason);

#endif
