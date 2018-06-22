#include <string.h>
#include "callMgr.h"
#include "dbg.h"

static int CallId = 0;
//to do change the CallStatus to INV_STATE
ErrorID CheckCallStatus(Call* _pCall, CallStatus expectedState)
{
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
        int res = 0;
        Call* pCall = *_pCall;
        // rtp to do. ice config.media info. and check error)
        InitIceConfig(&pCall->iceConfig);
        if (!_pConfig->turnHost) DBG_LOG("InitRtp _pHost NULL \n");
#if 1
        strncpy(&pCall->iceConfig.turnHost[0], "123.59.204.198:4478", MAX_TURN_HOST_SIZE);//_pConfig->turnHost, MAX_TURN_HOST_SIZE);
        //strncpy(&pCall->iceConfig.turnUsername[0], "root", MAX_TURN_USR_SIZE);// _pId);
        //strncpy(&pCall->iceConfig.turnPassword[0], "root", MAX_TURN_PWD_SIZE); //_pPassword);
#else
        strncpy(&pCall->iceConfig.turnHost[0], _pConfig->turnHost, MAX_TURN_HOST_SIZE);
        strncpy(&pCall->iceConfig.turnUsername[0], "root", MAX_TURN_USR_SIZE);// _pId);
        strncpy(&pCall->iceConfig.turnPassword[0], "root", MAX_TURN_PWD_SIZE); //_pPassword);
#endif
        pCall->iceConfig.userCallback = _pConfig->pCallback->OnRxRtp;
        pCall->iceConfig.pCbUserData = *_pCall;
        //todo check status
        DBG_LOG("CALLMakeCall %s %s %s %p\n",
                &pCall->iceConfig.turnHost[0], &pCall->iceConfig.turnUsername[0], &pCall->iceConfig.turnPassword[0], pCall->iceConfig.userCallback);
        res = InitPeerConnectoin(&pCall->pPeerConnection, &pCall->iceConfig);
        if (res != 0) {
                DBG_ERROR("InitPeerConnectoin failed %d \n", res);
                return RET_INTERAL_FAIL;
        }
        MediaConfigSet *_pVideo = _pConfig->pVideoConfigs;
        DBG_LOG("media config video count %d, streamType %x, codecType %x, nSampleOrClockRate %d \n",
                _pVideo->nCount, _pVideo->configs[0].streamType, _pVideo->configs[0].codecType, _pVideo->configs[0].nSampleOrClockRate);
        res = AddVideoTrack(pCall->pPeerConnection, _pVideo);
        if (res != 0) {
                DBG_ERROR("InitPeerConnectoin failed %d \n", res);
                return RET_INTERAL_FAIL;
        }
        MediaConfigSet *_pAudio = _pConfig->pAudioConfigs;
        DBG_LOG("media config audio count %d streamType %d, codecType %d, nSampleOrClockRate %d, nChannel %d \n",
                _pAudio->nCount, _pAudio->configs[0].streamType, _pAudio->configs[0].codecType, _pAudio->configs[0].nSampleOrClockRate, _pAudio->configs[0].nChannel);
        res = AddAudioTrack(pCall->pPeerConnection, _pAudio);
        if (res != 0) {
                DBG_ERROR("InitPeerConnectoin failed %d \n", res);
                return RET_INTERAL_FAIL;
        }
        pCall->pRemote = NULL;
        return RET_OK;
}

void CALLMakeNewCall(Call* _pCall)
{
        int nCallId;
        DBG_LOG("CALLMakeCall start url %s accountId %d pLocal %p\n", _pCall->url, _pCall->nAccountId, _pCall->pLocal);
        SipAnswerCode error = SipMakeNewCall(_pCall->nAccountId, _pCall->url, _pCall->pLocal, &nCallId);
        if (error != SIP_SUCCESS) {
                DBG_ERROR("SipMakeNewCall failed %d \n", error);
                return;
        }
        _pCall->nActualId = nCallId;
        DBG_LOG("CALLMakeCall end call id %d", _pCall->nActualId);
}

