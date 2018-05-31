#include "MediaStream.h"

void InitMediaConfig(IN MediaConfig * pMediaConfig)
{
        pj_bzero(pMediaConfig, sizeof(MediaConfig));
}

void InitMediaStream(IN MediaStream *_pMediaStraem)
{
        pj_bzero(_pMediaStraem, sizeof(MediaStream));
}

static void setMediaConfig(IN OUT MediaConfig *_pMediaConfig)
{
        for ( int i = 0; i < _pMediaConfig->nCount; i++) {
                switch (_pMediaConfig->configs[i].format) {
                        case MEDIA_FORMAT_PCMU:
                        case MEDIA_FORMAT_PCMA:
                        case MEDIA_FORMAT_G729:
                                _pMediaConfig->configs[i].nRtpDynamicType = _pMediaConfig->configs[i].format;
                                _pMediaConfig->configs[i].nChannel = 1;
                                _pMediaConfig->configs[i].nBitDepth = 8;
                                break;
                        case MEDIA_FORMAT_H264:
                        case MEDIA_FORMAT_H265:
                                break;
                }
        }
}

void AddMediaTrack(IN OUT MediaStream *_pMediaStraem, IN MediaConfig *_pMediaConfig, IN int _nIndex, IN MediaType _type,
                   IN void * _pPeerConnection)
{
        pj_assert(_pMediaStraem && _pMediaConfig);
        
        for (int i = 0; i < _pMediaStraem->nCount; i++) {
                if ( _pMediaStraem->streamTracks[i].type == _type ){
                        PJ_LOG(3, (__FILE__, "media type exists"));
                        return;
                }
        }
        
        _pMediaStraem->nCount++;
        
        _pMediaStraem->streamTracks[_nIndex].type = _type;
        _pMediaStraem->streamTracks[_nIndex].mediaConfig = *_pMediaConfig;
        _pMediaStraem->streamTracks[_nIndex].pPeerConnection = _pPeerConnection;
        
        setMediaConfig(&_pMediaStraem->streamTracks[_nIndex].mediaConfig);
}

int CreateSdpAudioMLine(IN pjmedia_endpt *_pMediaEndpt, IN pjmedia_transport_info *_pTransportInfo,
                        IN pj_pool_t * _pPool, IN MediaStreamTrack *_pMediaTrack, OUT pjmedia_sdp_media ** _pAudioSdp)
{
        pj_assert(_pMediaTrack->type == TYPE_AUDIO);
        
        pj_status_t status;
        status = pjmedia_endpt_create_audio_sdp(_pMediaEndpt, _pPool, &_pTransportInfo->sock_info, 0, _pAudioSdp);
        STATUS_CHECK(pjmedia_endpt_create_audio_sdp, status);
        
        pj_str_t * fmt;
        for ( int i = 0; i < _pMediaTrack->mediaConfig.nCount; i++) {
                switch (_pMediaTrack->mediaConfig.configs[i].format) {
                        case MEDIA_FORMAT_PCMU:
                        case MEDIA_FORMAT_PCMA:
                        case MEDIA_FORMAT_G729:
                                fmt = &((*_pAudioSdp)->desc.fmt[(*_pAudioSdp)->desc.fmt_count++]);
                                fmt->ptr = pj_pool_alloc(_pPool, 4);
                                fmt->slen = snprintf(fmt->ptr, 4, "%d", _pMediaTrack->mediaConfig.configs[i].nRtpDynamicType);
                                break;
                        case MEDIA_FORMAT_H264:
                        case MEDIA_FORMAT_H265:
                                break;
                }
        }
        
        return PJ_SUCCESS;
}

