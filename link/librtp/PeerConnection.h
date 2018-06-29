#ifndef __PEERCONNECTION_H__
#define __PEERCONNECTION_H__

#include <pjsip.h>
#include <pjmedia.h>
#include <pjmedia-codec.h>
#include <pjlib-util.h>
#include <pjlib.h>

#include "MediaStream.h"
#include "qrtc.h"


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

typedef struct _PeerConnection
{
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
        int bQuit;
        IceRole role;
}PeerConnection;

#endif