// make a call, user need to save call id . add parameter for ice info and media info.
Call* CALLMakeCall(AccountID _nAccountId, const char* id, const char* _pDestUri,
                   OUT int* _pCallId, CallConfig* _pConfig) 
{
        DBG_LOG("CALLMakeCall start \n");
        int nSize = 0;
        Call* pCall = (Call*)malloc(sizeof(Call));
        if (pCall == NULL) {
                return NULL;
        }

        nSize = strlen(id) + strlen(_pDestUri) + 50;// <sip:id@_pDestUri>
        char *pUri = (char *) malloc( nSize );
        if ( !pUri ) {
            DBG_ERROR("[ LIBSDK ] malloc error, malloc size %d\n", nSize );
            return NULL;
        }
        memset( pUri, 0, nSize );
        memset(pCall, 0, sizeof(Call));
        pCall->nActualId = -1;
        strcat( pUri, "<sip:" );
        strcat( pUri, id );
        strcat( pUri, "@" );
        strcat( pUri, _pDestUri );
        strcat( pUri, ";transport=tcp>\0" );
        ErrorID nId = InitRtp(&pCall, _pConfig);
        if (nId != RET_OK) {
                DBG_ERROR("InitRtp failed %d \n", nId);
                free(pCall);
                return NULL;
        }
        int res = 0;
        res = createOffer(pCall->pPeerConnection);
        if (res != 0) {
                DBG_ERROR("createOffer failed %d \n", res);
                free(pCall);
                return NULL;
        }
        strncpy(pCall->url, pUri, MAX_URL_SIZE);
        pCall->id = ++CallId;
        *_pCallId = pCall->id;
        pCall->nAccountId = _nAccountId;
        pCall->callStatus = INV_STATE_CALLING;
        CheckCallStatus(pCall, CALL_STATUS_RING);
        free(pUri);
        return pCall;
}