int CreateSdpVideoMLine(IN pjmedia_endpt *_pMediaEndpt, IN pjmedia_transport_info *_pTransportInfo,
                        IN pj_pool_t * _pPool, IN MediaStreamTrack *_pMediaTrack, OUT pjmedia_sdp_media ** _pVideoSdp)
{
        pj_assert(_pMediaTrack->type == TYPE_VIDEO);
        
        pj_status_t status;
        status = pjmedia_endpt_create_video_sdp(_pMediaEndpt, _pPool, &_pTransportInfo->sock_info, 0, _pVideoSdp);
        STATUS_CHECK(pjmedia_endpt_create_audio_sdp, status);
        
        pj_str_t * fmt;
        for ( int i = 0; i < _pMediaTrack->mediaConfig.nCount; i++) {
                switch (_pMediaTrack->mediaConfig.configs[i].format) {
                        case MEDIA_FORMAT_PCMU:
                        case MEDIA_FORMAT_PCMA:
                        case MEDIA_FORMAT_G729:
                                break;
                        case MEDIA_FORMAT_H264:
                        case MEDIA_FORMAT_H265:
                                fmt = &((*_pVideoSdp)->desc.fmt[(*_pVideoSdp)->desc.fmt_count++]);
                                fmt->ptr = pj_pool_alloc(_pPool, 4);
                                fmt->slen = snprintf(fmt->ptr, 4, "%d", _pMediaTrack->mediaConfig.configs[i].nRtpDynamicType);
                                pjmedia_sdp_attr *pAttr = NULL;
                                pjmedia_sdp_rtpmap rtpmap;
                                pj_bzero(&rtpmap, sizeof(rtpmap));
                                rtpmap.pt = *fmt;
                                rtpmap.clock_rate = _pMediaTrack->mediaConfig.configs[i].nSampleOrClockRate;
                                if (_pMediaTrack->mediaConfig.configs[i].format == MEDIA_FORMAT_H265) {
                                        rtpmap.enc_name = pj_str("H265");
                                } else {
                                        rtpmap.enc_name = pj_str("H264");
                                }
                                pjmedia_sdp_rtpmap_to_attr(_pPool, &rtpmap, &pAttr);
                                (*_pVideoSdp)->attr[(*_pVideoSdp)->attr_count++] = pAttr;
                                break;
                }
        }
        
        return PJ_SUCCESS;
}

static inline MediaStreamTrack * GetTrackByType(IN MediaStream * _pMediaStream, MediaType _type)
{
        for (int i = 0; i < sizeof(_pMediaStream->streamTracks) / sizeof(MediaStreamTrack); i++) {
                if (_pMediaStream->streamTracks[i].type == _type) {
                        return &_pMediaStream->streamTracks[i];
                }
        }
        return NULL;
}

MediaStreamTrack * GetAudioTrack(IN MediaStream * _pMediaStream)
{
        return GetTrackByType(_pMediaStream, TYPE_AUDIO);
}

MediaStreamTrack * GetVideoTrack(IN MediaStream * _pMediaStream)
{
        return GetTrackByType(_pMediaStream, TYPE_VIDEO);
}

int GetMediaTrackIndex(IN MediaStream * _pMediaStream, IN MediaStreamTrack *_pMediaStreamTrack)
{
        for (int i = 0; i < sizeof(_pMediaStream->streamTracks) / sizeof(MediaStreamTrack); i++) {
                if (&_pMediaStream->streamTracks[i] == _pMediaStreamTrack) {
                        return i;
                }
        }
        
        return -1;
}

static int setActiveCodecConfig(IN OUT MediaStreamTrack *_pMediaStreamTrack, int _nActivePt)
{
        if (_nActivePt < 0) {
                return -1;
        }
        int nCount = _pMediaStreamTrack->mediaConfig.nCount;
        for (int i = 0; i < nCount; i++) {
                if(_nActivePt == _pMediaStreamTrack->mediaConfig.configs[i].nRtpDynamicType){
                        _pMediaStreamTrack->mediaConfig.nUseIndex = i;
                        return PJ_SUCCESS;
                }
        }
        return -2;
}

int SetActiveCodec(IN OUT MediaStream *_pMediaStream, IN const pjmedia_sdp_session *_pActiveSdp)
{
        int nPt = -1;
        pj_status_t status;
        for ( int i = 0; i < _pActiveSdp->media_count; i++) {
                nPt = atoi(_pActiveSdp->media[i]->desc.fmt[0].ptr);
                status = setActiveCodecConfig(&_pMediaStream->streamTracks[i], nPt);
                STATUS_CHECK(setActiveCodecConfig, status);
        }
        
        return PJ_SUCCESS;
}


//packetizer
static pj_str_t pcmuPktzName = {"pcmu", 4};
static pj_str_t pcmaPktzName = {"pcma", 4};
static pj_str_t aacPktzName  = {"aac", 3};
static pj_str_t h264PktzName = {"h264", 4};
static pj_str_t h265PktzName = {"h265", 4};

