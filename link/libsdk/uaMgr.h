#ifndef UAMAR_H
#define UAMAR_H

#include "sdk_local.h"

// register a account
// @return UA struct point. If return NULL, error.
UA* UARegister(const char* id, const char* password, const char* sigHost,
               const char* mediaHost, MqttOptions* options,
               MediaConfigSet* pVideo, MediaConfigSet* pAudio);

ErrorID UAUnRegister(UA* _pUa);
// make a call, user need to save call id
ErrorID UAMakeCall(UA* _pUa, const char* id, const char* host, OUT int* callID);
ErrorID UAAnswerCall(UA* _pUa, int nCallId);
ErrorID UARejectCall(UA* _pUa, int nCallId);
// hangup a call
ErrorID UAHangupCall(UA* _pUa, int nCallId);
// send a packet
ErrorID UASendPacket(UA* _pUa, int callID, Stream streamID, const char* buffer, int size, int64_t nTimestamp);
// poll a event
// if the EventData have video or audio data
// the user shuould copy the the packet data as soon as possible
ErrorID UAPollEvent(UA* _pUa, EventType* type, Event* event, int timeOut);

// mqtt report
ErrorID UAReport(UA* _pUa, const char* message, int length);

SipAnswerCode UAOnIncomingCall(UA* _pUa, const int _nCallId, const char *pFrom, const void *pMedia);

void UAOnRegStatusChange(UA* _pUa, const SipAnswerCode RegStatusCode);

void UAOnCallStateChange(UA* _pUa, const int nCallId, const SipInviteState State, const SipAnswerCode StatusCode, const void *pMedia);

#endif
