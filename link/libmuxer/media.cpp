#include "media.hpp"

using namespace muxer;

//
// AudioResampler
//

const int AudioResampler::CHANNELS;
const AVSampleFormat AudioResampler::SAMPLE_FMT;
const int AudioResampler::FRAME_SIZE;

AudioResampler::AudioResampler()
{
}

AudioResampler::~AudioResampler()
{
        if (pSwr_ != nullptr) {
                swr_free(&pSwr_);
        }
}

int AudioResampler::Init(IN const std::shared_ptr<MediaFrame>& _pFrame)
{
        if (pSwr_ != nullptr) {
                Warn("internal: resampler: already init");
                return -1;
        }

        // for fdkaac encoder, input samples should be PCM signed16le, otherwise do resampling
        if (_pFrame->AvFrame()->format != AudioResampler::SAMPLE_FMT ||
            _pFrame->AvFrame()->channel_layout != AudioResampler::CHANNEL_LAYOUT ||
            _pFrame->AvFrame()->sample_rate != AudioResampler::SAMPLE_RATE) {
                Info("input sample_format=%d, need sample_format=%d, initiate resampling",
                     _pFrame->AvFrame()->format, AudioResampler::SAMPLE_FMT);
                pSwr_ = swr_alloc();
                av_opt_set_int(pSwr_, "in_channel_layout", av_get_default_channel_layout(_pFrame->AvFrame()->channels), 0);
                av_opt_set_int(pSwr_, "out_channel_layout", AudioResampler::CHANNEL_LAYOUT, 0);
                av_opt_set_int(pSwr_, "in_sample_rate", _pFrame->AvFrame()->sample_rate, 0);
                av_opt_set_int(pSwr_, "out_sample_rate", AudioResampler::SAMPLE_RATE, 0);
                av_opt_set_sample_fmt(pSwr_, "in_sample_fmt", static_cast<AVSampleFormat>(_pFrame->AvFrame()->format), 0);
                av_opt_set_sample_fmt(pSwr_, "out_sample_fmt", AudioResampler::SAMPLE_FMT,  0);
                if (swr_init(pSwr_) != 0) {
                        Error("could not initiate resampling");
                        return -1;
                }
                // save original audio attributes
                nOrigSamplerate_ = _pFrame->AvFrame()->sample_rate;
                nOrigChannels_ = _pFrame->AvFrame()->channels;
                nOrigForamt_ = _pFrame->AvFrame()->format;
        }

        return 0;
}

int AudioResampler::Reset()
{
        if (pSwr_ != nullptr) {
                swr_free(&pSwr_);
                pSwr_ = nullptr;
        }

        return 0;
}