pj_status_t pcmu_packetize(IN MediaPacketier *_pKtz,
                           IN pj_uint8_t *_pBitstream,
                           IN pj_size_t _nBitstreamLen,
                           IN unsigned *_pBitstreamPos,
                           OUT const pj_uint8_t **_pPayload,
                           OUT pj_size_t *_nPayloadLen)
{
        pj_assert(_nBitstreamLen >= *_pBitstreamPos);

        *_pPayload = _pBitstream + *_pBitstreamPos;
        if (_nBitstreamLen - *_pBitstreamPos <= 1300){
                *_nPayloadLen = _nBitstreamLen - *_pBitstreamPos;
        } else {
                *_nPayloadLen = 1300;
                *_pBitstreamPos = *_pBitstreamPos + 1300;
        }

        return PJ_SUCCESS;
}

pj_status_t pcmu_unpacketize(IN OUT MediaPacketier *_pKtz,
                             IN const pj_uint8_t *_pPayload,
                             IN pj_size_t   _nPlyloadLen,
                             OUT pj_uint8_t **_pBitstream,
                             OUT unsigned   *_pBitstreamPos,
                             IN int _nRtpMarker)
{
        *_pBitstream = (pj_uint8_t *)_pPayload;
        *_pBitstreamPos = _nPlyloadLen;
        return PJ_SUCCESS;
}

pj_status_t h264_packetize(IN MediaPacketier *_pKtz,
                           IN pj_uint8_t *_pBitstream,
                           IN pj_size_t _nBitstreamLen,
                           IN unsigned *_pBitstreamPos,
                           OUT const pj_uint8_t **_pPayload,
                           OUT pj_size_t *_nPlyloadLen)
{
        H264Packetizer *pPktz = (H264Packetizer *)_pKtz;
        pj_status_t status;
        status = pjmedia_h264_packetize(pPktz->pH264Packetizer, _pBitstream, _nBitstreamLen,
                                        _pBitstreamPos, _pPayload, _nPlyloadLen);

        return status;
}

pj_status_t h264_unpacketize(IN OUT MediaPacketier *_pKtz,
                             IN const pj_uint8_t *_pPayload,
                             IN pj_size_t   _nPlyloadLen,
                             OUT pj_uint8_t **_pBitstream,
                             OUT unsigned   *_pBitstreamPos,
                             IN int _nRtpMarker)
{
        H264Packetizer *pPktz = (H264Packetizer *)_pKtz;

        pj_status_t status = PJ_SUCCESS;

        if (pPktz->pUnpackBuf == NULL) {
                pj_pool_alloc(pPktz->pH264PacketizerPool, 100*1024);
                pPktz->nUnpackBufCap = 100*1024;
                pPktz->nUnpackBufLen = 0;
        }
        if (pPktz->bShouldReset) {
                pPktz->nUnpackBufLen = 0;
                pPktz->bShouldReset = PJ_FALSE;
                _pBitstreamPos = 0;
        }

        //32. because h264 will insert into 0x000001 delimiter
        if (pPktz->nUnpackBufLen + _nPlyloadLen + 32 > pPktz->nUnpackBufCap) {
                void * pTmp = pj_pool_alloc(pPktz->pH264PacketizerPool, pPktz->nUnpackBufCap * 2);
                pj_memcpy(pTmp, pPktz->pUnpackBuf, pPktz->nUnpackBufLen);
                pPktz->pUnpackBuf = pTmp;
                pPktz->nUnpackBufCap *= 2;
        }

        unsigned nUnpackLen = 0;
        status = pjmedia_h264_unpacketize(pPktz->pH264Packetizer, _pPayload, _nPlyloadLen,
                                          pPktz->pUnpackBuf, pPktz->nUnpackBufCap, &nUnpackLen);
        pPktz->nUnpackBufLen += nUnpackLen;

        int nType = _pPayload[0] & 0x1F;
        if (nType == 24) { //stap-A
                pPktz->nUnpackBufLen = 0;
                *_pBitstreamPos = pPktz->nUnpackBufLen;
                *_pBitstream = pPktz->pUnpackBuf;
        } else if (nType == 28) { //FU-A
                if (_nRtpMarker) {
                        pPktz->bShouldReset = PJ_TRUE;
                        *_pBitstreamPos = pPktz->nUnpackBufLen;
                        *_pBitstream = pPktz->pUnpackBuf;
                } else {
                        *_pBitstreamPos = 0;
                }
        }
        return status;
}

