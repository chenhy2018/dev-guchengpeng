#include "fastfwd.hpp"

using namespace fastfwd;

//
// MediaPacket
//

MediaPacket::MediaPacket(IN const AVStream& _avStream, IN const AVPacket* _pAvPacket, IN bool _bRef)
        :bRef_(_bRef)
{
        // save codec pointer
        pAvCodecPar_ = _avStream.codecpar;
        StreamType stream = static_cast<StreamType>(pAvCodecPar_->codec_type);
        if (stream != STREAM_AUDIO && stream != STREAM_VIDEO) {
                stream = STREAM_DATA;
        }
        Stream(stream);
        Codec(static_cast<CodecType>(pAvCodecPar_->codec_id));
        Width(pAvCodecPar_->width);
        Height(pAvCodecPar_->height);
        SampleRate(pAvCodecPar_->sample_rate);
        Channels(pAvCodecPar_->channels);

        // copy packet
        pAvPacket_ = const_cast<AVPacket*>(_pAvPacket);
}

MediaPacket::MediaPacket()
{
        pAvPacket_ = av_packet_alloc();
        av_init_packet(pAvPacket_);
}

MediaPacket::MediaPacket(IN const AVCodecContext* _pAvCodecContext)
        :MediaPacket()
{
        pAvCodecContext_ = const_cast<AVCodecContext*>(_pAvCodecContext);
}

MediaPacket::~MediaPacket()
{
        if (bRef_) {
                av_packet_unref(pAvPacket_);
        } else {
                av_packet_free(&pAvPacket_);
        }
}

AVPacket* MediaPacket::AvPacket() const
{
        return const_cast<AVPacket*>(pAvPacket_);
}

AVCodecParameters* MediaPacket::AvCodecParameters() const
{
        return pAvCodecPar_;
}

AVCodecContext* MediaPacket::AvCodecContext() const
{
        return pAvCodecContext_;
}

uint64_t MediaPacket::Pts() const
{
        return pAvPacket_->pts;
}

void MediaPacket::Pts(uint64_t _pts)
{
        pAvPacket_->pts = _pts;
}

uint64_t MediaPacket::Dts() const
{
        return pAvPacket_->dts;
}

void MediaPacket::Dts(uint64_t _dts)
{
        pAvPacket_->dts = _dts;
}

StreamType MediaPacket::Stream() const
{
        return stream_;
}

void MediaPacket::Stream(StreamType _type)
{
        stream_ = _type;
}

CodecType MediaPacket::Codec() const
{
        return codec_;
}

void MediaPacket::Codec(CodecType _type)
{
        codec_ = _type;
}

char* MediaPacket::Data()const
{
        return reinterpret_cast<char*>(pAvPacket_->data);
}

int MediaPacket::Size() const
{
        return static_cast<int>(pAvPacket_->size);
}

void MediaPacket::Print() const
{
        Info("packet: pts=%lu, dts=%lu, stream=%d, codec=%d, size=%lu %s",
             static_cast<unsigned long>(pAvPacket_->pts), static_cast<unsigned long>(pAvPacket_->dts),
             Stream(), Codec(), static_cast<unsigned long>(pAvPacket_->size), IsKey() ? "KEY": "");
}

void MediaPacket::Dump(const std::string& _title) const
{
        Debug("%spts=%lu, dts=%lu, stream=%d, codec=%d, size=%lu", _title.c_str(),
              static_cast<unsigned long>(pAvPacket_->pts), static_cast<unsigned long>(pAvPacket_->dts),
              Stream(), Codec(), static_cast<unsigned long>(pAvPacket_->size));
        global::PrintMem(Data(), Size());
}

int MediaPacket::Width() const
{
        return nWidth_;
}

int MediaPacket::Height() const
{
        return nHeight_;
}

void MediaPacket::Width(int _nValue)
{
        nWidth_ = _nValue;
}

void MediaPacket::Height(int _nValue)
{
        nHeight_ = _nValue;
}

int MediaPacket::SampleRate() const
{
        return nSampleRate_;
}

