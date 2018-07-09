#ifndef __PEERCONNECTION_H__
#define __PEERCONNECTION_H__

#include <pjsip.h>
#include <pjmedia.h>
#include <pjmedia-codec.h>
#include <pjlib-util.h>
#include <pjlib.h>

#include "MediaStream.h"
#include "qrtc.h"
#include "../util/queue.h"


typedef enum _IceRole{
        ICE_ROLE_NONE,
        ICE_ROLE_OFFERER,
        ICE_ROLE_ANSWERER
}IceRole;

typedef struct _TransportIce
{
        pjmedia_transport *pTransport;
        pj_pool_t         *pNegotiationPool;
        IceState iceState;
        pj_ice_strans_cfg iceConfig;
        void *pPeerConnection;
}TransportIce;

#define PC_STATUS_ALLOC 0
#define PC_STATUS_INIT_OK 1
#define PC_STATUS_CREATE_OFFER_OK 2
#define PC_STATUS_CREATE_ANSWER_OK 3
#define PC_STATUS_SET_REMOTE_OK 4
#define PC_STATUS_NEG_OK 5
#define PC_STATUS_INIT_FAIL 11
#define PC_STATUS_CREATE_OFFER_FAIL 12
#define PC_STATUS_CREATE_ANSWER_FAIL 13
#define PC_STATUS_SET_REMOTE_FAIL 14
#define PC_STATUS_NEG_FAIL 15

typedef struct _PeerConnection
{
        int nState;
        IceConfig         userIceConfig;
        TransportIce      transportIce[2]; //audio and video
        int               nAvIndex[2];
        MediaStream       mediaStream;
        pjmedia_sdp_session *pLocalSdp;
        pjmedia_sdp_session *pRemoteSdp;
        pjmedia_sdp_session *pOfferSdp; //if invoke createOffer same as pLocalSdp
        pjmedia_sdp_session *pAnswerSdp; //if invoke createAnswer same as pRemoteSdp
        pjmedia_sdp_neg *pIceNeg;
        pj_pool_t *pNegPool;
        IceNegInfo iceNegInfo;
        pj_pool_t *pSdpPool;
        int nNegSuccess;
        int nGatherCandidateSuccessCount;
        int nIsFailCallbackDone;
        pj_mutex_t *pMutex;
        pj_pool_t *pMutexPool;
        int nQuitCnt;
        IceRole role;
        int nDestroy;
        pj_pool_t *pGrpPool;
        pj_grp_lock_t *pGrpLock1;
        pj_grp_lock_t *pGrpLock2;
        char * pRemoteSdpStr;
        int nSdpStrLen;
}PeerConnection;


#define MQ_TYPE_SEND 2
#define MQ_TYPE_CREATE_OFFER 3
#define MQ_TYPE_CREATE_ANSWER 4
#define MQ_TYPE_NEG 7
#define MQ_TYPE_RELEASE 8

typedef struct _RtpMqMsg{
        Message msg;
        int nType;
        PeerConnection *pPeerConnection;
        RtpPacket pkt;
        void * pArg;
}RtpMqMsg;

#endif
