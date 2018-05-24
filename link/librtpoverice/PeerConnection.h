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

#include "MediaStream.h"

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

    void (*onIceComplete)(IN pjmedia_transport *pTransport, IN pj_ice_strans_op op,
                          IN pj_status_t status);
    void (*onIceComplete2)(IN pjmedia_transport *pTransport, IN pj_ice_strans_op op,
                          IN pj_status_t status, void * pUserData);
    
}IceConfig;

typedef struct _TransportIce
{
    pjmedia_transport *pTransport;
    pj_ioqueue_t      *pIoQueue;
    pj_pool_t         *pIoqueuePool;
    pj_timer_heap_t   *pTimerHeap;
    pj_pool_t         *pTimerHeapPool;
    pj_thread_t       *pPollThread;
    pj_pool_t         *pThreadPool;
}TransportIce;

typedef enum _IceState{
    ICE_STATE_INIT,
    ICE_STATE_GATHERING_OK,
    ICE_STATE_NEGOTIATION_OK,
    ICE_STATE_FAIL,
}IceState;

typedef struct _PeerConnectoin
{
    IceConfig         userIceConfig;
    pj_ice_strans_cfg iceConfig;
    TransportIce      transportIce[2]; //audio and video
    int               nAvIndex[2];
    pj_pool_factory   *pPoolFactory;
    pjmedia_endpt     *pMediaEndpt;
    MediaStream       mediaStream;
    pjmedia_sdp_session *pOfferSdp;
    pjmedia_sdp_session *pAnswerSdp;
    int bQuit;
    IceState iceState;
}PeerConnectoin;

void InitIceConfig(IN OUT IceConfig *pIceConfig);

int InitPeerConnectoin(IN OUT PeerConnectoin * pPeerConnectoin, IN IceConfig *pIceConfig,
                       IN pj_pool_factory * pPoolFactory);

int AddAudioTrack(IN OUT PeerConnectoin * pPeerConnection, IN MediaConfig *pAudioConfig);
int AddVideoTrack(IN OUT PeerConnectoin * pPeerConnection, IN MediaConfig *pVideoConfig);
int createOffer(IN OUT PeerConnectoin * pPeerConnection, IN pj_pool_t * pPool, OUT pjmedia_sdp_session **pOffer);
int createAnswer(IN OUT PeerConnectoin * pPeerConnection, IN pj_pool_t * pPool,
                 IN pjmedia_sdp_session *pOffer, OUT pjmedia_sdp_session **pAnswer);
void setLocalDescription(IN OUT PeerConnectoin * pPeerConnection, IN pjmedia_sdp_session * pLocalSdp);
void setRemoteDescription(IN OUT PeerConnectoin * pPeerConnection, IN pjmedia_sdp_session * pRemoteSdp);

void ReleasePeerConnectoin(IN OUT PeerConnectoin * _pPeerConnection);




#endif
