#ifndef CallMGR_H
#define CallMGR_H

#include "sdk_local.h"

// make a call, user need to save call id
Call* CALLMakeCall(AccountID _nAccountId, const char* id, const char* host, IN const int nCallId,
                   CallConfig* _pConfig);

void CALLMakeNewCall(Call* _pCall);

ErrorID CALLAnswerCall(Call* _pCall);

ErrorID CALLRejectCall(Call* _pCall);
// hangup a call
ErrorID CALLHangupCall(Call* _pCall);

#ifdef WITH_P2P
// send a packet
ErrorID CALLSendPacket(Call* _pCall, Stream streamID, const uint8_t* buffer, int size, int64_t nTimestamp);
#endif

// poll a event
// if the EventData have video or audio data
// the user shuould copy the the packet data as soon as possible
ErrorID CALLPollEvent(Call* _pCall, EventType* type, Event* event, int timeOut);

//Callback, asyn
SipAnswerCode CALLOnIncomingCall(Call** _pCall, const int _nAccountId, const int _nCallId, const char *pFrom,
                                 const void *pMedia, CallConfig* _pConfig);

void CALLOnCallStateChange(Call** _pCall, const SipInviteState State, const SipAnswerCode StatusCode, const void *pMedia);

#endif