int MediaPacket::Channels() const
{
        return nChannels_;
}

void MediaPacket::SampleRate(int _nValue)
{
        nSampleRate_ = _nValue;
}

void MediaPacket::Channels(int _nValue)
{
        nChannels_ = _nValue;
}

bool MediaPacket::IsKey() const
{
        return ((pAvPacket_->flags & AV_PKT_FLAG_KEY) != 0);
}

void MediaPacket::SetKey()
{
        pAvPacket_->flags |= AV_PKT_FLAG_KEY;
}


//
// Statistic
//

Statistic::Statistic()
        : nInBytes_(0),
          nOutBytes_(0),
          nCounts_(0)
{
        bThreadExit_.store(false);
        stat_ = std::thread([this]() {
                int nInBytes = 0, nOutBytes = 0, nCount = 0;
                while (bThreadExit_.load() == false) {
                        std::this_thread::sleep_for(std::chrono::seconds(nStatPeriod));
                        GetStat(nInBytes, nOutBytes, nCount);
                        Info("STAT: in=%dkbps (%d bytes), out=%dkbps (%d bytes), pumper_concurrent=%d",
                                nInBytes * 8 / (1024 * nStatPeriod), nInBytes,
                                nOutBytes * 8 / (1024 * nStatPeriod), nOutBytes,
                                nCount);
                nInBytes = nOutBytes = nCount = 0;
                }
        });
}

Statistic::~Statistic()
{
        bThreadExit_ = true;
        if (stat_.joinable()) {
                stat_.join();
        }
        return;
}

void Statistic::IncInBytes(IN int _nBytes)
{
        nInBytes_ += _nBytes;
}

void Statistic::IncOutBytes(IN int _nBytes)
{
        nOutBytes_ += _nBytes;
}

void Statistic::IncCount()
{
        nCounts_++;
}

void Statistic::DecCount()
{
        nCounts_--;
}

void Statistic::GetStat(OUT int& _nInBytes,OUT int& _nOutBytes, OUT int& _nCounts)
{
        _nInBytes = nInBytes_.exchange(0);
        _nOutBytes = nOutBytes_.exchange(0);
        _nCounts = nCounts_;
}


//
// FileSink
//

FileSink::FileSink(IN const std::shared_ptr<SharedQueue<std::vector<char>>> _pPool,
                   IN int _nBlockSize, IN int _nXspeed, IN std::string _fmt)
        :pPool_(_pPool),
         nBlockSize_(_nBlockSize),
         nXspeed_(_nXspeed),
         probeQ_(100)
{
        if (_fmt == "flv") {
                path_ = "output.flv";
                nTimeBase_ = FASTFWD_FLV_TIME_BASE;

        } else if (_fmt == "fmp4") {
                path_ = "output.mp4";
                nTimeBase_ = FASTFWD_FMP4_TIME_BASE;
        } else if (_fmt == "ts") {
                path_ = "output.ts";
                nTimeBase_ = FASTFWD_TS_TIME_BASE;
        }
        nMinGOP_ = nTimeBase_ * 8;
}

FileSink::~FileSink()
{
        if (pOutputContext_ != nullptr) {
                // write trailer first
                av_write_trailer(pOutputContext_);

                // destroy contexts
                avformat_free_context(pOutputContext_);
                pOutputContext_ = nullptr;
        }

        if (pAvIoContext_) {
                av_freep(&pAvIoContext_->buffer);
                av_freep(&pAvIoContext_);
        }

        av_bsf_free(&pAacBsf_);
}

