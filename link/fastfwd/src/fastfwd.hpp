#ifndef __FASTFWD_HPP__
#define __FASTFWD_HPP__

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
#include <fstream>

extern "C"
{
#include "libavutil/opt.h"
#include "libavformat/avformat.h"
#include "libavcodec/avcodec.h"
#include "libavutil/avutil.h"
#include "libavutil/imgutils.h"
#include "libswresample/swresample.h"
#include "libswscale/swscale.h"
}

#include "util.hpp"

#define IN
#define OUT


#define LogFormat(level, str, fmt, arg...)                              \
        do {                                                            \
                if ((unsigned int)(level) <= fastfwd::global::nLogLevel) { \
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


namespace fastfwd {

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

        const int x2 = 2;
        const int x4 = 4;
        const int x8 = 8;
        const int x16 = 16;
        const int x32 = 32;

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

        class MediaPacket
        {
        public:
                MediaPacket(IN const AVStream& pAvStream, IN const AVPacket* pAvPacket, IN bool bRef = true);
                MediaPacket(IN const AVCodecContext* _pAvCodecContext);
                ~MediaPacket();
                MediaPacket();
                MediaPacket(const MediaPacket&) = delete; // no copy for risk concern

                // get raw AV structs
                AVPacket* AvPacket() const;
                AVCodecParameters* AvCodecParameters() const; // used by receiver and decoder
                AVCodecContext* AvCodecContext() const;       // used by encoder and sender

                // pts and dts
                uint64_t Pts() const;
                void Pts(IN uint64_t);
                uint64_t Dts() const;
                void Dts(IN uint64_t);

                // stream and codec
                StreamType Stream() const;
                void Stream(IN StreamType);
                CodecType Codec() const;
                void Codec(IN CodecType);

                // data fields
                char* Data()const;
                int Size() const;

                // util
                void Print() const;
                void Dump(const std::string& title = "") const;

                // video
                int Width() const;
                int Height() const;
                void Width(int);
                void Height(int);
                bool IsKey() const;
                void SetKey();

                // audio
                int SampleRate() const;
                int Channels() const;
                void SampleRate(int);
                void Channels(int);

        private:
                bool bRef_ = false; // true: unref the packet, false: free the packet
                AVPacket* pAvPacket_ =  nullptr;
                AVCodecParameters* pAvCodecPar_ = nullptr;
                AVCodecContext* pAvCodecContext_ = nullptr;

                // save following fields seperately
                StreamType stream_;
                CodecType codec_;

                // video specific
                int nWidth_ = -1, nHeight_ = -1;
                int nSampleRate_ = -1, nChannels_ = -1;
        };

        // AvReceiver
        typedef const std::function<int(const std::shared_ptr<MediaPacket>)> PacketHandlerType;
        class AvReceiver
        {
        public:
                AvReceiver();
                ~AvReceiver();
                int Receive(IN const std::string& url, IN int nXspeed, IN PacketHandlerType& callback);
        private:
                struct StreamInfo;
                static int AvInterruptCallback(void* pContext);
                static bool EmulateFramerate(IN int64_t nPts, OUT StreamInfo& stream, IN int _nXspeed);
        private:
                std::chrono::high_resolution_clock::time_point start_;
                long nTimeout_ = 10000; // 10 seconds timeout by default

                struct AVFormatContext* pAvContext_ = nullptr;
                struct StreamInfo {
                        struct AVStream* pAvStream;
                        int64_t nCount = -1;
                        std::chrono::high_resolution_clock::time_point start;
                };
                std::vector<StreamInfo> streams_;
        };

        class FileSink
        {
        public:
                FileSink(IN const std::shared_ptr<SharedQueue<std::vector<char>>> pPool, IN int nBlockSize, IN int nXspeed);
                ~FileSink();
                int Write(IN const std::shared_ptr<MediaPacket>& pPacket);
        private:
                int AddStream(IN const std::shared_ptr<MediaPacket>& pPacket);
                bool WriteHeader();
                bool WritePackets(IN int nXspeed);
                bool Init();
                static int WriteFunction(IN void *pOpaque, IN uint8_t* pBuf, IN int nSize);
        private:
                std::shared_ptr<SharedQueue<std::vector<char>>> pPool_;
                int nBlockSize_;
                int nXspeed_;

                AVFormatContext* pOutputContext_ = nullptr;
                AVBSFContext *pAacBsf_ = nullptr;

                bool bAudioAdded_ = false;
                bool bVideoAdded_ = false;
                int nStreamIndex_ = 0;
                std::unordered_map<int, int> streams_;
                SharedQueue<std::shared_ptr<MediaPacket>> probeQ_;

                std::string path_ = "";
                size_t nCount_ = 0;
                long long nLastPtsOri_ = -1;
                long long nLastPtsMod_ = -1;

                AVIOContext* pAvIoContext_ = nullptr;
                uint8_t* pMemBuffer_ = nullptr;
        };

        class StreamPumper
        {
        public:
                static const int ok = 0;
                static const int eof = -5;
                static const int econn = -10;
        public:
                StreamPumper(IN const std::string& url, IN int nXspeed = fastfwd::x2, IN int nBlockSize = 1024);
                ~StreamPumper();
                int Pump(OUT std::vector<char>& stream, IN int nTimeout);
        private:
                void StartPumper();
                void StopPumper();
        private:
                // parameters
                std::string url_;
                int nXspeed_;
                int nBlockSize_;

                // pumper thread
                std::shared_ptr<SharedQueue<std::vector<char>>> pPool_ = nullptr;
                std::thread pumper_;
                std::mutex pumpLck_;
                std::atomic<bool> bPumperStopped_;
                std::atomic<int> nStatus_;
                bool bStarted_ = false;
                std::once_flag avformatInit_;

                // util
                std::unique_ptr<AvReceiver> pReceiver_ = nullptr;
                std::unique_ptr<FileSink> pSink_ = nullptr;
        };
}

#endif
