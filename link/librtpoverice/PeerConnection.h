#ifndef __PEERCONNECTION_H__
#define __PEERCONNECTION_H__

#ifndef IN
#define IN
#endif
#ifndef OUT
#define OUT
#endif

#include <pjsip.h>
#include <pjmedia.h>
#include <pjmedia-codec.h>
#include <pjlib-util.h>
#include <pjlib.h>
#ifndef __APPLE__
#include <inttypes.h>
#endif

#include "MediaStream.h"

typedef void(*RtpReceiveCallback)(void *user_data, void *pkt, pj_ssize_t);
typedef void(*RtcpReceiveCallback)(void *usr_data, void*pkt, pj_ssize_t);

#define MAX_NAMESERVER_SIZE 128
#define MAX_STUN_HOST_SIZE  128
#define MAX_TURN_HOST_SIZE  128
#define MAX_TURN_USR_SIZE   64
#define MAX_TURN_PWD_SIZE   64
#define MAX_ICE_USRPWD_SIZE 80

typedef struct _IceConfig
{
    unsigned     nComponents;
    char         nameserver[MAX_NAMESERVER_SIZE];
    int          bRegular;
    int          nKeepAlive;
    //stun
    int          nMaxHosts;
    char         stunHost[MAX_STUN_HOST_SIZE];
    //turn
    int          bTurnTcp;
    char         turnHost[MAX_TURN_HOST_SIZE];
    char         turnUsername[MAX_TURN_USR_SIZE];
    char         turnPassword[MAX_TURN_PWD_SIZE];

    //TODO not use now
    void (*onIceComplete)(IN pjmedia_transport *pTransport, IN pj_ice_strans_op op,
                          IN pj_status_t status);
    void (*onIceComplete2)(IN pjmedia_transport *pTransport, IN pj_ice_strans_op op,
                          IN pj_status_t status, void * pUserData);
    
}IceConfig;

typedef enum _IceState{
    ICE_STATE_INIT,
    ICE_STATE_GATHERING_OK,
    ICE_STATE_NEGOTIATION_OK,
    ICE_STATE_FAIL,
}IceState;

typedef enum _IceRole{
    ICE_ROLE_NONE,
    ICE_ROLE_OFFERER,
    ICE_ROLE_ANSWERER
}IceRole;

typedef struct _TransportIce
{
    pjmedia_transport *pTransport;
    pj_ioqueue_t      *pIoQueue;
    pj_pool_t         *pIoqueuePool;
    pj_timer_heap_t   *pTimerHeap;
    pj_pool_t         *pTimerHeapPool;
    pj_thread_t       *pPollThread;
    pj_pool_t         *pThreadPool;
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
    pj_pool_factory   *pPoolFactory;
    pjmedia_endpt     *pMediaEndpt;
    MediaStream       mediaStream;
    pjmedia_sdp_session *pOfferSdp;
    pjmedia_sdp_session *pAnswerSdp;
    pjmedia_sdp_neg *pIceNeg;
    pj_pool_t *pNegPool;
    int bQuit;
    IceRole role;
}PeerConnection;

typedef struct _RtpPacket{
    uint8_t * pData;
    int nDataLen;
    int64_t nTimestamp;
    MediaType type;
}RtpPacket;

void InitIceConfig(IN OUT IceConfig *pIceConfig);

void InitPeerConnectoin(IN OUT PeerConnection * pPeerConnectoin, IN pj_pool_factory * pPoolFactory,
                        IN IceConfig *pIceConfig);

int AddAudioTrack(IN OUT PeerConnection * pPeerConnection, IN MediaConfig *pAudioConfig);
int AddVideoTrack(IN OUT PeerConnection * pPeerConnection, IN MediaConfig *pVideoConfig);
int createOffer(IN OUT PeerConnection * pPeerConnection, IN pj_pool_t * pPool, OUT pjmedia_sdp_session **pOffer);
int createAnswer(IN OUT PeerConnection * pPeerConnection, IN pj_pool_t * pPool,
                 IN pjmedia_sdp_session *pOffer, OUT pjmedia_sdp_session **pAnswer);
int setLocalDescription(IN OUT PeerConnection * pPeerConnection, IN pjmedia_sdp_session * pSdp);
int setRemoteDescription(IN OUT PeerConnection * pPeerConnection, IN pjmedia_sdp_session * pSdp);
int StartNegotiation(IN PeerConnection * pPeerConnection);

int SendPacket(IN PeerConnection *pPeerConnection, IN RtpPacket * pPacket);

void ReleasePeerConnectoin(IN OUT PeerConnection * _pPeerConnection);


#endif