int AudioResampler::Resample(IN const std::shared_ptr<MediaFrame>& _pInFrame, OUT std::vector<uint8_t>& _buffer)
{
        // do nothing if format,layout and samplerate are identical
        if (_pInFrame->AvFrame()->format == AudioResampler::SAMPLE_FMT &&
            _pInFrame->AvFrame()->channel_layout == AudioResampler::CHANNEL_LAYOUT &&
            _pInFrame->AvFrame()->sample_rate == AudioResampler::SAMPLE_RATE) {
                _buffer.resize(_pInFrame->AvFrame()->linesize[0]);
                std::copy(_pInFrame->AvFrame()->data[0], _pInFrame->AvFrame()->data[0] + _pInFrame->AvFrame()->linesize[0],
                          _buffer.begin());
                return 0;
        }

        // initial
        if (pSwr_ == nullptr) {
                if (Init(_pInFrame) != 0) {
                        Error("could not init resampling");
                        return -1;
                }
        }
        // incoming audio attrs are changed during transport
        if (_pInFrame->AvFrame()->sample_rate != nOrigSamplerate_ ||
            _pInFrame->AvFrame()->channels != nOrigChannels_ ||
            _pInFrame->AvFrame()->format != nOrigForamt_) {
                Info("resampler: audio parameters have changed, sr=%d, ch=%d, fmt=%d -> sr=%d, ch=%d, fmt=%d", nOrigSamplerate_,
                     nOrigChannels_, nOrigForamt_, _pInFrame->AvFrame()->sample_rate, _pInFrame->AvFrame()->channels,
                     _pInFrame->AvFrame()->format);
                // reset the resampler
                Reset();
                // reinit
                if (Init(_pInFrame) != 0) {
                        pSwr_ = nullptr;
                        return -3;
                }
        }

        int nRetVal;
        uint8_t **pDstData = nullptr;
        int nDstLinesize;
        int nDstBufSize;
        int nDstNbSamples = av_rescale_rnd(_pInFrame->AvFrame()->nb_samples, AudioResampler::SAMPLE_RATE,
                                            _pInFrame->AvFrame()->sample_rate, AV_ROUND_UP);
        int nMaxDstNbSamples = nDstNbSamples;

        // get output buffer
        nRetVal = av_samples_alloc_array_and_samples(&pDstData, &nDstLinesize, AudioResampler::CHANNELS,
                                                 nDstNbSamples, AudioResampler::SAMPLE_FMT, 0);
        if (nRetVal < 0) {
                Error("resampler: could not allocate destination samples");
                return -1;
        }

        // get output samples
        nDstNbSamples = av_rescale_rnd(swr_get_delay(pSwr_, _pInFrame->AvFrame()->sample_rate) + _pInFrame->AvFrame()->nb_samples,
                                       AudioResampler::SAMPLE_RATE, _pInFrame->AvFrame()->sample_rate, AV_ROUND_UP);
        if (nDstNbSamples > nMaxDstNbSamples) {
                av_freep(&pDstData[0]);
                nRetVal = av_samples_alloc(pDstData, &nDstLinesize, AudioResampler::CHANNELS,
                                           nDstNbSamples, AudioResampler::SAMPLE_FMT, 1);
                if (nRetVal < 0) {
                        Error("resampler: could not allocate sample buffer");
                        return -1;
                }
                nMaxDstNbSamples = nDstNbSamples;
        }

        // convert !!
        nRetVal = swr_convert(pSwr_, pDstData, nDstNbSamples, (const uint8_t **)_pInFrame->AvFrame()->extended_data,
                              _pInFrame->AvFrame()->nb_samples);
        if (nRetVal < 0) {
                Error("resampler: converting failed");
                return -1;
        }

        // get output buffer size
        nDstBufSize = av_samples_get_buffer_size(&nDstLinesize, AudioResampler::CHANNELS, nRetVal, AudioResampler::SAMPLE_FMT, 1);
        if (nDstBufSize < 0) {
                Error("resampler: could not get sample buffer size");
                return -1;
        }

        _buffer.resize(nDstBufSize);
        std::copy(pDstData[0], pDstData[0] + nDstBufSize, _buffer.begin());

        // cleanup
        if (pDstData)
                av_freep(&pDstData[0]);
        av_freep(&pDstData);


        return 0;
}

//
// VideoRescaler
//

// TODO delete
const AVPixelFormat VideoRescaler::PIXEL_FMT;

VideoRescaler::VideoRescaler(IN int _nWidth, IN int _nHeight, IN const AVPixelFormat _format,
                             IN bool _bStretchMode, IN int _nBgColor)
{
        if (_nWidth <= 0 || _nHeight <= 0) {
                Error("rescale: resize to width=%d, height=%d", _nWidth, _nHeight);
                return;
        }

        nW_ = _nWidth;
        nH_ = _nHeight;
        format_ = _format;
        bStretchMode_ = _bStretchMode;
        nZoomBgColor_ = _nBgColor;
}

VideoRescaler::~VideoRescaler()
{
        if (pSws_ != nullptr) {
                sws_freeContext(pSws_);
        }
}

int VideoRescaler::Reset(IN int _nWidth, IN int _nHeight, IN const AVPixelFormat _format,
                         IN bool _bStretchMode, IN int _nBgColor)
{
        if (_nWidth <= 0 || _nHeight <= 0) {
                Error("rescale: resize to width=%d, height=%d", _nWidth, _nHeight);
                return -1;
        }

        nW_ = _nWidth;
        nH_ = _nHeight;
        format_ = _format;
        bStretchMode_ = _bStretchMode;
        nZoomBgColor_ = _nBgColor;

        if (pSws_ != nullptr) {
                sws_freeContext(pSws_);
                pSws_ = nullptr;
        }

        return 0;
}

int VideoRescaler::Init(IN const std::shared_ptr<MediaFrame>& _pFrame)
{
        if (pSws_ != nullptr) {
                Warn("internal: rescale: already init");
                return -1;
        }

        auto pAvf = _pFrame->AvFrame();
        nZoomW_ = nW_, nZoomH_ = nH_, nZoomX_ = 0, nZoomY_ = 0;

        Info("input color_space=%d, need color_space=%d, resize_to=%dx%d, stretch=%d initiate rescaling",
             pAvf->format, format_, nW_, nH_, bStretchMode_);

        // non-stretch mode keeps constant aspect ratio, need calculate proper width and height
        if (!bStretchMode_) {
                auto fOrigRatio = static_cast<float>(pAvf->width) / pAvf->height;
                auto fTargetRatio = static_cast<float>(nW_) / nH_;

                if (fTargetRatio > fOrigRatio) {
                        nZoomW_ = nZoomH_ * fOrigRatio;
                        nZoomX_ = nW_ / 2 - nZoomW_ / 2;
                } else if (fTargetRatio < fOrigRatio) {
                        nZoomH_ = nZoomW_ / fOrigRatio;
                        nZoomY_ = nH_ / 2 - nZoomH_ / 2;
                } else {
                        bStretchMode_ = true;
                }
        }

        // configure rescaling context
        pSws_ = sws_getContext(pAvf->width, pAvf->height,
                               static_cast<AVPixelFormat>(pAvf->format), nZoomW_, nZoomH_, format_,
                               SWS_BICUBIC, nullptr, nullptr, nullptr);
        if (pSws_ == nullptr) {
                Error("rescaler initialization failed");
                return -1;
        }

        // save original sample source configurations
        nOrigW_ = pAvf->width;
        nOrigH_ = pAvf->height;
        origFormat_ = static_cast<AVPixelFormat>(pAvf->format);

        return 0;
}