ErrorID CALLAnswerCall(Call* _pCall)
{
        ErrorID id = CheckCallStatus(_pCall, CALL_STATUS_ESTABLISHED);
        if (id != RET_OK) {
              return id;
        }
        if (_pCall->nActualId == INVALID_CALL_ID) {
              return RET_CALL_NOT_EXIST;
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
        DBG_LOG("CALLAnswerCall  %p id %d \n", _pCall, _pCall->nActualId);
        return SipAnswerCall(_pCall->nActualId, OK, "answser call", _pCall->pLocal);
}

ErrorID CALLRejectCall(Call* _pCall)
{
        ErrorID id = CheckCallStatus(_pCall, CALL_STATUS_REJECT);
        if (id != RET_OK) {
              return id;
        }
        if (_pCall->nActualId == INVALID_CALL_ID) {
              return RET_CALL_NOT_EXIST;
        }
        id = SipAnswerCall(_pCall->nActualId, BUSY_HERE, "reject call", _pCall->pLocal);
        return id;
}

// hangup a call
ErrorID CALLHangupCall(Call* _pCall)
{
        ErrorID id = CheckCallStatus(_pCall, CALL_STATUS_HANGUP);
        if (id != RET_OK) {
              return id;
        }
        if (_pCall->nActualId == INVALID_CALL_ID) {
              return RET_CALL_NOT_EXIST;
        }
        SipHangUp(_pCall->nActualId);
        return id;
}

// send a packet
ErrorID CALLSendPacket(Call* _pCall, Stream streamID, const uint8_t* buffer, int size, int64_t nTimestamp)
{
        ErrorID id = CheckCallStatus(_pCall, CALL_STATUS_REGISTERED);
        if (id != RET_OK) {
              return id;
        }
        RtpStreamType type;
        if (streamID == STREAM_AUDIO) {
                type = RTP_STREAM_AUDIO;
        }
        else {
                type = RTP_STREAM_VIDEO;
        }
        RtpPacket packet = {(uint8_t*)(buffer), size, nTimestamp, type};
        return SendRtpPacket(_pCall->pPeerConnection, &packet);
}

SipAnswerCode CALLOnIncomingCall(Call** _pCall, const int _nAccountId, const int _nCallId, const char *pFrom,
                                 const void *pMedia, CallConfig* _pConfig)
{
        Call* pCall = (Call*)malloc(sizeof(Call));
        if (pCall == NULL) {
                *_pCall = NULL;
                return NOT_ACCEPTABLE;
        }
        memset(pCall, 0, sizeof(Call));
        *_pCall = pCall;
        pCall->id = ++CallId;
        pCall->nActualId = _nCallId;
        pCall->nAccountId = _nAccountId;
        pCall->callStatus =  INV_STATE_INCOMING;
        strncpy(pCall->from, pFrom, MAX_FROM_NAME_SIZE);
        pCall->from[MAX_FROM_NAME_SIZE - 1] = '\0';
        DBG_LOG("call %p CALLOnIncomingCall id %d\n", pCall, pCall->id);
        ErrorID id = InitRtp(&pCall, _pConfig);
        if (id != RET_OK) {
              DBG_ERROR("InitRtp failed %d\n", id);
              return INTERNAL_SERVER_ERROR;
        }
        pCall->pRemote = (pjmedia_sdp_session*)pMedia;
        int res = 0;
        DBG_LOG("pPeerConnection pRemote %p %p\n", pCall->pPeerConnection, pCall->pRemote);
        res = setRemoteDescription(pCall->pPeerConnection, pCall->pRemote);
        if (res != 0) {
                DBG_ERROR("setRemoteDescription failed %d \n", res);
                return INTERNAL_SERVER_ERROR;
        }
        res = createAnswer(pCall->pPeerConnection, pCall->pRemote);
        if (res != 0) {
                DBG_ERROR("createAnswer failed %d \n", res);
                return INTERNAL_SERVER_ERROR;
        }

        return OK;
}

void CALLOnCallStateChange(Call** _pCall, const SipInviteState State, const SipAnswerCode StatusCode, const void *pMedia)
{
        int res = 0;
        (*_pCall)->callStatus = State;
        if ((*_pCall)->callStatus == INV_STATE_CONNECTING) {
                DBG_LOG("====================stats %d state %d call %p\n", State, StatusCode, *_pCall);
                if (pMedia != NULL) {
                        res = setRemoteDescription((*_pCall)->pPeerConnection, (pjmedia_sdp_session*)(pMedia));
                }
        }
        else if ((*_pCall)->callStatus == INV_STATE_CONFIRMED) {
                if (res == 0) {
                        res = StartNegotiation((*_pCall)->pPeerConnection);
                }
                if (res != 0) {
                        DBG_ERROR("StartNegotiation failed %d %d todo\n", res, (*_pCall)->nActualId);
                        SipHangUp((*_pCall)->nActualId);
                        //SipAnswerCall((*_pCall)->nActualId, INTERNAL_SERVER_ERROR, "StartNegotiation failed", NULL);
                }
        }
        DBG_LOG("stats %d state %d call %p id %d aid %d \n", State, StatusCode, *_pCall, (*_pCall)->id, (*_pCall)->nActualId);
        if ((*_pCall)->callStatus == INV_STATE_DISCONNECTED) {
                //CALLHangupCall(*_pCall);
                if ((*_pCall)->pPeerConnection) {
                         ReleasePeerConnectoin((*_pCall)->pPeerConnection);
                         (*_pCall)->pPeerConnection = NULL;
                }
                free(*_pCall);
                DBG_LOG("Free call\n");
                *_pCall = NULL;
        }
        DBG_LOG("stats CALLOnCallStateChange end\n");
}

ErrorID CALLPollEvent(Call* _pCall, EventType* type, Event* event, int timeOut)
{
        //not used in current time.
        return RET_OK;
}
