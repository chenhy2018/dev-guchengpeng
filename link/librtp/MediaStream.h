#ifndef __MEDIASTREAM_H__
#define __MEDIASTREAM_H__

#include <pjsip.h>
#include <pjmedia.h>
#include <pjmedia-codec.h>
#include <pjlib-util.h>
#include <pjlib.h>
#include <pjmedia-codec/h264_packetizer.h>
#include "qrtc.h"


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

//TODO JitterBuffer add to MediaStreamTrack
typedef struct _JitterBuffer {
        pj_pool_t *pJbufPool;
        pjmedia_jbuf *pJbuf;
        pj_uint16_t nLastRecvRtpSeq;
}JitterBuffer;

typedef struct _MediaStreamTrack
{
        StreamType   type;
        pj_timestamp hzPerSecond;
        pj_timestamp nextRtcpTimestamp;
        pj_timestamp nSysTimeBase;
        uint64_t nLastPktTimestamp;
        uint64_t nFirstPktTimestamp;
        MediaConfig mediaConfig;
        pjmedia_rtp_session  rtpSession;
        pjmedia_rtcp_session rtcpSession;
        MediaPacketier *pMediaPacketier;
        pj_pool_t *pPacketizerPool;
        void *pPeerConnection;
}MediaStreamTrack;

typedef struct _MediaStream
{
        int nCount;
        MediaStreamTrack streamTracks[2]; //for audio and video
}MediaStream;

void InitMediaStream(IN MediaStream *pMediaStraem);
void AddMediaTrack(IN OUT MediaStream *pMediaStraem, IN MediaConfig *pMediaConfig, IN int nIndex, IN StreamType type,
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
#endif