int VideoRescaler::Rescale(IN const std::shared_ptr<MediaFrame>& _pInFrame, OUT std::shared_ptr<MediaFrame>& _pOutFrame)
{
        if (pSws_ == nullptr) {
                if (Init(_pInFrame) != 0) {
                        pSws_ = nullptr;
                        return -1;
                }
        }

        // the incoming frame resolution changed, reinit the sws
        if (_pInFrame->AvFrame()->width != nOrigW_ || _pInFrame->AvFrame()->height != nOrigH_ ||
            _pInFrame->AvFrame()->format != origFormat_) {
                Info("rescaler: video parameters have changed, w=%d, h=%d, fmt=%d -> w=%d, h=%d, fmt=%d", nOrigW_, nOrigH_,
                     origFormat_, _pInFrame->AvFrame()->width, _pInFrame->AvFrame()->height, _pInFrame->AvFrame()->format);
                if (Reset(nW_, nH_, format_) != 0) {
                        Error("rescaler: reinit failed");
                        return -2;
                }
                if (Init(_pInFrame) != 0) {
                        pSws_ = nullptr;
                        return -3;
                }
        }

        auto pRescaled = std::make_shared<MediaFrame>(nZoomW_, nZoomH_, format_);
        pRescaled->Stream(STREAM_VIDEO);
        pRescaled->Codec(CODEC_H264);

        // scale !!
        int nStatus = sws_scale(pSws_, _pInFrame->AvFrame()->data, _pInFrame->AvFrame()->linesize, 0,
                                _pInFrame->AvFrame()->height, pRescaled->AvFrame()->data, pRescaled->AvFrame()->linesize);
        if (nStatus < 0) {
                Error("rescale: failed, status=%d", nStatus);
                return -1;
        }

        // non-stretch mode
        if (!bStretchMode_) {
                // create result picture
                _pOutFrame = std::make_shared<MediaFrame>(nW_, nH_, format_, nZoomBgColor_);
                _pOutFrame->Stream(STREAM_VIDEO);
                _pOutFrame->Codec(CODEC_H264);

                // set internal offset
                pRescaled->X(nZoomX_);
                pRescaled->Y(nZoomY_);

                // copy rescaled picture to the center of output frame
                merge::Overlay(pRescaled, _pOutFrame, merge::OVERWRITE_ALPHA);
        } else {
                _pOutFrame = pRescaled;
        }

        return 0;
}

int VideoRescaler::TargetW()
{
        return nW_;
}

int VideoRescaler::TargetH()
{
        return nH_;
}

void sound::Gain(INOUT const std::shared_ptr<MediaFrame>& _pFrame, IN int _nPercent)
{
        sound::Gain(_pFrame->AvFrame()->data[0], _pFrame->AvFrame()->linesize[0], _nPercent);
}

void sound::Gain(INOUT const uint8_t* _pData, IN int _nSize, IN int _nPercent)
{
        if (_nSize <= 0 || _nSize % 2 != 0) {
                Warn("gain: size not positive or size is not even");
                return;
        }
        if (_nPercent < 0) {
                _nPercent = 0;
        }
        if (_nPercent > 300) {
                _nPercent = 300;
        }

        int16_t* p16 = (int16_t*)_pData;
        for (int i = 0; i < _nSize; i += 2) {
                int32_t nGained = static_cast<int32_t>(*p16) * _nPercent / 100;
                if (nGained < -0x80000) {
                        nGained = -0x80000;
                } else if (nGained > 0x7fff) {
                        nGained = 0x7fff;
                }
                *p16++ = nGained;
        }
}

