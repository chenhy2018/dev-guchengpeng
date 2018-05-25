#ifndef __MEDIASTREAM_H__
#define __MEDIASTREAM_H__

#include <pjsip.h>
#include <pjmedia.h>
#include <pjmedia-codec.h>
#include <pjlib-util.h>
#include <pjlib.h>

#define _s_l_(x) #x
#define _str_line_(x) _s_l_(x)
#define __STR_LINE__ _str_line_(__LINE__)

#define THIS_FILE "PeerConnection.c"
#define ASSERT_RETURN_CHECK(expr, errmsg) if ( (expr) == NULL ) { \
        PJ_LOG(3,(__FILE__, "%s", errmsg));;\
        return -1;}
#if 0
#define STATUS_CHECK(info, s) if ( s != PJ_SUCCESS) { \
        status_perror(__STR_LINE__, #info, s);\
        return s;}
#endif
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

typedef enum _MediaFormat
{
    MEDIA_FORMAT_PCMU,
    MEDIA_FORMAT_H264,
}MediaFromat;

typedef enum _MediaType
{
    TYPE_AUDIO,
    TYPE_VIDEO,
}MediaType;

typedef union _MediaConfig{
    int nRtpDynamicType;
    MediaFromat format;
    int nSampleOrClockRate;
    int nChannel;
    int nBitDepth;
}MediaConfig;

typedef struct _MediaStreamTrack
{
    MediaType   type;
    pj_timestamp hzPerSecond;
    pj_timestamp nextRtpTimestamp;
    pj_timestamp nextRtcpTimestamp;
    MediaConfig mediaConfig;
    pjmedia_rtp_session  rtpSession;
    pjmedia_rtcp_session rtcpSession;
}MediaStreamTrack;

typedef struct _MediaStream
{
    int nCount;
    MediaStreamTrack streamTracks[2]; //for audio and video
}MediaStream;

void InitMediaConfig(IN MediaConfig * pMediaConfig);
void InitMediaStream(IN MediaStream *pMediaStraem);
void AddMediaTrack(IN OUT MediaStream *pMediaStraem, IN MediaConfig *pMediaConfig, IN int nIndex, IN MediaType type);
int CreateSdpAudioMLine(IN pjmedia_endpt *pMediaEndpt, IN pjmedia_transport_info *pTransportInfo,
                        IN pj_pool_t * pPool, IN MediaStreamTrack *pMediaTrack, OUT pjmedia_sdp_media ** pAudioSdp);
int CreateSdpVideoMLine(IN pjmedia_endpt *pMediaEndpt, IN pjmedia_transport_info *pTransportInfo,
                        IN pj_pool_t * pPool, IN MediaStreamTrack *pMediaTrack, OUT pjmedia_sdp_media ** pVideoSdp);

MediaStreamTrack * GetAudioTrack(IN MediaStream * pMediaStream);
MediaStreamTrack * GetVideoTrack(IN MediaStream * pMediaStream);
int GetMediaTrackIndex(IN MediaStream * pMediaStream, IN MediaStreamTrack *pMediaStreamTrack);
#endif
