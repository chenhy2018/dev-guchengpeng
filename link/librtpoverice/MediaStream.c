#include "MediaStream.h"

void InitMediaConfig(IN MediaConfig * pMediaConfig)
{
    pj_bzero(pMediaConfig, sizeof(MediaConfig));
}

void InitMediaStream(IN MediaStream *_pMediaStraem)
{
    pj_bzero(_pMediaStraem, sizeof(MediaStream));
}

void AddMediaTrack(IN OUT MediaStream *_pMediaStraem, IN MediaConfig *_pMediaConfig, IN int _nIndex, IN MediaType _type)
{
    pj_assert(_pMediaStraem && _pMediaConfig);
    
    for (int i = 0; i < _pMediaStraem->nCount; i++) {
        if ( _pMediaStraem->streamTracks[i].type == _type ){
            return;
        }
    }
    
    _pMediaStraem->nCount++;

    _pMediaStraem->streamTracks[_nIndex].type = _type;
    _pMediaStraem->streamTracks[_nIndex].mediaConfig = *_pMediaConfig;

    switch (_pMediaConfig->format) {
        case MEDIA_FORMAT_PCMU:
            _pMediaConfig->nChannel = 1;
            _pMediaConfig->nBitDepth = 8;
            break;
        case MEDIA_FORMAT_H264:
            break;
    }
}

int CreateSdpAudioMLine(IN pjmedia_endpt *_pMediaEndpt, IN pjmedia_transport_info *_pTransportInfo,
                        IN pj_pool_t * _pPool, IN MediaStreamTrack *_pMediaTrack, OUT pjmedia_sdp_media ** _pAudioSdp)
{
    pj_assert(_pMediaTrack->type == TYPE_AUDIO);
    
    pj_status_t status;
    status = pjmedia_endpt_create_audio_sdp(_pMediaEndpt, _pPool, &_pTransportInfo->sock_info, 0, _pAudioSdp);
    STATUS_CHECK(pjmedia_endpt_create_audio_sdp, status);
    
    pj_str_t * fmt;
    switch (_pMediaTrack->mediaConfig.format) {
        case MEDIA_FORMAT_PCMU:
            fmt = &((*_pAudioSdp)->desc.fmt[(*_pAudioSdp)->desc.fmt_count++]);
            fmt->ptr = "0";
            fmt->slen = 1;
            break;
        case MEDIA_FORMAT_H264:
            break;
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
    switch (_pMediaTrack->mediaConfig.format) {
        case MEDIA_FORMAT_H264:
            fmt = &((*_pVideoSdp)->desc.fmt[(*_pVideoSdp)->desc.fmt_count++]);
            fmt->ptr = "116";
            fmt->slen = 3;
            pjmedia_sdp_attr *pAttr = NULL;
            pjmedia_sdp_rtpmap rtpmap;
            pj_bzero(&rtpmap, sizeof(rtpmap));
            rtpmap.pt = pj_str("116");
            rtpmap.clock_rate = 90000;
            rtpmap.enc_name = pj_str("H264");
            pjmedia_sdp_rtpmap_to_attr(_pPool, &rtpmap, &pAttr);
            (*_pVideoSdp)->attr[(*_pVideoSdp)->attr_count++] = pAttr;
            break;
        case MEDIA_FORMAT_PCMU:
            break;
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