void merge::Overlay(IN const std::shared_ptr<MediaFrame>& _pFrom, OUT std::shared_ptr<MediaFrame>& _pTo, int _nFlag)
{
        // format check
        AVFrame* pFrom = _pFrom->AvFrame();
        AVFrame* pTo = _pTo->AvFrame();
        if ((pFrom->format != AV_PIX_FMT_YUV420P && pFrom->format != AV_PIX_FMT_YUVA420P) ||
            (pTo->format != AV_PIX_FMT_YUV420P && pTo->format != AV_PIX_FMT_YUVA420P)) {
                Warn("internal: yuv overlay: src_fmt=%d or dst_fmt=%d is not yuv420p", pFrom->format, pTo->format);
                return;
        }

        // get overlay rect
        int32_t nFromOffsetX, nFromOffsetY, nToOffsetX, nToOffsetY, nTargetH, nTargetW;
        if (GetRect(_pFrom, _pTo, nFromOffsetX, nFromOffsetY, nToOffsetX, nToOffsetY, nTargetW, nTargetH) == false) {
                return;
        }

        // Y plane

        // linesize[] might not be equal to the actual width of the video due to alignment
        // (refer to the last arg of av_frame_get_buffer())
        int32_t nFromOffset = pFrom->linesize[0] * nFromOffsetY + nFromOffsetX;
        int32_t nToOffset = pTo->linesize[0] * nToOffsetY + nToOffsetX;

        if (pFrom->format == AV_PIX_FMT_YUVA420P) {
                // blend with alpha
                for (int32_t i = 0; i < nTargetH; ++i) {
                        for (int32_t j = 0; j < nTargetW; ++j) {
                                uint8_t* pDst = pTo->data[0] + nToOffset + pTo->linesize[0] * i + j;
                                uint8_t* pSrc = pFrom->data[0] + nFromOffset + pFrom->linesize[0] * i + j;
                                uint8_t* pAlpha = pFrom->data[3] + nFromOffset + pFrom->linesize[3] * i + j;
                                uint32_t nAlpha = *pAlpha;
                                *pDst = (((*pSrc) * (nAlpha) + (*pDst) * (0xff - nAlpha)) >> 8);
                        }
                        if (_nFlag & merge::OVERWRITE_ALPHA) {
                                std::memcpy(pTo->data[3] + pTo->linesize[3] * i + nToOffset,
                                            pFrom->data[3] + pFrom->linesize[3] * i + nFromOffset,
                                            nTargetW);
                        }
                }
        } else  {
                // copy Y plane data from src frame to dst frame
                for (int32_t i = 0; i < nTargetH; ++i) {
                        std::memcpy(pTo->data[0] + pTo->linesize[0] * i + nToOffset,
                                    pFrom->data[0] + pFrom->linesize[0] * i + nFromOffset,
                                    nTargetW);
                }
        }

        // UV plane

        int32_t nFromUVOffsetX  = nFromOffsetX / 2;     // row UV data offset of pFrom
        int32_t nFromUVOffsetY  = nFromOffsetY / 2;     // colume UV data offset of pFrom
        int32_t nToUVOffsetX    = nToOffsetX / 2;       // row UV data offset of pTo
        int32_t nToUVOffsetY    = nToOffsetY / 2;       // colume UV data offset of pTo
        int32_t nTargetUVX      = nTargetW / 2;         // width of mix area for UV
        int32_t nTargetUVY      = nTargetH / 2;         // height of mix area for UV

        int32_t nFromUVOffset   = pFrom->linesize[1] * nFromUVOffsetY + nFromUVOffsetX;
        int32_t nToUVOffset     = pTo->linesize[1] * nToUVOffsetY + nToUVOffsetX;

        if (pFrom->format == AV_PIX_FMT_YUVA420P) {
                // blend with alpha
                for (int32_t i = 0; i < nTargetUVY; ++i) {
                        for (int32_t j = 0; j < nTargetUVX; ++j) {
                                // get average alpha value
                                uint8_t* pAlpha0 = pFrom->data[3] + nFromOffset + pFrom->linesize[3] * i * 2 + j * 2;
                                uint8_t* pAlpha1 = pFrom->data[3] + nFromOffset + pFrom->linesize[3] * i * 2 + j * 2 + 1;
                                uint8_t* pAlpha2 = pFrom->data[3] + nFromOffset + pFrom->linesize[3] * (i * 2 + 1) + j * 2;
                                uint8_t* pAlpha3 = pFrom->data[3] + nFromOffset + pFrom->linesize[3] * (i * 2 + 1) + j * 2 + 1;
                                uint32_t nAlpha = 0;
                                nAlpha = (nAlpha + *pAlpha0 + *pAlpha1 + *pAlpha2 + *pAlpha3) / 4;

                                // U
                                uint8_t* pDst = pTo->data[1] + nToUVOffset + pTo->linesize[1] * i + j;
                                uint8_t* pSrc = pFrom->data[1] + nFromUVOffset + pFrom->linesize[1] * i + j;
                                *pDst = (((*pSrc) * (nAlpha) + (*pDst) * (0xff - nAlpha)) >> 8);
                                // V
                                pDst = pTo->data[2] + nToUVOffset + pTo->linesize[2] * i + j;
                                pSrc = pFrom->data[2] + nFromUVOffset + pFrom->linesize[2] * i + j;
                                *pDst = (((*pSrc) * (nAlpha) + (*pDst) * (0xff - nAlpha)) >> 8);
                        }
                }
        } else {
                // copy UV plane data from src to dst
                for (int32_t j = 0; j < nTargetUVY; ++j) {
                        std::memcpy(pTo->data[1] + nToUVOffset + pTo->linesize[1] * j,
                                    pFrom->data[1] + nFromUVOffset + pFrom->linesize[1] * j, nTargetUVX);
                        std::memcpy(pTo->data[2] + nToUVOffset + pTo->linesize[2] * j,
                                    pFrom->data[2] + nFromUVOffset + pFrom->linesize[2] * j, nTargetUVX);
                }
        }
}