int FileSink::Write(IN const std::shared_ptr<MediaPacket>& _pPacket)
{
        // initialize contexts
        if (Init() == false) {
                return -1;
        }

        // push packets into the buffer queue
        if (probeQ_.TryPush(_pPacket) == false) {
                Error("file sink: expected stream was not detected, video=%d audio=%d", bVideoAdded_, bAudioAdded_);
                return -5;
        }

        // if expected streams are ready, start to write packets
        if (bVideoAdded_) {
                WritePackets(nXspeed_);
        } else {
                // add video and audio streams to context
                switch (_pPacket->Stream()) {
                case STREAM_AUDIO:
                        if (!bAudioAdded_) {
                                streams_[static_cast<int>(STREAM_AUDIO)] = AddStream(_pPacket);
                                bAudioAdded_ = true;
                        }
                        break;
                case STREAM_VIDEO:
                        if (!bVideoAdded_) {
                                streams_[static_cast<int>(STREAM_VIDEO)] = AddStream(_pPacket);
                                bVideoAdded_ = true;
                        }
                        break;
                default:
                        ;
                }

                // write file header
                if (bVideoAdded_ && WriteHeader() == false) {
                        return -2;
                }
        }

        return 0;
}

bool FileSink::WritePackets(IN int _nXspeed)
{
        // handle packet queue
        do {
                // get from queue
                std::shared_ptr<MediaPacket> pPkt;
                if (probeQ_.TryPop(pPkt) == false) {
                        break;
                }
                auto pAvPkt = pPkt->AvPacket();

                // handle abnoraml pts
                if (pAvPkt->pts < 0) {
                        Warn("file sink: abnormal pts value %ld, drop", pAvPkt->pts);
                        continue;
                }

                /* Frame rate control principle:
                 * 1. We assume that the GOP of one stream is constant and the maximum of GOP is 8 second.
                 * 2. nCurrentGOP is interval of current frame's PTS and last frame's PTS. nMinGOP_ is the
                 * minimum GOP of a stream. if nCurrentGOP is larger than two time of nMinGOP_, we assume
                 * that the frame is not in the same time slot as the last frame and fitted the PTS as
                 * last frame's PTS + nMinGOP_ / _nXspeed.
                */
                if (bIsFirstPkt_) {
                        bIsFirstPkt_ = false;
                        nLastPtsOriginal_ = pAvPkt->pts;
                        pAvPkt->pts = 0;
                } else {
                        long long nCurrentGOP = pAvPkt->pts - nLastPtsOriginal_;
                        nLastPtsOriginal_ = pAvPkt->pts;
                        if (nCurrentGOP > nMinGOP_ * 2) {
                                pAvPkt->pts = nLastPtsFitted_ + nMinGOP_ / _nXspeed;
                        } else {
                                nMinGOP_ = (nCurrentGOP < nMinGOP_) ? nCurrentGOP : nMinGOP_;
                                pAvPkt->pts = nLastPtsFitted_ + nCurrentGOP / _nXspeed;
                        }
                        nLastPtsFitted_ = pAvPkt->pts;
                }
                pAvPkt->dts = pAvPkt->pts;
                Debug("count: %ld, pts: %ld", nCount_, pAvPkt->pts);

                // handle stream index
                pAvPkt->stream_index = streams_[static_cast<int>(pPkt->Stream())];

                // create bsf
                switch (pPkt->Stream()) {
                case STREAM_AUDIO: {
                        int nPosition = path_.length() - 4;
                        if (nPosition > 0 && path_.substr(nPosition).compare(".flv") == 0) {
                                if (pAacBsf_ == nullptr) {
                                        auto filter = av_bsf_get_by_name("aac_adtstoasc");
                                        av_bsf_alloc(filter, &pAacBsf_);
                                        auto pStream = pOutputContext_->streams[pAvPkt->stream_index];
                                        avcodec_parameters_copy(pAacBsf_->par_in, pStream->codecpar);
                                        av_bsf_init(pAacBsf_);
                                }
                                av_bsf_send_packet(pAacBsf_, pAvPkt);
                                while (av_bsf_receive_packet(pAacBsf_, pAvPkt) == 0);
                        }
                        break;
                }
                default:
                        ;
                }

                // write to file
                int nStatus = av_interleaved_write_frame(pOutputContext_, pAvPkt);
                if (nStatus < 0) {
                        Warn("file sink: write failed: %d", nStatus);
                }
                nCount_++;

        } while(true);

        return true;
}

