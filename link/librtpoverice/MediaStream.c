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
}

int CreateSdpAudioMLine(IN pjmedia_endpt *_pMediaEndpt, IN pjmedia_transport_info *_pTransportInfo,
                        IN pj_pool_t * _pPool, IN MediaStreamTrack *_pMediaTrack, OUT pjmedia_sdp_media ** _pAudioSdp)
{
    pj_assert(_pMediaTrack->type == TYPE_AUDIO);
    
    pj_status_t status;
    status = pjmedia_endpt_create_audio_sdp(_pMediaEndpt, _pPool, &_pTransportInfo->sock_info, 0, _pAudioSdp);
    STATUS_CHECK(pjmedia_endpt_create_audio_sdp, status);
    
    switch (_pMediaTrack->mediaConfig.audioConfig.format) {
        case MEDIA_FORMAT_PCMU:
            (*_pAudioSdp)->desc.fmt[(*_pAudioSdp)->desc.fmt_count++] = pj_str("0");
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
    status = pjmedia_endpt_create_audio_sdp(_pMediaEndpt, _pPool, &_pTransportInfo->sock_info, 0, _pVideoSdp);
    STATUS_CHECK(pjmedia_endpt_create_audio_sdp, status);
    
    switch (_pMediaTrack->mediaConfig.videoConfig.format) {
        case MEDIA_FORMAT_H264:
            (*_pVideoSdp)->desc.fmt[(*_pVideoSdp)->desc.fmt_count++] = pj_str("116");
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