bool merge::GetRect(IN const std::shared_ptr<MediaFrame>& _pFrom, OUT std::shared_ptr<MediaFrame>& _pTo,
             OUT int32_t& _nFromOffsetX, OUT int32_t& _nFromOffsetY, OUT int32_t& _nToOffsetX, OUT int32_t& _nToOffsetY,
             OUT int32_t& _nTargetW, OUT int32_t& _nTargetH)
{
        AVFrame* pFrom = _pFrom->AvFrame();
        AVFrame* pTo = _pTo->AvFrame();
        auto nX = _pFrom->X();
        auto nY = _pFrom->Y();

        if (pFrom == nullptr || pTo == nullptr) {
                return false;
        }

        // x or y is beyond width and height of target frame
        if (nX >= pTo->width || nY >= pTo->height) {
                return false;
        }

        // left-top offset point of source frame from which source frame will be copied
        _nFromOffsetX = 0;
        _nFromOffsetY = 0;

        // left-top offset point of target frame to which source frame will copy data
        _nToOffsetX = 0;
        _nToOffsetY = 0;

        // final resolution of the source frame
        _nTargetH = 0;
        _nTargetW = 0;

        if (nX < 0) {
                if (nX + pFrom->width < 0) {
                        // whole frame is to the left side of target
                        return false;
                }
                _nFromOffsetX = -nX;
                _nToOffsetX = 0;
                _nTargetW = (pFrom->width + nX < pTo->width) ? (pFrom->width + nX) : pTo->width;
        } else {
                _nFromOffsetX = 0;
                _nToOffsetX = nX;
                _nTargetW = (_nToOffsetX + pFrom->width > pTo->width) ? (pTo->width - _nToOffsetX) : pFrom->width;
        }

        if (nY < 0) {
                if (nY + pFrom->height < 0) {
                        // whole original frame is beyond top side of target
                        return false;
                }
                _nFromOffsetY = -nY;
                _nToOffsetY = 0;
                _nTargetH = (pFrom->height + nY < pTo->height) ? (pFrom->height + nY) : pTo->height;
        } else {
                _nFromOffsetY = 0;
                _nToOffsetY = nY;
                _nTargetH = (pFrom->height + _nToOffsetY > pTo->height) ? (pTo->height - _nToOffsetY) : pFrom->height;
        }

        return true;
}

//
// AvFilter
//

AvFilter::AvFilter()
{
}

AvFilter::~AvFilter()
{
        // clean the filter graph
        if (pFilterGraph_ != nullptr) {
                avfilter_graph_free(&pFilterGraph_);
                pFilterGraph_ = nullptr;
        }
}

bool AvFilter::AddBorderFilter(IN int _nSize, IN int _nColor)
{
        auto compose = [](IN int _nSize, IN int _nColor) {
                // get hex string
                auto toHex = [](int _nNum, int _nFill) -> std::string {
                        std::stringstream stream;
                        stream << std::setfill('0') << std::setw(_nFill) << std::hex << _nNum;
                        return stream.str();
                };

                std::string desc = "drawbox=";
                desc += "t=" + std::to_string(_nSize) + ":";
                desc += "color=0x" + toHex(_nColor & 0xffffff, 6);

                return desc;
        };

        auto it = filters_.find(border_.key);
        if (it != filters_.end()) {
                if (_nSize != border_.nSize || _nColor != border_.nColor) {
                        filters_[border_.key] = compose(_nSize, _nColor);
                }
        } else {
                filters_[border_.key] = compose(_nSize, _nColor);
        }

        border_.nSize = _nSize;
        border_.nColor = _nColor;

        return true;
}