bool FileSink::Init()
{
        // initialize context
        if (pOutputContext_ == nullptr) {
                // create output context
                avformat_alloc_output_context2(&pOutputContext_, nullptr, nullptr, path_.c_str());
                if (pOutputContext_ == nullptr) {
                        Error("file sink: could not create context for file: %s", path_.c_str());
                        return false;
                }

                pMemBuffer_ = (uint8_t*)av_malloc(nBlockSize_);
                pAvIoContext_ = avio_alloc_context(pMemBuffer_, nBlockSize_, 1, this, nullptr, WriteFunction, nullptr);
                pOutputContext_->pb = pAvIoContext_;

                AVOutputFormat* pFormat = nullptr;
                if (path_ == "output.flv") {
                        pFormat = av_guess_format("flv", nullptr, nullptr);
                } else if (path_ == "output.mp4") {
                        pFormat = av_guess_format("mp4", nullptr, nullptr);
                } else if (path_ == "output.ts") {
                        pFormat = av_guess_format("mpegts", nullptr, nullptr);
                }

                if (pFormat == nullptr) {
                        Error("format not found");
                        return false;
                }
                pOutputContext_->oformat = pFormat;
        }

        return true;
}

int FileSink::WriteFunction(IN void *_pOpaque, IN uint8_t* _pBuf, IN int _nSize)
{
        FileSink* pThis = (FileSink*)_pOpaque;

        // check if pool is valid
        if (pThis->pPool_ == nullptr) {
                Error("pool is null");
                return 0;
        }

        char* pBuf = (char*)_pBuf;
        pThis->pPool_->Push(std::vector<char>(pBuf, pBuf + _nSize));

        return _nSize;
}

int FileSink::AddStream(IN const std::shared_ptr<MediaPacket>& _pPacket)
{
        AVStream* pStream = avformat_new_stream(pOutputContext_, nullptr);
        if (pStream == nullptr) {
                Error("file sink: failed allocating output stream");
                return -1;
        }

        if (_pPacket->AvCodecParameters() == nullptr) {
                Error("file sink: source codec parameter pointer is null");
                return -2;
        }

        auto nStatus = avcodec_parameters_copy(pStream->codecpar, _pPacket->AvCodecParameters());
        if (nStatus < 0) {
                Error("file sink: failed to copy codec parameters");
                return -3;
        }
        pStream->codecpar->codec_tag = 0;

        return nStreamIndex_++;
}

bool FileSink::WriteHeader()
{
        AVDictionary *pOption = nullptr;
        if (path_ == "output.mp4") {
                av_dict_set(&pOption, "movflags", "frag_keyframe+empty_moov", 0);
        }
        // write file header according to the file suffix
        auto nStatus = avformat_write_header(pOutputContext_, &pOption);
        if (nStatus < 0) {
                Error("file sink: could not write file header: %d", nStatus);
                return false;
        }

        return true;
}

//
// AvReceiver
//

AvReceiver::AvReceiver()
{
}

AvReceiver::~AvReceiver()
{
        avformat_close_input(&pAvContext_);
        pAvContext_ = nullptr;
}

int AvReceiver::AvInterruptCallback(void* _pContext)
{
        using namespace std::chrono;
        AvReceiver* pReceiver = reinterpret_cast<AvReceiver*>(_pContext);
        high_resolution_clock::time_point now = high_resolution_clock::now();
        if (duration_cast<milliseconds>(now - pReceiver->start_).count() > pReceiver->nTimeout_) {
                Error("receiver timeout, %lu milliseconds", pReceiver->nTimeout_);
                return -1;
        }

        return 0;
}

