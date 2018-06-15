#ifndef __QRTC_H__
#define __QRTC_H__
#include <stdint.h>

#define _s_l_(x) #x
#define _str_line_(x) _s_l_(x)
#define __STR_LINE__ _str_line_(__LINE__)

#define MY_PJ_LOG(level,fmt,...) \
if(level <= pj_log_get_level()) pj_log_wrapper_##level((THIS_FILE ":" __STR_LINE__ ": ", fmt, ##__VA_ARGS__))


#define ASSERT_RETURN_CHECK(expr, errmsg) if ( (expr) == NULL ) { \
PJ_LOG(1,(__FILE__, "%s", errmsg));;\
return -1;}

#define STATUS_CHECK(info, s) if ( s != PJ_SUCCESS) { \
char errmsg[PJ_ERR_MSG_SIZE];\
pj_strerror(s, errmsg, sizeof(errmsg));\
PJ_LOG(1,(__FILE__, "%s: %s [code=%d]", #info, errmsg, s));\
return s;}

#ifndef IN
#define IN
#endif
#ifndef OUT
#define OUT
#endif

typedef enum _RtpStreamType
{
        RTP_STREAM_AUDIO,
        RTP_STREAM_VIDEO,
        RTP_STREAM_DATA
}RtpStreamType;

typedef enum _CodecType
{
        MEDIA_FORMAT_PCMU = 0,
        MEDIA_FORMAT_PCMA = 8,
        MEDIA_FORMAT_G729 = 18,
        MEDIA_FORMAT_H264 = 96,
        MEDIA_FORMAT_H265 = 98,
}CodecType;

#define MAX_CODEC_LEN 5
typedef struct _MediaConfig{
        RtpStreamType streamType;
        CodecType codecType; //also rtp type
        int nSampleOrClockRate;
        int nChannel;
}MediaConfig;

typedef struct _MediaConfigSet{
        MediaConfig configs[MAX_CODEC_LEN];
        int nCount;
        int nUseIndex;
}MediaConfigSet;

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
        
        int          nForceStun;
        UserCallback userCallback;
        void *       pCbUserData;
        
}IceConfig;

typedef enum _IceState{
        ICE_STATE_INIT,
        ICE_STATE_FAIL,
        ICE_STATE_GATHERING_FAIL,
        ICE_STATE_NEGOTIATION_FAIL,
        ICE_STATE_GATHERING_OK,
        ICE_STATE_NEGOTIATION_OK,
}IceState;

typedef struct _RtpPacket{
        uint8_t * pData;
        int nDataLen;
        uint64_t nTimestamp;
        RtpStreamType type;
        CodecType format;
}RtpPacket;

typedef struct _IceNegInfo {
        IceState state;
        const MediaConfig* configs[2];
        int nCount;
        void *pData;
}IceNegInfo;


void InitMediaConfig(IN MediaConfigSet * pMediaConfig);
void InitIceConfig(IN OUT IceConfig *pIceConfig);

typedef struct _PeerConnection PeerConnection;
int InitPeerConnectoin(OUT PeerConnection ** pPeerConnectoin,
                        IN IceConfig *pIceConfig);

int AddAudioTrack(IN OUT PeerConnection * pPeerConnection, IN MediaConfigSet *pAudioConfig);
int AddVideoTrack(IN OUT PeerConnection * pPeerConnection, IN MediaConfigSet *pVideoConfig);
int createOffer(IN OUT PeerConnection * pPeerConnection);
int createAnswer(IN OUT PeerConnection * pPeerConnection, IN void *pOffer);
int setRemoteDescription(IN OUT PeerConnection * pPeerConnection, IN void * pSdp);
int StartNegotiation(IN PeerConnection * pPeerConnection);

int SendRtpPacket(IN PeerConnection *pPeerConnection, IN OUT RtpPacket * pPacket);

int ReleasePeerConnectoin(IN OUT PeerConnection * _pPeerConnection);

#endif