static pj_status_t createH264Packetizer(IN pj_pool_t *_pPktzPool, OUT MediaPacketier **_pPktz)
{
        H264Packetizer *pPktz = pj_pool_alloc(_pPktzPool, sizeof(H264Packetizer));
        PJ_ASSERT_RETURN(pPktz, -2);
        pj_bzero(pPktz, sizeof(H264Packetizer));
        pPktz->pH264PacketizerPool = _pPktzPool;

        pjmedia_h264_packetizer_cfg cfg;
        cfg.mode = PJMEDIA_H264_PACKETIZER_MODE_NON_INTERLEAVED;
        cfg.mtu = PJMEDIA_MAX_MTU;

        pj_status_t status;
        status = pjmedia_h264_packetizer_create(_pPktzPool,
                                                NULL, &pPktz->pH264Packetizer);
        STATUS_CHECK(pjmedia_h264_packetizer_create, status);

        *_pPktz = (MediaPacketier *)pPktz;
        (*_pPktz)->pOperation->packetize = h264_packetize;
        (*_pPktz)->pOperation->unpacketize = h264_unpacketize;

        return PJ_SUCCESS;
}

static pj_status_t createPcmuPacketizer(IN pj_pool_t *_pPktzPool, OUT MediaPacketier **_pPktz)
{
        PcmuPacketizer *pPktz = pj_pool_alloc(_pPktzPool, sizeof(PcmuPacketizer));
        PJ_ASSERT_RETURN(pPktz, -2);
        pj_bzero(pPktz, sizeof(PcmuPacketizer));
        pPktz->pPcmuPacketizerPool = _pPktzPool;

        *_pPktz = (MediaPacketier *)pPktz;

        (*_pPktz)->pOperation->packetize = pcmu_packetize;
        (*_pPktz)->pOperation->unpacketize = pcmu_unpacketize;

        return PJ_SUCCESS;
}

pj_status_t createPacketizer(IN char *_pName, IN int _nNameLen, IN pj_pool_t *_pPktzPool, OUT MediaPacketier **_pPktz)
{
        pj_assert(_nNameLen < 5);

        //to lowercase
        char lowerCase[4];
        for (int i = 0; i < _nNameLen; i++) {
                if (_pName[i] >= 'A' && _pName[i] <= 'Z') {
                        lowerCase[i] = _pName[i] + 32;
                } else {
                        lowerCase[i] = _pName[i];
                }
        }

        pj_str_t pktzName = {lowerCase, _nNameLen};

        if (pj_memcmp(&pktzName, &pcmuPktzName, pcmuPktzName.slen) == 0) {
                return createPcmuPacketizer(_pPktzPool, _pPktz);
        } else if (pj_memcmp(&pktzName, &pcmaPktzName, pcmaPktzName.slen) == 0) {
                
        } else if (pj_memcmp(&pktzName, &aacPktzName, aacPktzName.slen) == 0) {
                
        } else if (pj_memcmp(&pktzName, &h264PktzName, h264PktzName.slen) == 0) {
                return createH264Packetizer(_pPktzPool, _pPktz);
        } else if (pj_memcmp(&pktzName, &h265PktzName, h265PktzName.slen) == 0) {
                
        }

        return -1;;
}

pj_status_t MediaPacketize(IN MediaPacketier *_pPktz,IN pj_uint8_t *_pBitstream, IN pj_size_t _nBitstreamLen,
                           IN unsigned *_pBitstreamPos, OUT const pj_uint8_t **_pPayload, OUT pj_size_t *_nPlyloadLen)
{
        return _pPktz->pOperation->packetize(_pPktz, _pBitstream, _nBitstreamLen, _pBitstreamPos, _pPayload, _nPlyloadLen);
}

pj_status_t MediaUnPacketize(IN OUT MediaPacketier *_pPKtz, IN const pj_uint8_t *_pPayload, IN pj_size_t _nPlyloadLen,
                             OUT pj_uint8_t **_pBitstream, OUT unsigned *_pBitstreamPos, IN int _nRtpMarker)
{
        return _pPKtz->pOperation->unpacketize(_pPKtz, _pPayload, _nPlyloadLen, _pBitstream, _pBitstreamPos, _nRtpMarker);
}
