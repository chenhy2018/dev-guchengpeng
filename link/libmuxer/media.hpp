#ifndef __MEDIA_HPP__
#define __MEDIA_HPP__

#include "common.hpp"

namespace muxer
{
        class MediaFrame;
        class AudioResampler
        {
        public:
                // TODO move following default values
                static const int CHANNELS = 2;
                static const AVSampleFormat SAMPLE_FMT = AV_SAMPLE_FMT_S16;
                static const int CHANNEL_LAYOUT = AV_CH_LAYOUT_STEREO;
                static const int FRAME_SIZE = 1024;
                static const int SAMPLE_RATE = 44100;
        public:
                AudioResampler();
                ~AudioResampler();
                int Resample(IN const std::shared_ptr<MediaFrame>& _pInFrame, OUT std::vector<uint8_t>& buffer);
        private:
                int Init(IN const std::shared_ptr<MediaFrame>& pFrame);
                int Reset();
        private:
                SwrContext* pSwr_ = nullptr; // for resampling
                int nOrigSamplerate_, nOrigChannels_, nOrigForamt_;
        };

        class VideoRescaler
        {
        public:
                // TODO move following default values
                static const AVPixelFormat PIXEL_FMT = AV_PIX_FMT_YUV420P;
        public:
                VideoRescaler(IN int nWidth, IN int nHeight, IN const AVPixelFormat format = VideoRescaler::PIXEL_FMT,
                              IN bool bStretchMode = true, IN int nBgColor = 0x0);
                ~VideoRescaler();
                int Rescale(IN const std::shared_ptr<MediaFrame>& pInFrame, OUT std::shared_ptr<MediaFrame>& pOutFrame);
                int Reset(IN int nWidth, IN int nHeight, IN const AVPixelFormat format = VideoRescaler::PIXEL_FMT,
                          IN bool bStretchMode = false, IN int nBgColor = 0x0);
                int TargetW();
                int TargetH();
        private:
                int Init(IN const std::shared_ptr<MediaFrame>& pFrame);
        private:
                SwsContext* pSws_ = nullptr;
                int nW_, nH_, nOrigW_, nOrigH_;
                AVPixelFormat format_, origFormat_;

                // by default, Rescaler is init with stretch mode
                bool bStretchMode_;
                int nZoomBgColor_;
                int nZoomW_, nZoomH_;
                int nZoomX_, nZoomY_; // picture offset after zooming
        };

        namespace sound {
                void Gain(INOUT const std::shared_ptr<MediaFrame>& pFrame, IN int nPercent = 100);
                void Gain(INOUT const uint8_t* pData, IN int nSize, IN int nPercent = 100);
        }

        namespace color {
                inline void RgbToYuv(IN int _nRgb, OUT uint8_t& _nY, uint8_t& _nU, uint8_t& _nV) {
                        uint8_t nB = _nRgb >> 16;
                        uint8_t nG = (_nRgb >> 8) & 0xff;
                        uint8_t nR = _nRgb & 0xff;
                        _nY = static_cast<uint8_t>((0.257 * nR) + (0.504 * nG) + (0.098 * nB) + 16);
                        _nU = static_cast<uint8_t>((0.439 * nR) - (0.368 * nG) - (0.071 * nB) + 128);
                        _nV = static_cast<uint8_t>(-(0.148 * nR) - (0.291 * nG) + (0.439 * nB) + 128);
                }
        }

        namespace merge {
                const int OVERWRITE_ALPHA = 0x0001;
                void Overlay(IN const std::shared_ptr<MediaFrame>& pFrom, OUT std::shared_ptr<MediaFrame>& pTo, int nFlag = 0);
                bool GetRect(IN const std::shared_ptr<MediaFrame>& pFrom, OUT std::shared_ptr<MediaFrame>& pTo,
                             OUT int32_t& nFromOffsetX, OUT int32_t& nFromOffsetY, OUT int32_t& nToOffsetX, OUT int32_t& nToOffsetY,
                             OUT int32_t& nTargetW, OUT int32_t& nTargetH);
        }

        class AvFilter
        {
        public:
                AvFilter();
                ~AvFilter();
                bool AddBorderFilter(IN int nSize, IN int nColor);
                bool GetPicture(INOUT std::shared_ptr<MediaFrame>& pFrame);
        private:
                int Reset(IN const std::shared_ptr<MediaFrame>& pFrame);
        private:
                struct Border
                {
                        const std::string key = "border";
                        int nSize;
                        int nColor;
                } border_;

                std::map<std::string, std::string> filters_; // filter configs
                std::string filterDesc_ = "";

                // libavfilter contexts
                AVFilterContext *pBufferSinkContext_ = nullptr;
                AVFilterContext *pBufferSrcContext_ = nullptr;
                AVFilterGraph *pFilterGraph_ = nullptr;
        };

        class TextRender
        {
        public:
                TextRender(IN const std::string& text);
                ~TextRender();
                bool GetPicture(IN const std::string& text, IN const std::string& font, IN int nSize, IN int nColor, IN int nOffsetX,
                                IN int nOffsetY, IN int nW, IN int nH, IN int nBgColor, IN int nOpacity, IN int nSpeedx, IN int nSpeedy,
                                OUT std::shared_ptr<MediaFrame>& pFrame);
        private:
                int Reset(IN const std::string& font, IN int nSize, IN int nColor, IN int nOffsetX, IN int nOffsetY, IN int nW, IN int nH,
                          IN int nBgColor, IN int nOpacity, IN int nSpeedx, IN int nSpeedy);
        private:
                std::string text_;
                std::string font_ = "";
                int nSize_ = -1, nW_ = -1, nH_ = -1, nFtColor_ = -1, nBgColor_ = -1, nOpacity_ = -1, nSpeedx_ = -1, nSpeedy_ = -1;
                int nOffsetX_ = -1, nOffsetY_ = -1;

                // libavfilter contexts
                AVFilterContext *pBufferSinkContext_ = nullptr;
                AVFilterContext *pBufferSrcContext_ = nullptr;
                AVFilterGraph *pFilterGraph_ = nullptr;
        };
}

#endif