bool AvFilter::GetPicture(INOUT std::shared_ptr<MediaFrame>& _pFrame)
{
        // get filter config, if config changed, reset the filter
        std::string desc = "";
        for (auto it = filters_.begin(); it != filters_.end();) {
                desc += it->second;
                it++;
                if (it != filters_.end()) {
                        desc += ",";
                }
        }
        if (desc.empty()) {
                return true;
        }
        if (desc.compare(filterDesc_) != 0) {
                if (Reset(_pFrame) != 0) {
                        return false;
                }
        }
        filters_.clear();

        // feed one frame
        int nRet;
        char error[50];
        nRet = av_buffersrc_add_frame_flags(pBufferSrcContext_, _pFrame->AvFrame(), AV_BUFFERSRC_FLAG_KEEP_REF);
        if (nRet < 0) {
                av_strerror(nRet, error, sizeof(error));
                Error("avfilter: could not send frames to the filter graph: %s", error);
                return false;
        }

        // get a filtered frame
        bool bGotFrame = false;
        while (1) {
                auto pOutFrame = std::make_shared<MediaFrame>();
                int nRet = av_buffersink_get_frame(pBufferSinkContext_, pOutFrame->AvFrame());
                if (nRet == AVERROR(EAGAIN) || nRet == AVERROR_EOF) {
                        return bGotFrame;
                }
                if (nRet < 0) {
                        av_strerror(nRet, error, sizeof(error));
                        Error("avfilter: could not get frames from the filter graph: %s", error);
                        return false;
                }
                bGotFrame = true;
                _pFrame = pOutFrame;
        }

        return true;
}

int AvFilter::Reset(IN const std::shared_ptr<MediaFrame>& _pFrame)
{
        if (filters_.size() == 0) {
                return 1;
        }

        // reset the graph
        if (pFilterGraph_ != nullptr) {
                avfilter_graph_free(&pFilterGraph_);
                pFilterGraph_ = nullptr;
        }

        // alloc filters and filter graph
        const AVFilter *pBufferSrc  = avfilter_get_by_name("buffer");
        const AVFilter *pBufferSink = avfilter_get_by_name("buffersink");
        AVFilterInOut *pOutputs = avfilter_inout_alloc();
        AVFilterInOut *pInputs  = avfilter_inout_alloc();
        pFilterGraph_ = avfilter_graph_alloc();
        if (!pOutputs || !pInputs || !pFilterGraph_) {
                Error("avfilter: filter could not be created: in=%p, out=%p, graph=%p",
                      pInputs, pOutputs, pFilterGraph_);
                return false;
        }

        int nRet = 0;
        const AVFrame* pAvFrame = _pFrame->AvFrame();
        do {
                // create filters
                char args[512];
                snprintf(args, sizeof(args), "video_size=%dx%d:pix_fmt=%d:time_base=1/25:pixel_aspect=0/1",
                         pAvFrame->width, pAvFrame->height, pAvFrame->format);
                nRet = avfilter_graph_create_filter(&pBufferSrcContext_, pBufferSrc, "in", args, nullptr, pFilterGraph_);
                if (nRet < 0) {
                        Error("avfilter: buffer source init failed: %d", nRet);
                        break;
                }
                nRet = avfilter_graph_create_filter(&pBufferSinkContext_, pBufferSink, "out", nullptr, nullptr, pFilterGraph_);
                if (nRet < 0) {
                        Error("avfilter: buffer sink init failed: %d", nRet);
                        break;
                }
                enum AVPixelFormat fmts[] = {static_cast<AVPixelFormat>(pAvFrame->format), AV_PIX_FMT_NONE};
                nRet = av_opt_set_int_list(pBufferSinkContext_, "pix_fmts", fmts, AV_PIX_FMT_NONE, AV_OPT_SEARCH_CHILDREN);
                if (nRet < 0) {
                        Error("avfilter: failed to set output format");
                        break;
                }

                // input pad
                pOutputs->name       = av_strdup("in");
                pOutputs->filter_ctx = pBufferSrcContext_;
                pOutputs->pad_idx    = 0;
                pOutputs->next       = nullptr;
                // output pad
                pInputs->name       = av_strdup("out");
                pInputs->filter_ctx = pBufferSinkContext_;
                pInputs->pad_idx    = 0;
                pInputs->next       = nullptr;

                // get hex string
                auto toHex = [](int _nNum, int _nFill) -> std::string {
                        std::stringstream stream;
                        stream << std::setfill('0') << std::setw(_nFill) << std::hex << _nNum;
                        return stream.str();
                };

                // filter description
                std::string desc = "";
                for (auto it = filters_.begin(); it != filters_.end();) {
                        desc += it->second;
                        it++;
                        if (it != filters_.end()) {
                                desc += ",";
                        }
                }
                Info("avfilter: %s", desc.c_str());
                filterDesc_ = desc;

                // setup filter graph
                if ((nRet = avfilter_graph_parse_ptr(pFilterGraph_, desc.c_str(), &pInputs, &pOutputs, nullptr)) < 0) {
                        avfilter_graph_free(&pFilterGraph_);
                        Error("avfilter: could not parse filter graph: %d", nRet);
                        break;
                }
                if ((nRet = avfilter_graph_config(pFilterGraph_, nullptr)) < 0) {
                        avfilter_graph_free(&pFilterGraph_);
                        Error("avfilter: could not config filter graph: %d", nRet);
                        break;
                }
        } while(0);

        avfilter_inout_free(&pInputs);
        avfilter_inout_free(&pOutputs);

        return nRet;
}

