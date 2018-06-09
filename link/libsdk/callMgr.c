#include <string.h>
#include "callMgr.h"
#include "dbg.h"

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

ErrorID InitRtp(Call** _pCall, CallConfig* _pConfig)
{
        Call* pCall = *_pCall;
        // rtp to do. ice config.media info. and check error)
        DBG_LOG("InitRtp aaa \n");
        InitIceConfig(&pCall->iceConfig);
        DBG_LOG("InitRtp bb \n");
        if (_pConfig->turnHost) DBG_LOG("InitRtp _pHost NULL \n");
        strcpy(&pCall->iceConfig.turnHost[0], _pConfig->turnHost);
        strcpy(&pCall->iceConfig.turnUsername[0], "root");// _pId);
        strcpy(&pCall->iceConfig.turnPassword[0], "root"); //_pPassword);
        pCall->iceConfig.userCallback = _pConfig->pCallback->OnRxRtp;
        //todo check status
        DBG_LOG("CALLMakeCall %s %s %s %p\n",
                &pCall->iceConfig.turnHost[0], &pCall->iceConfig.turnUsername[0], &pCall->iceConfig.turnPassword[0], pCall->iceConfig.userCallback);
        InitPeerConnectoin(&pCall->pPeerConnection, &pCall->iceConfig);
        MediaConfigSet *_pVideo = _pConfig->pVideoConfigs;
        DBG_LOG("media config video count %d, streamType %d, codecType %d, nSampleOrClockRate %d \n",
                _pVideo->nCount, _pVideo->configs[0].streamType, _pVideo->configs[0].codecType, _pVideo->configs[0].nSampleOrClockRate);
        AddVideoTrack(pCall->pPeerConnection, _pVideo);
        MediaConfigSet *_pAudio = _pConfig->pAudioConfigs;
        DBG_LOG("media config audio count %d streamType %d, codecType %d, nSampleOrClockRate %d, nChannel %d \n",
                _pAudio->nCount, _pAudio->configs[0].streamType, _pAudio->configs[0].codecType, _pAudio->configs[0].nSampleOrClockRate);
        AddAudioTrack(pCall->pPeerConnection, _pAudio);
        pCall->pAnswer = NULL;
        return RET_OK;
}
                
// make a call, user need to save call id . add parameter for ice info and media info.
Call* CALLMakeCall(AccountID _nAccountId, const char* id, const char* _pDestUri,
                   OUT int* _pCallId, CallConfig* _pConfig) 
{
        DBG_LOG("CALLMakeCall start \n");
        Call* pCall = (Call*)malloc(sizeof(Call));
        if (pCall == NULL) {
                return NULL;
        }
        memset(pCall, 0, sizeof(Call));
        InitRtp(&pCall, _pConfig);
        //createOffer(pCall->pPeerConnection, &pCall->pOffer);
        CreateTmpSDP(&pCall->pOffer);
        setLocalDescription(pCall->pPeerConnection, pCall->pOffer);
        SipMakeNewCall(_nAccountId, _pDestUri, pCall->pOffer, _pCallId);
        pCall->id = *_pCallId;
        pCall->callStatus = CALL_STATUS_REGISTERED;
        CheckCallStatus(pCall, CALL_STATUS_RING);
        DBG_LOG("CALLMakeCall end %p \n", pCall);
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
        return SendRtpPacket(_pCall->pPeerConnection, &packet);
}

SipAnswerCode CALLOnIncomingCall(Call** _pCall, const int _nCallId, const char *pFrom,
                                 const void *pMedia, CallConfig* _pConfig)
{
        Call* pCall = (Call*)malloc(sizeof(Call));
        if (pCall == NULL) {
                return 1;
        }
        memset(pCall, 0, sizeof(Call));
        *_pCall = pCall;
        pCall->id = _nCallId;
        // rtp to do. improved
        pCall->pOffer = pMedia;
        // rtp to do. ice config.media info. and check error)
        DBG_LOG("call %p\n", pCall);
        InitRtp(&pCall, _pConfig);
        setRemoteDescription(pCall->pPeerConnection, pCall->pOffer);
        DBG_LOG("call answer call\n");
        //createAnswer(pCall->pPeerConnection, pCall->pOffer, &pCall->pAnswer);
        CreateTmpSDP(&pCall->pAnswer);
        DBG_LOG("call answer call end\n");
        //setLocalDescription(pCall->pPeerConnection, pCall->pAnswer);
        DBG_LOG("call answer call end 1\n");
}

void CALLOnCallStateChange(Call** _pCall, const SipInviteState State, const SipAnswerCode StatusCode, const void *pMedia)
{
        (*_pCall)->callStatus = State;
        //todo free disconnected call.
        DBG_LOG("stats %d state %d call %p\n", State, StatusCode, *_pCall);
        if (StatusCode >= 400 && (*_pCall)->callStatus == INV_STATE_DISCONNECTED) {
                //CALLHangupCall(*_pCall);
                ReleasePeerConnectoin((*_pCall)->pPeerConnection);
                free(*_pCall);
                DBG_LOG("Free call\n");
        }
}

ErrorID CALLPollEvent(Call* _pCall, EventType* type, Event* event, int timeOut)
{
        //not used in current time.
        return RET_OK;
}
