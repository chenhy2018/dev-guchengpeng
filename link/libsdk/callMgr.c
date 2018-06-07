#include <string.h>
#include "callMgr.h"
#include "dbg.h"

// Todo send to message queue.
static void onRxRtp(void *_pUserData, CallbackType _type, void *_pCbData)
{
#if 0
        switch (_type){
                case CALLBACK_ICE:{
                        IceNegInfo *pInfo = (IceNegInfo *)_pCbData;
                        MY_PJ_LOG(3, "==========>callback_ice: state:%d", pInfo->state);
                        for ( int i = 0; i < pInfo->nCount; i++) {
                                MY_PJ_LOG(3, " codec type:%d", pInfo->configs[i]->codecType);
                        }
                }
                        break;
                case CALLBACK_RTP:{
                        RtpPacket *pPkt = (RtpPacket *)_pCbData;
                        pj_ssize_t nLen = pPkt->nDataLen;
                        if (pPkt->type == STREAM_AUDIO && nLen == 160) {
                                pj_file_write(gPcmuFd, pPkt->pData, &nLen);
                        } else if (pPkt->type == STREAM_VIDEO) {
                                pj_file_write(gH264Fd, pPkt->pData, &nLen);
                        }
                }
                        break;
                case CALLBACK_RTCP:
                        fprintf(stderr, "==========>callback_rtcp\n");
                        break;
        }
#endif
}

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
Call* CALLMakeCall(AccountID _nAccountId, const char* id, const char* _pDestUri,
                   OUT int* _pCallId, MediaConfig* _pVideo, MediaConfig* _pAudio,
                   const char* _pId, const char* _pPassword, const char* _pHost) 
{
        DBG_LOG("CALLMakeCall start \n");
        Call* pCall = (Call*)malloc(sizeof(Call));
        if (pCall == NULL) {
                return NULL;
        }
        memset(pCall, 0, sizeof(Call));
        // rtp to do. ice config.media info. and check error)
        InitIceConfig(&pCall->iceConfig);
        strcpy(&pCall->iceConfig.turnHost[0], _pHost);
        strcpy(&pCall->iceConfig.turnUsername[0], _pId);
        strcpy(&pCall->iceConfig.turnPassword[0], _pPassword);
        pCall->iceConfig.userCallback = onRxRtp;
        //todo check status
        InitPeerConnectoin(&pCall->pPeerConnection, &pCall->iceConfig);
        AddVideoTrack(pCall->pPeerConnection, _pVideo);
        AddAudioTrack(pCall->pPeerConnection, _pAudio);
        createOffer(pCall->pPeerConnection, &pCall->pOffer);
        setLocalDescription(pCall->pPeerConnection, pCall->pOffer);
        pCall->pAnswer = NULL;
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
        return SendPacket_1(_pCall->pPeerConnection, &packet);
        //return RET_OK;
}

SipAnswerCode CALLOnIncomingCall(Call** _pCall, const int _nCallId, const char *pFrom,
                                 const void *pMedia, MediaConfig* _pVideo, MediaConfig* _pAudio,
                                 const char* _pId, const char* _pPassword, const char* _pHost)
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
        AddVideoTrack(pCall->pPeerConnection, _pVideo);
        AddAudioTrack(pCall->pPeerConnection, _pAudio);
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

ErrorID CALLPollEvent(Call* _pCall, EventType* type, Event* event, int timeOut)
{
        //not used in current time.
        return RET_OK;
}