//
// TextRender
//

TextRender::TextRender(IN const std::string& _text)
        :text_(_text)
{
}

TextRender::~TextRender()
{
        // clean the filter graph
        if (pFilterGraph_ != nullptr) {
                avfilter_graph_free(&pFilterGraph_);
                pFilterGraph_ = nullptr;
        }
}

bool TextRender::GetPicture(IN const std::string& _text, IN const std::string& _font, IN int _nSize, IN int _nColor, IN int _nOffsetX,
                            IN int _nOffsetY, IN int _nW, IN int _nH, IN int _nBgColor, IN int _nOpacity, IN int _nSpeedx, IN int _nSpeedy,
                            OUT std::shared_ptr<MediaFrame>& _pFrame)
{
        if (_text.compare(text_) != 0 || _font.compare(font_) != 0 || _nSize != nSize_ || _nW != nW_ || _nH != nH_ || _nColor != nFtColor_ ||
            _nBgColor != nBgColor_ || _nOpacity != nOpacity_ || _nSpeedx != nSpeedx_  || _nSpeedy != nSpeedy_ ||
            _nOffsetX != nOffsetX_ || _nOffsetY != nOffsetY_ || pFilterGraph_ == nullptr) {
                text_ = _text; // update text
                if (Reset(_font, _nSize, _nColor, _nOffsetX, _nOffsetY, _nW, _nH, _nBgColor, _nOpacity, _nSpeedx, _nSpeedy) != 0) {
                        return false;
                }
        }

        // feed one frame
        int nRet;
        char error[50];
        auto pInFrame = std::make_shared<MediaFrame>(_nW, _nH, AV_PIX_FMT_YUVA420P, 0x000000);
        nRet = av_buffersrc_add_frame_flags(pBufferSrcContext_, pInFrame->AvFrame(), AV_BUFFERSRC_FLAG_KEEP_REF);
        if (nRet < 0) {
                av_strerror(nRet, error, sizeof(error));
                Error("text render: could not send frames to the filter graph: %s", error);
                return false;
        }

        // get a filtered frame
        bool bGotFrame = false;
        while (1) {
                auto pOutFrame = std::make_shared<MediaFrame>();
                int nRet = av_buffersink_get_frame(pBufferSinkContext_, pOutFrame->AvFrame());
                if (nRet == AVERROR(EAGAIN) || nRet == AVERROR_EOF) {
                        return bGotFrame;
                }
                if (nRet < 0) {
                        av_strerror(nRet, error, sizeof(error));
                        Error("text render: could not get frames from the filter graph: %s", error);
                        return false;
                }
                bGotFrame = true;
                _pFrame = pOutFrame;
        }

        return true;
}

