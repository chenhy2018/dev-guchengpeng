#include <string.h>
#include "callMgr.h"

//to do change the CallStatus to INV_STATE
ErrorID CheckCallStatus(Call* _pCall, CallStatus expectedState)
{
       if (_pCall->pPeerConnection == NULL) {
               return RET_CALL_INVAILD_CONNECTION;
       }
       if (_pCall->pOffer == NULL) {
               return RET_CALL_INVAILD_CONNECTION;
       }
       if (expectedState == CALL_STATUS_INCOMING || expectedState == CALL_STATUS_RING) {
               switch (_pCall->callStatus) {
                       case INV_STATE_NULL:
                              return RET_OK;
                       default:
                              return RET_CALL_INVAILD_OPERATING;
               }
       }
       if (expectedState == CALL_STATUS_ESTABLISHED || expectedState == CALL_STATUS_REJECT) {
               switch (_pCall->callStatus) {
                      case INV_STATE_CALLING:
                      case INV_STATE_INCOMING:
                               return RET_OK;
                      default:
                               return RET_CALL_INVAILD_OPERATING;
               }
       }
       if (expectedState == CALL_STATUS_HANGUP) {
               switch (_pCall->callStatus) {
                      case INV_STATE_CONFIRMED:
                               return RET_OK;
                      default:
                               return RET_CALL_INVAILD_OPERATING;
               }
       }
       // for send packet used.
       if (expectedState == CALL_STATUS_REGISTERED) {
             switch (_pCall->callStatus) {
                      case INV_STATE_CONFIRMED:
                               return RET_OK;
                      default:
                               return RET_CALL_INVAILD_OPERATING;
             }
       }
       return RET_FAIL;
}
// make a call, user need to save call id . add parameter for ice info and media info.
Call* CALLMakeCall(AccountID _nAccountId, const char* id, const char* _pDestUri, OUT int* _pCallId)
{
        Call* pCall = (Call*)malloc(sizeof(Call));
        memset(pCall, 0, sizeof(Call));
        // rtp to do. ice config.media info. and check error)
        InitIceConfig(&pCall->iceConfig);
        InitPeerConnectoin(&pCall->pPeerConnection, &pCall->iceConfig);
        AddVideoTrack(pCall->pPeerConnection, &pCall->videoConfig);
        AddAudioTrack(pCall->pPeerConnection, &pCall->audioConfig);
        createOffer(pCall->pPeerConnection, &pCall->pOffer);
        setLocalDescription(pCall->pPeerConnection, pCall->pOffer);
        pCall->pAnswer = NULL;
        *_pCallId = SipMakeNewCall(_nAccountId, _pDestUri, pCall->pOffer);
        pCall->id = *_pCallId;
        pCall->callStatus = CALL_STATUS_REGISTERED;
        CheckCallStatus(pCall, CALL_STATUS_RING);
        return pCall;
}

ErrorID CALLAnswerCall(Call* _pCall)
{
        ErrorID id = CheckCallStatus(_pCall, CALL_STATUS_ESTABLISHED);
        if (id != RET_OK) {
              return id;
        }
        // rtp to do. ice config.media info. and check error)
#if 0
        InitIceConfig(&pCall->iceConfig);
        InitPeerConnectoin(&pCall->pPeerConnection, &pCall->iceConfig);
        AddVideoTrack(pCall->pPeerConnection, &pCall->videoConfig);
        AddAudioTrack(pCall->pPeerConnection, &pCall->audioConfig);
        setRemoteDescription(_pCall->pPeerConnection, _pCall->pOffer);
        createAnswer(_pCall->pPeerConnection, _pCall->pOffer, &_pCall->pAnswer);
        setLocalDescription(pCall->pPeerConnection, pCall->pAnswer);
#endif
        return SipAnswerCall(_pCall->id, OK, "answser call", _pCall->pAnswer);
}

ErrorID CALLRejectCall(Call* _pCall)
{
        ErrorID id = CheckCallStatus(_pCall, CALL_STATUS_REJECT);
        if (id != RET_OK) {
              return id;
        }
        // rtp to do. ice config.media info. and check error)
#if 0
        InitIceConfig(&pCall->iceConfig);
        InitPeerConnectoin(&pCall->pPeerConnection, &pCall->iceConfig);
        AddVideoTrack(pCall->pPeerConnection, &pCall->videoConfig);
        AddAudioTrack(pCall->pPeerConnection, &pCall->audioConfig);
        setRemoteDescription(_pCall->pPeerConnection, _pCall->pOffer);
        createAnswer(_pCall->pPeerConnection, _pCall->pOffer, &_pCall->pAnswer);
        setLocalDescription(pCall->pPeerConnection, pCall->pAnswer);
#endif
        id = SipAnswerCall(_pCall->id, BUSY_HERE, "reject call", _pCall->pAnswer);
        ReleasePeerConnectoin(_pCall->pPeerConnection);
        free(_pCall);
}

// hangup a call
ErrorID CALLHangupCall(Call* _pCall)
{
        ErrorID id = CheckCallStatus(_pCall, CALL_STATUS_HANGUP);
        if (id != RET_OK) {
              return id;
        }
        // check return.  
        SipHangUp(_pCall->id);
        ReleasePeerConnectoin(_pCall->pPeerConnection);
        free(_pCall);
        return id;
}

// send a packet
ErrorID CALLSendPacket(Call* _pCall, Stream streamID, const uint8_t* buffer, int size, int64_t nTimestamp)
{
        ErrorID id = CheckCallStatus(_pCall, CALL_STATUS_REGISTERED);
        if (id != RET_OK) {
              return id;
        }
        RtpPacket packet = {buffer, size, nTimestamp, streamID};
        return SendPacket_1(_pCall->pPeerConnection, &packet);
        //return RET_OK;
}

SipAnswerCode CALLOnIncomingCall(Call** _pCall, const int _nCallId, const char *pFrom, const void *pMedia)
{
        Call* pCall = (Call*)malloc(sizeof(Call));
        memset(pCall, 0, sizeof(Call));
        *_pCall = pCall;
        pCall->id = _nCallId;
        // rtp to do. improved
        pCall->pOffer = pMedia;
        // rtp to do. ice config.media info. and check error)
        InitIceConfig(&pCall->iceConfig);
        InitPeerConnectoin(&pCall->pPeerConnection, &pCall->iceConfig);
        AddVideoTrack(pCall->pPeerConnection, &pCall->videoConfig);
        AddAudioTrack(pCall->pPeerConnection, &pCall->audioConfig);
        setRemoteDescription(pCall->pPeerConnection, pCall->pOffer);
        createAnswer(pCall->pPeerConnection, pCall->pOffer, &pCall->pAnswer);
        setLocalDescription(pCall->pPeerConnection, pCall->pAnswer);
}

void CALLOnCallStateChange(Call* _pCall, const SipInviteState State, const SipAnswerCode StatusCode, const void *pMedia)
{
        _pCall->callStatus = State;
        //todo free disconnected call.
        if (StatusCode >= 400 || _pCall->callStatus == INV_STATE_DISCONNECTED) {
                ReleasePeerConnectoin(_pCall->pPeerConnection);
                free(_pCall);
        }
}
