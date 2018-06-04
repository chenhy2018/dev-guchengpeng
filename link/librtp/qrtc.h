#ifndef __QRTC_H__
#define __QRTC_H__
#ifndef __APPLE__
#include <inttypes.h>
#endif

#define _s_l_(x) #x
#define _str_line_(x) _s_l_(x)
#define __STR_LINE__ _str_line_(__LINE__)

#define ASSERT_RETURN_CHECK(expr, errmsg) if ( (expr) == NULL ) { \
PJ_LOG(3,(__FILE__, "%s", errmsg));;\
return -1;}

#define STATUS_CHECK(info, s) if ( s != PJ_SUCCESS) { \
char errmsg[PJ_ERR_MSG_SIZE];\
pj_strerror(s, errmsg, sizeof(errmsg));\
PJ_LOG(3,(__FILE__, "%s: %s [code=%d]", #info, errmsg, s));\
return s;}

#ifndef IN
#define IN
#endif
#ifndef OUT
#define OUT
#endif

typedef enum _MediaType
{
        TYPE_AUDIO,
        TYPE_VIDEO,
}MediaType;

typedef enum _MediaFormat
{
        MEDIA_FORMAT_PCMU = 0,
        MEDIA_FORMAT_PCMA = 8,
        MEDIA_FORMAT_G729 = 18,
        MEDIA_FORMAT_H264 = 96,
        MEDIA_FORMAT_H265 = 98,
}MediaFromat;

#define MAX_CODEC_LEN 5
typedef struct _AvParam{
        int nRtpDynamicType;
        MediaFromat format;
        int nSampleOrClockRate;
        int nChannel;
        int nBitDepth;
}AvParam;
typedef struct _MediaConfig{
        AvParam configs[MAX_CODEC_LEN];
        int nCount;
        int nUseIndex;
}MediaConfig;

typedef enum _CallbackType{
        CALLBACK_ICE,
        CALLBACK_RTP,
        CALLBACK_RTCP
}CallbackType;

typedef void(*UserCallback)(void *pUserData, CallbackType type, void *pCbData);

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
        
        UserCallback userCallback;
        void *       pCbUserData;
        
}IceConfig;

typedef enum _IceState{
        ICE_STATE_INIT,
        ICE_STATE_GATHERING_OK,
        ICE_STATE_NEGOTIATION_OK,
        ICE_STATE_FAIL,
}IceState;

typedef struct _RtpPacket{
        uint8_t * pData;
        int nDataLen;
        uint64_t nTimestamp;
        MediaType type;
        MediaFromat format;
}RtpPacket;

typedef struct _IceNegInfo {
        IceState state;
        const AvParam* configs[2];
        int nCount;
}IceNegInfo;


void InitMediaConfig(IN MediaConfig * pMediaConfig);
void InitIceConfig(IN OUT IceConfig *pIceConfig);

typedef struct _PeerConnection PeerConnection;
int InitPeerConnectoin(IN OUT PeerConnection ** pPeerConnectoin,
                        IN IceConfig *pIceConfig);

int AddAudioTrack(IN OUT PeerConnection * pPeerConnection, IN MediaConfig *pAudioConfig);
int AddVideoTrack(IN OUT PeerConnection * pPeerConnection, IN MediaConfig *pVideoConfig);
int createOffer(IN OUT PeerConnection * pPeerConnection,  OUT pjmedia_sdp_session **pOffer);
int createAnswer(IN OUT PeerConnection * pPeerConnection, IN pjmedia_sdp_session *pOffer,
                 OUT pjmedia_sdp_session **pAnswer);
int setLocalDescription(IN OUT PeerConnection * pPeerConnection, IN pjmedia_sdp_session * pSdp);
int setRemoteDescription(IN OUT PeerConnection * pPeerConnection, IN pjmedia_sdp_session * pSdp);
int StartNegotiation(IN PeerConnection * pPeerConnection);

int SendPacket(IN PeerConnection *pPeerConnection, IN OUT RtpPacket * pPacket);

void ReleasePeerConnectoin(IN OUT PeerConnection * _pPeerConnection);

#endif