int AvReceiver::Receive(IN const std::string& _url, IN int _nXspeed, IN std::string _fmt, IN PacketHandlerType& _callback)
{
        if (pAvContext_ != nullptr) {
                Warn("internal: reuse of Receiver is not recommended");
        }

        // allocate AV context
        pAvContext_ = avformat_alloc_context();
        if (pAvContext_ == nullptr) {
                Error("av context could not be created");
                return -1;
        }

        // for timeout timer
        std::string option;
        nTimeout_ = 10 * 1000; // 10 seconds
        Warn("receiver timeout=%lu milliseconds", nTimeout_);
        pAvContext_->interrupt_callback.callback = AvReceiver::AvInterruptCallback;
        pAvContext_->interrupt_callback.opaque = this;
        start_ = std::chrono::high_resolution_clock::now();

        // open input stream
        Debug("input URL: %s", _url.c_str());
        int nStatus = avformat_open_input(&pAvContext_, _url.c_str(), 0, 0);
        if (nStatus < 0) {
                static char str[500];
                memset(str, 0, sizeof(str));
                std::string err = av_make_error_string(str, 500, nStatus);
                Error("could not open input stream: %s: %s", _url.c_str(), err.c_str());
                return -1;
        }

        // get stream info
        nStatus = avformat_find_stream_info(pAvContext_, 0);
        if (nStatus < 0) {
                Error("could not get stream info");
                return -1;
        }
        for (unsigned int i = 0; i < pAvContext_->nb_streams; i++) {
                struct AVStream * pAvStream = pAvContext_->streams[i];
                streams_.push_back(StreamInfo{pAvStream, -1});
                Debug("stream is found: avstream=%d, avcodec=%d",
                     pAvStream->codecpar->codec_type, pAvStream->codecpar->codec_id);
        }

        AVPacket avPacket;
        av_init_packet(&avPacket);

        while (true) {
                nStatus = av_read_frame(pAvContext_, &avPacket);
                if (nStatus < 0) {
                        return nStatus;
                }

                // statistic input bytes
                pStat_->IncInBytes(avPacket.size);

                if (avPacket.stream_index < 0 ||
                    static_cast<unsigned int>(avPacket.stream_index) >= pAvContext_->nb_streams) {
                        Warn("invalid stream index in packet");
                        av_packet_unref(&avPacket);
                        continue;
                }

                // only mux with keyframe
                auto streamType = static_cast<StreamType>(streams_[avPacket.stream_index].pAvStream->codecpar->codec_type);
                if ((streamType != STREAM_VIDEO) || !(avPacket.flags & AV_PKT_FLAG_KEY)) {
                        av_packet_unref(&avPacket);
                        continue;
                }

                // if avformat detects another stream during transport, we have to ignore the packets of the stream
                if (static_cast<size_t>(avPacket.stream_index) < streams_.size()) {
                        // we need all PTS/DTS use milliseconds, sometimes they are macroseconds such as TS streams
                        AVRational tb;
                        if (_fmt == "flv") {
                                tb = AVRational{1, FASTFWD_FLV_TIME_BASE};
                        } else if (_fmt == "fmp4") {
                                tb = AVRational{1, FASTFWD_FMP4_TIME_BASE};
                        } else if (_fmt == "ts") {
                                tb = AVRational{1, FASTFWD_TS_TIME_BASE};
                        }

                        AVRounding r = static_cast<AVRounding>(AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
                        avPacket.dts = av_rescale_q_rnd(avPacket.dts, streams_[avPacket.stream_index].pAvStream->time_base, tb, r);
                        avPacket.pts = av_rescale_q_rnd(avPacket.pts, streams_[avPacket.stream_index].pAvStream->time_base, tb, r);

                        // emulate framerate @ 1.0x speed
                        if (EmulateFramerate(avPacket.pts, streams_[avPacket.stream_index], _nXspeed) == true) {
                                int nStatus = _callback(std::make_shared<MediaPacket>(*streams_[avPacket.stream_index].pAvStream,
                                                                                      &avPacket));
                                if (nStatus != 0) {
                                        return nStatus;
                                }
                        } else {
                                av_packet_unref(&avPacket);
                        }
                } else {
                        av_packet_unref(&avPacket);
                }
                start_ = std::chrono::high_resolution_clock::now();
        }

        return 0;
}

bool AvReceiver::EmulateFramerate(IN int64_t _nPts, OUT StreamInfo& _stream, IN int _nXspeed)
{
        if (_nPts < 0) {
                Warn("receiver: minus pts received pts=%ld, drop", _nPts);
                return false;
        }

        using namespace std::chrono;
        if (_stream.nCount < 0) {
                _stream.start = high_resolution_clock::now();
                _stream.nCount = 0;
        }
        high_resolution_clock::time_point now = high_resolution_clock::now();
        int64_t nPlaytime = _stream.nCount * 2000 / _nXspeed;
        auto nDuration = duration_cast<milliseconds>(now - _stream.start).count();
        if (nPlaytime > nDuration) {
                auto delay = nPlaytime - nDuration;
                if (delay > 10000) {
                        Warn("receiver: fps emulation: delay > 10s (delay=%ld), skip", delay);
                } else {
                        std::this_thread::sleep_for(std::chrono::milliseconds(1000));
                }
        }

        _stream.nCount++;

        Debug("wall_clock=%ld vs play_clock=%ld", nDuration, nPlaytime);

        return true;
}


//
// StreamPumper
//

StreamPumper::StreamPumper(IN const std::string& _url, IN int _nXspeed, IN const std::string& _fmt, IN int _nBlockSize)
        :url_(_url),
         nXspeed_(_nXspeed),
         fmt_(_fmt),
         nBlockSize_(_nBlockSize)
{
        std::call_once(avformatInit_, [](){
                        avformat_network_init();
                        av_log_set_level(AV_LOG_WARNING);
                });

        bPumperStopped_.store(false);
        pPool_ = std::make_shared<SharedQueue<std::vector<char>>>(); // TODO, notice memory exaustion

        // validate x speed
        auto bFound = false;
        std::vector<int> validSpeed{2, 4, 8, 16, 32};
        for (auto& n : validSpeed) {
                if (n == _nXspeed) {
                        bFound = true;
                        break;
                }
        }
        if (!bFound) {
                nXspeed_ = fastfwd::x2;                
        }

        // validate format
        bFound = false;
        std::vector<std::string> validFmt{"fmp4", "flv", "ts"};
        for (auto& s : validFmt) {
            if (s == _fmt) {
                bFound = true;
                break;
            }
        }

        if(!bFound) {
            fmt_ = "flv";
        }

        // validate block size
        if (_nBlockSize < 4096) {
                nBlockSize_ = 4096;
        }

        pStat_->IncCount();
}

StreamPumper::~StreamPumper()
{
        StopPumper();
        pStat_->DecCount();
}

void StreamPumper::StartPumper()
{
        pSink_ = std::make_unique<FileSink>(pPool_, nBlockSize_, nXspeed_, fmt_);
        pReceiver_ = std::make_unique<AvReceiver>();

        auto pumper = [this]() {

                // input receiver callback will get each packet from input sources and put them into
                // file sink, which is responsible for muxing header and body, then push in the pool
                auto recv = [this](IN const std::shared_ptr<MediaPacket> _pPacket) -> int {
                        if (bPumperStopped_.load() == true) {
                                return -1;
                        }

                        //_pPacket->Print();

                        if (pSink_->Write(_pPacket) < 0) {
                                return -1;
                        }

                        return 0;
                };

                while (bPumperStopped_.load() != true) {
                        auto nStatus = pReceiver_->Receive(url_, nXspeed_, fmt_, recv);
                        if (nStatus == AVERROR_EOF) {
                                nStatus_.store(StreamPumper::eof);
                        } else {
                                nStatus_.store(StreamPumper::econn);
                        }
                        break;
                }
        };

        pumper_ = std::thread(pumper);
}

void StreamPumper::StopPumper()
{
        bPumperStopped_.store(true);
        if (pumper_.joinable()) {
                pumper_.join();
        }
}

int StreamPumper::Pump(OUT std::vector<char>& _stream, IN int _nTimeout)
{
        std::lock_guard<std::mutex> lock(pumpLck_);

        // start engine
        if (!bStarted_) {
                StartPumper();
                bStarted_ = true;
        }

        // return streams if any
        if (pPool_->PopWithTimeout(_stream, std::chrono::milliseconds(_nTimeout)) == false) {
                return nStatus_.load();
        }

        return 0;
}