int TextRender::Reset(IN const std::string& _font, IN int _nSize, IN int _nColor, IN int _nOffsetX, IN int _nOffsetY,
                      IN int _nW, IN int _nH, IN int _nBgColor, IN int _nOpacity, IN int _nSpeedx, IN int _nSpeedy)
{
        // reset the graph
        if (pFilterGraph_ != nullptr) {
                avfilter_graph_free(&pFilterGraph_);
                pFilterGraph_ = nullptr;
        }

        // alloc filters and filter graph
        const AVFilter *pBufferSrc  = avfilter_get_by_name("buffer");
        const AVFilter *pBufferSink = avfilter_get_by_name("buffersink");
        AVFilterInOut *pOutputs = avfilter_inout_alloc();
        AVFilterInOut *pInputs  = avfilter_inout_alloc();
        pFilterGraph_ = avfilter_graph_alloc();
        if (!pOutputs || !pInputs || !pFilterGraph_) {
                Error("text render: filter could not be created: in=%p, out=%p, graph=%p",
                      pInputs, pOutputs, pFilterGraph_);
                return false;
        }

        int nRet = 0;
        do {
                // create filters
                char args[512];
                snprintf(args, sizeof(args), "video_size=%dx%d:pix_fmt=%d:time_base=1/25:pixel_aspect=0/1",
                         _nW, _nH, AV_PIX_FMT_YUVA420P);
                nRet = avfilter_graph_create_filter(&pBufferSrcContext_, pBufferSrc, "in", args, nullptr, pFilterGraph_);
                if (nRet < 0) {
                        Error("text render: buffer source init failed: %d", nRet);
                        break;
                }
                nRet = avfilter_graph_create_filter(&pBufferSinkContext_, pBufferSink, "out", nullptr, nullptr, pFilterGraph_);
                if (nRet < 0) {
                        Error("text render: buffer sink init failed: %d", nRet);
                        break;
                }
                enum AVPixelFormat fmts[] = {AV_PIX_FMT_YUVA420P, AV_PIX_FMT_NONE};
                nRet = av_opt_set_int_list(pBufferSinkContext_, "pix_fmts", fmts, AV_PIX_FMT_NONE, AV_OPT_SEARCH_CHILDREN);
                if (nRet < 0) {
                        Error("text render: failed to set output format");
                        break;
                }

                // input pad
                pOutputs->name       = av_strdup("in");
                pOutputs->filter_ctx = pBufferSrcContext_;
                pOutputs->pad_idx    = 0;
                pOutputs->next       = nullptr;
                // output pad
                pInputs->name       = av_strdup("out");
                pInputs->filter_ctx = pBufferSinkContext_;
                pInputs->pad_idx    = 0;
                pInputs->next       = nullptr;

                // get hex string
                auto toHex = [](int _nNum, int _nFill) -> std::string {
                        std::stringstream stream;
                        stream << std::setfill('0') << std::setw(_nFill) << std::hex << _nNum;
                        return stream.str();
                };

                // drawbox filter description
                std::string desc = "";
                desc += "drawbox=color=0x" + toHex(_nBgColor & 0xffffff, 6) + "@0x" + toHex(_nOpacity & 0xff, 2) + ":";
                desc += "width=" + std::to_string(_nW) + ":";
                desc += "height=" + std::to_string(_nH) + ":";
                desc += "t=max,";
                // drawtext filter description
                desc += "drawtext=fontfile=" + _font + ".ttf:";
                desc += "text='" + text_ + "':";
                desc += "fontsize=" + std::to_string(_nSize) + ":";
                desc += "fontcolor=0x" + toHex(_nColor & 0xffffff, 6) + "@0xff:";
                // text animation
                if (_nSpeedy > 0) {
                        desc += "y='if(gte(y,main_h), 0-text_h, (-text_h+mod((" + std::to_string(_nSpeedy) + "*n/25), main_h+text_h)))':";
                } else if (_nSpeedy < 0) {
                        desc += "y='if(gte(0,y+text_h), main_h, (main_h-mod((" + std::to_string(-_nSpeedy) + "*n/25), main_h+text_h)))':";
                } else {
                        desc += "y=" + std::to_string(_nOffsetY) + ":";
                }
                if (_nSpeedx > 0) {
                        desc += "x='if(gte(x,main_w), 0-text_w, (-text_w+mod((" + std::to_string(_nSpeedx) + "*n/25), main_w+text_w)))'";
                } else if (_nSpeedx < 0) {
                        desc += "x='if(gte(0,x+text_w), main_w, (main_w-mod((" + std::to_string(-_nSpeedx) + "*n/25), main_w+text_w)))'";
                } else {
                        desc += "x=" + std::to_string(_nOffsetX);
                }
                // print text filter
                Info("text render: text filter: %s", desc.c_str());

                // setup filter graph
                if ((nRet = avfilter_graph_parse_ptr(pFilterGraph_, desc.c_str(), &pInputs, &pOutputs, nullptr)) < 0) {
                        avfilter_graph_free(&pFilterGraph_);
                        Error("text render: could not parse filter graph: %d", nRet);
                        break;
                }
                if ((nRet = avfilter_graph_config(pFilterGraph_, nullptr)) < 0) {
                        avfilter_graph_free(&pFilterGraph_);
                        Error("text render: could not config filter graph: %d", nRet);
                        break;
                }

                // update internal attributes
                font_ = _font;
                nSize_ = _nSize;
                nFtColor_ = _nColor;
                nOffsetX_ = _nOffsetX;
                nOffsetY_ = _nOffsetY;
                nW_ = _nW;
                nH_ = _nH;
                nBgColor_ = _nBgColor;
                nOpacity_ = _nOpacity;
                nSpeedx_ = _nSpeedx;
                nSpeedy_ = _nSpeedy;
        } while(0);

        avfilter_inout_free(&pInputs);
        avfilter_inout_free(&pOutputs);

        return nRet;
}
