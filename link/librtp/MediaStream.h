#ifndef __MEDIASTREAM_H__
#define __MEDIASTREAM_H__

#include <pjsip.h>
#include <pjmedia.h>
#include <pjmedia-codec.h>
#include <pjlib-util.h>
#include <pjlib.h>
#include <pjmedia-codec/h264_packetizer.h>
#include "JitterBuffer.h"

#define ULLONG_MAX 18446744073709551615ul
#define TS_BASE_VALUE 40

typedef struct _MediaPacketier MediaPacketier;
typedef struct _PacketierOperation {
        pj_status_t (*packetize)(IN OUT MediaPacketier *pKtz,
                                 IN pj_uint8_t *pBitstream,
                                 IN pj_size_t nBitstreamLen,
                                 IN unsigned *pBitstreamPos,
                                 OUT const pj_uint8_t **pPayload,
                                 OUT pj_size_t *nPlyloadLen);

        pj_status_t (*unpacketize)(IN OUT MediaPacketier *pKtz,
                                   IN const pj_uint8_t *pPayload,
                                   IN pj_size_t   nPlyloadLen,
                                   OUT pj_uint8_t **pBitstream,
                                   OUT unsigned   *pBitstreamPos,
                                   IN int nRtpMarker, OUT pj_bool_t *pTryAgain);

}PacketierOperation;

typedef struct _MediaPacketier {
        PacketierOperation pOperation;
}MediaPacketier;

typedef struct _PcmuPacketizer {
        MediaPacketier base;
        pj_pool_t *pPcmuPacketizerPool;
}PcmuPacketizer;

typedef struct _H264Packetizer {
        MediaPacketier base;
        pj_pool_t *pH264PacketizerPool;
        pjmedia_h264_packetizer *pH264Packetizer;
        uint8_t *pUnpackBuf;
        unsigned nUnpackBufCap;
        unsigned nUnpackBufLen;
        pj_bool_t bFuAStartbit;
}H264Packetizer;

typedef struct _MediaStreamTrack
{
        RtpStreamType   type;
        pj_timestamp hzPerSecond;
        pj_timestamp nextRtcpTimestamp;
        pj_timestamp nSysTimeBase;
        uint64_t nLastSendPktTimestamp;
        uint64_t nFirstSendPktTimestamp;
        uint32_t nLastRecvPktTimestamp;
        uint64_t nMostLastRecvTimeAcc;
        MediaConfigSet mediaConfig;
        pjmedia_rtp_session  rtpSession;
        pjmedia_rtcp_session rtcpSession;
        MediaPacketier *pMediaPacketier;
        pj_pool_t *pPacketizerPool;
        void *pPeerConnection;
        JitterBuffer jbuf;
}MediaStreamTrack;

typedef struct _MediaStream
{
        int nCount;
        MediaStreamTrack streamTracks[2]; //for audio and video
}MediaStream;

void InitMediaStream(IN MediaStream *pMediaStraem);
void AddMediaTrack(IN OUT MediaStream *pMediaStraem, IN MediaConfigSet *pMediaConfig, IN int nIndex, IN RtpStreamType type,
                   IN void * pPeerConnection);
int CreateSdpAudioMLine(IN pjmedia_endpt *pMediaEndpt, IN pjmedia_transport_info *pTransportInfo,
                        IN pj_pool_t * pPool, IN MediaStreamTrack *pMediaTrack, OUT pjmedia_sdp_media ** pAudioSdp);
int CreateSdpVideoMLine(IN pjmedia_endpt *pMediaEndpt, IN pjmedia_transport_info *pTransportInfo,
                        IN pj_pool_t * pPool, IN MediaStreamTrack *pMediaTrack, OUT pjmedia_sdp_media ** pVideoSdp);
int SetActiveCodec( IN OUT MediaStream *pMediaStream, IN const pjmedia_sdp_session *pActiveSdp);
MediaStreamTrack * GetAudioTrack(IN MediaStream * pMediaStream);
MediaStreamTrack * GetVideoTrack(IN MediaStream * pMediaStream);
int GetMediaTrackIndex(IN MediaStream * pMediaStream, IN MediaStreamTrack *pMediaStreamTrack);

pj_status_t CreatePacketizer(IN char *pName, IN int nNameLen, IN pj_pool_t *pPktzPool, OUT MediaPacketier **pPktz);
pj_status_t MediaPacketize(IN MediaPacketier *pPktz,IN pj_uint8_t *pBitstream, IN pj_size_t nBitstreamLen,
                           IN unsigned *pBitstreamPos, OUT const pj_uint8_t **pPayload, OUT pj_size_t *nPlyloadLen);
pj_status_t MediaUnPacketize(IN OUT MediaPacketier *pKtz, IN const pj_uint8_t *pPayload, IN pj_size_t nPlyloadLen,
                           OUT pj_uint8_t **pBitstream, OUT unsigned *pBitstreamPos, IN int nRtpMarker, IN pj_bool_t *pTryAgain);
pj_status_t createJitterBuffer(IN MediaStreamTrack *pMediaStreamTrack, IN pj_pool_factory *pPoolFactory);
pj_status_t MediaConfigSetIsValid(MediaConfigSet *pConfig);
void DestroyMediaStream(IN MediaStream *pMediaStraem);
#endif
