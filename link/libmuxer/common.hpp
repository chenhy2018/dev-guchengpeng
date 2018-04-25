#ifndef __COMMON_HPP__
#define __COMMON_HPP__

#include <string>
#include <memory>
#include <vector>
#include <iostream>
#include <unordered_map>
#include <map>
#include <cstring>
#include <functional>
#include <thread>
#include <stdexcept>
#include <ctime>
#include <chrono>
#include <atomic>
#include <mutex>
#include <algorithm>
#include <sys/time.h>
#include <unistd.h>
#include <random>
#include <shared_mutex>
#include <sstream>
#include <iomanip>

extern "C"
{
#include "libavutil/opt.h"
#include "libavformat/avformat.h"
#include "libavcodec/avcodec.h"
#include "libavutil/avutil.h"
#include "libavutil/imgutils.h"
#include "libswresample/swresample.h"
#include "libswscale/swscale.h"
#include <libavfilter/avfiltergraph.h>
#include <libavfilter/buffersink.h>
#include <libavfilter/buffersrc.h>
}

extern "C"
{
#include "rtmp.h"
#include "rtmp_sys.h"
#include "amf.h"
}

#ifndef IN
#define IN
#endif

#ifndef OUT
#define OUT
#endif

#ifndef INOUT
#define INOUT
#endif

#define LogFormat(level, str, fmt, arg...)                              \
        do {                                                            \
                if ((unsigned int)(level) <= muxer::global::nLogLevel) { \
                        struct timeval tv;                              \
                        struct tm* ptm;                                 \
                        char timeFmt[32];                               \
                        gettimeofday(&tv, nullptr);                     \
                        strftime(timeFmt, sizeof(timeFmt), "%Y-%m-%d %H:%M:%S", localtime(&tv.tv_sec)); \
                        fprintf(stderr, "[%s] %s.%03lu: " fmt "\n",     \
                                str, timeFmt, tv.tv_usec / 1000, ##arg); \
                }                                                       \
        } while(0)

#define Fatal(fmt, arg...)                                              \
        do {                                                            \
                LogFormat(1, "F", fmt, ##arg);                          \
                LogFormat(1, "F", "fatal error, will exit");            \
                exit(1);                                                \
        } while(0)

#define Error(fmt, arg...)                              \
        do {                                            \
                LogFormat(2, "E", fmt, ##arg);          \
        } while(0)

#define Warn(fmt, arg...)                               \
        do {                                            \
                LogFormat(3, "W", fmt, ##arg);          \
        } while(0)

#define Info(fmt, arg...)                               \
        do {                                            \
                LogFormat(4, "I", fmt, ##arg);          \
        } while(0)

#define Debug(fmt, arg...)                              \
        do {                                            \
                LogFormat(5, "D", fmt, ##arg);          \
        } while(0)

#define Verbose(fmt, arg...)                            \
        do {                                            \
                LogFormat(6, "V", fmt, ##arg);          \
        } while(0)

namespace muxer
{
        namespace global
        {
                extern unsigned int nLogLevel;

                inline void PrintMem(const void* _pAddr, unsigned long _nBytes)
                {
                        if (_pAddr == nullptr || _nBytes == 0) {
                                return;
                        }

                        const int nBytesPerLine = 16;
                        unsigned char* p = reinterpret_cast<unsigned char*>(const_cast<void*>(_pAddr));
                        std::string line;
                        char value[6];
                        unsigned int i = 0;

                        Verbose("========== memory <%p+%lu> ==========", _pAddr, _nBytes);
                        for (; i < _nBytes; ++i) {
                                if (i % nBytesPerLine == 0) {
                                        line = "";
                                }

                                // "0xAB \0" => 6 bytes
                                snprintf(value, 6, "0x%02x ", p[i]);
                                line += value;

                                if (((i + 1) % nBytesPerLine) == 0) {
                                        Verbose("<%p>: %s", p + i - nBytesPerLine + 1, line.c_str());
                                }
                        }

                        // print rest bytes
                        if (_nBytes % nBytesPerLine != 0) {
                                Verbose("<%p>: %s", p + i - (_nBytes % nBytesPerLine), line.c_str());
                        }
                        Verbose("========== end of <%p+%lu> ==========", _pAddr, _nBytes);
                }
        }

        typedef enum
        {
                STREAM_VIDEO = AVMEDIA_TYPE_VIDEO,
                STREAM_AUDIO = AVMEDIA_TYPE_AUDIO,
                STREAM_DATA = AVMEDIA_TYPE_DATA
        } StreamType;

        typedef enum
        {
                // video
                CODEC_H264 = AV_CODEC_ID_H264,
                CODEC_VC1  = AV_CODEC_ID_VC1,

                // Audio
                CODEC_MP3  = AV_CODEC_ID_MP3,
                CODEC_AAC  = AV_CODEC_ID_AAC,
                CODEC_WAV1 = AV_CODEC_ID_WMAV1,
                CODEC_WAV2 = AV_CODEC_ID_WMAV2,

                // others
                CODEC_FLV_METADATA = AV_CODEC_ID_FFMETADATA,
                CODEC_UNKNOWN = AV_CODEC_ID_NONE
        } CodecType;

        //
        // option map
        //

        namespace options {
                const std::string width  = "w";
                const std::string height = "h";
                const std::string x      = "x";
                const std::string y      = "y";
                const std::string z      = "z";
                const std::string hidden = "hidden";
                const std::string muted  = "muted";
                const std::string vbitrate = "vb";
                const std::string abitrate = "ab";
                const std::string bgcolor  = "bg";
                const std::string ftcolor = "fg";
                const std::string volume  = "vol";
                const std::string stretch = "stretch";
                const std::string opacity = "opacity";
                const std::string font    = "font";
                const std::string speedx  = "speedx";
                const std::string speedy  = "speedy";
                const std::string size    = "size";
                const std::string offsetx = "offsetx";
                const std::string offsety = "offsety";
                const std::string text    = "text";
                const std::string bdcolor = "border_color";
                const std::string bdsize  = "border_size";
        }

        class OptionMap
        {
        public:
                virtual bool GetOption(IN const std::string& key, OUT std::string& value);
                virtual bool GetOption(IN const std::string& key, OUT int& value);
                virtual bool GetOption(IN const std::string& key);
                virtual void SetOption(IN const std::string& flag);
                virtual void SetOption(IN const std::string& key, IN const std::string& value);
                virtual void SetOption(IN const std::string& key, IN int val);
                virtual void DelOption(IN const std::string& key);
                virtual void GetOptions(IN const OptionMap& opts);
        protected:
                std::unordered_map<std::string, std::string> params_;
                std::mutex paramsLck_;

                std::unordered_map<std::string, int> intparams_;
                std::mutex intparamsLck_;
        };

        typedef OptionMap Option;
}

#include "util.hpp"
#include "packet.hpp"
#include "media.hpp"
#include "stat.hpp"

#endif
