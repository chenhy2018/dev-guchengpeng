#include "input.hpp"

using namespace muxer;

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

int AvReceiver::Receive(IN const std::string& _url, IN PacketHandlerType& _callback)
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
        Info("receiver timeout=%lu milliseconds", nTimeout_);
        pAvContext_->interrupt_callback.callback = AvReceiver::AvInterruptCallback;
        pAvContext_->interrupt_callback.opaque = this;
        start_ = std::chrono::high_resolution_clock::now();

        // open input stream
        Info("input URL: %s", _url.c_str());
        int nStatus = avformat_open_input(&pAvContext_, _url.c_str(), 0, 0);
        if (nStatus < 0) {
                Error("could not open input stream: %s", _url.c_str());
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
                Info("stream is found: avstream=%d, avcodec=%d",
                     pAvStream->codecpar->codec_type, pAvStream->codecpar->codec_id);
        }

        AVPacket avPacket;
        av_init_packet(&avPacket);

        while (av_read_frame(pAvContext_, &avPacket) >= 0) {
                if (avPacket.stream_index < 0 ||
                    static_cast<unsigned int>(avPacket.stream_index) >= pAvContext_->nb_streams) {
                        Warn("invalid stream index in packet");
                        av_packet_unref(&avPacket);
                        continue;
                }

                // if avformat detects another stream during transport, we have to ignore the packets of the stream
                if (static_cast<size_t>(avPacket.stream_index) < streams_.size()) {
                        // we need all PTS/DTS use milliseconds, sometimes they are macroseconds such as TS streams
                        AVRational tb = AVRational{1, 1000};
                        AVRounding r = static_cast<AVRounding>(AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
                        avPacket.dts = av_rescale_q_rnd(avPacket.dts, streams_[avPacket.stream_index].pAvStream->time_base, tb, r);
                        avPacket.pts = av_rescale_q_rnd(avPacket.pts, streams_[avPacket.stream_index].pAvStream->time_base, tb, r);

                        // emulate framerate @ 1.0x speed
                        if (EmulateFramerate(avPacket.pts, streams_[avPacket.stream_index]) == true) {
                                int nStatus = _callback(std::make_unique<MediaPacket>(*streams_[avPacket.stream_index].pAvStream,
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

bool AvReceiver::EmulateFramerate(IN int64_t _nPts, OUT StreamInfo& _stream)
{
        if (_nPts < 0) {
                Warn("receiver: minus pts received pts=%ld, drop", _nPts);
                return false;
        }

        using namespace std::chrono;
        if (_stream.nFirstPts < 0) {
                _stream.start = high_resolution_clock::now();
                _stream.nFirstPts = _nPts;
        }
        high_resolution_clock::time_point now = high_resolution_clock::now();
        int64_t nPlaytime = _nPts - _stream.nFirstPts;
        auto nDuration = duration_cast<milliseconds>(now - _stream.start).count();
        if (nPlaytime > nDuration) {
                auto delay = nPlaytime - nDuration;
                if (delay > 10000) {
                        Warn("receiver: fps emulation: delay > 10s (delay=%ld), skip", delay);
                } else {
                        msleep(delay);
                }
        }

        return true;
}

//
// AvDecoder
//

AvDecoder::AvDecoder()
{
}

AvDecoder::~AvDecoder()
{
        if (bIsDecoderAvailable_) {
                avcodec_close(pAvDecoderContext_);
        }
        if (pAvDecoderContext_ != nullptr) {
                avcodec_free_context(&pAvDecoderContext_);
        }
}

int AvDecoder::Init(IN const std::unique_ptr<MediaPacket>& _pPacket)
{
        // create decoder
        if (pAvDecoderContext_ == nullptr) {
                // find decoder
                AVCodec *pAvCodec = avcodec_find_decoder(static_cast<AVCodecID>(_pPacket->Codec()));
                if (pAvCodec == nullptr) {
                        Error("could not find AV decoder for codec_id=%d", _pPacket->Codec());
                        return -1;
                }

                // initiate AVCodecContext
                pAvDecoderContext_ = avcodec_alloc_context3(pAvCodec);
                if (pAvDecoderContext_ == nullptr) {
                        Error("could not allocate AV codec context");
                        return -1;
                }

                // if the packet is from libavformat
                // just use context parameters in AVStream to get one directly otherwise fake one
                if (_pPacket->AvCodecParameters() != nullptr) {
                        if (avcodec_parameters_to_context(pAvDecoderContext_, _pPacket->AvCodecParameters()) < 0){
                                Error("could not copy decoder context");
                                return -1;
                        }
                }

                // open it
                if (avcodec_open2(pAvDecoderContext_, pAvCodec, nullptr) < 0) {
                        Error("could not open decoder");
                        return -1;
                } else {
                        Info("open decoder: stream=%d, codec=%d", _pPacket->Stream(), _pPacket->Codec());
                        bIsDecoderAvailable_ = true;
                }
        }

        return 0;
}

int AvDecoder::Decode(IN const std::unique_ptr<MediaPacket>& _pPacket, IN FrameHandlerType& _callback)
{
        if (Init(_pPacket) < 0) {
                return -1;
        }

        int nStatus;

        //
        // decode ! and get one frame to encode
        //
        do {
                bool bNeedSendAgain = false;
                int nStatus = avcodec_send_packet(pAvDecoderContext_, _pPacket->AvPacket());
                if (nStatus != 0) {
                        if (nStatus == AVERROR(EAGAIN)) {
                                Warn("decoder internal: assert failed, we should not get EAGAIN");
                                bNeedSendAgain = true;
                        } else {
                                Error("decoder: could not send frame, status=%d", nStatus);
                                _pPacket->Print();
                                return -1;
                        }
                }

                while (1) {
                        // allocate a frame for outputs
                        auto pFrame = std::make_shared<MediaFrame>();
                        pFrame->Stream(_pPacket->Stream());
                        pFrame->Codec(_pPacket->Codec());

                        nStatus = avcodec_receive_frame(pAvDecoderContext_, pFrame->AvFrame());
                        if (nStatus == 0) {
                                int nStatus = _callback(pFrame);
                                if (nStatus < 0) {
                                        return nStatus;
                                }
                                if (bNeedSendAgain) {
                                        break;
                                }
                        } else if (nStatus == AVERROR(EAGAIN)) {
                                return 0;
                        } else {
                                Error("decoder: could not receive frame, status=%d", nStatus);
                                _pPacket->Print();
                                return -1;
                        }
                }
        } while(1);

        return 0;
}

//
// Input
//

Input::Input(IN const std::string& _name)
        :OptionMap(),
         name_(_name)
{
}

Input::~Input()
{
}

std::string Input::Name()
{
        return name_;
}

void Input::AttachMuxer(IN const std::string& _mux)
{
        attachedMuxer_ = _mux;
}

std::string& Input::AttachedMuxer()
{
        return attachedMuxer_;
}

void Input::DetachMuxer()
{
        attachedMuxer_ = "";
}

void Input::PresetVideo(INOUT std::shared_ptr<MediaFrame>& _pFrame)
{
}

void Input::ProsetVideo(INOUT std::shared_ptr<MediaFrame>& _pFrame)
{
        int nSize = 5, nColor = 0;
        if (GetOption(options::bdcolor, nColor) == true || GetOption(options::bdsize, nSize) == true) {
                avFilter.AddBorderFilter(nSize, nColor);
        }
        avFilter.GetPicture(_pFrame);

        // set x,y,z coordinate
        int nX, nY, nZ;
        if (GetOption(options::x, nX) == true) {
                _pFrame->X(nX);
        }
        if (GetOption(options::y, nY) == true) {
                _pFrame->Y(nY);
        }
        if (GetOption(options::z, nZ) == true) {
                _pFrame->Z(nZ);
        }
}

//
// Stream
//

Stream::Stream(IN const std::string& _name)
        :Input(_name),
         videoQ_(Stream::VIDEO_Q_LEN),
         audioQ_(Stream::AUDIO_Q_LEN)
{
        bReceiverExit_.store(false);

        // reserve sample buffer pool for format S16
        sampleBuffer_.reserve(AudioResampler::FRAME_SIZE * AudioResampler::CHANNELS * 2 * 4);
        sampleBuffer_.resize(0);
}

// start thread => receiver loop => decoder loop
void Stream::Start(IN const std::string& _url)
{
        stat.Start(1);

        if (_url.empty()) {
                // receiver thread won't be up
                bReceiverExit_.store(true);
        }

        auto recv = [this, _url] {
                while (bReceiverExit_.load() == false) {
                        auto avReceiver = std::make_unique<AvReceiver>();
                        auto vDecoder = std::make_unique<AvDecoder>();
                        auto aDecoder = std::make_unique<AvDecoder>();

                        auto receiverHook = [&](IN const std::unique_ptr<MediaPacket> _pPacket) -> int {
                                stat.OneSample(_pPacket->Size());

                                if (bReceiverExit_.load() == true) {
                                        return -1;
                                }

                                auto decoderHook = [&](const std::shared_ptr<MediaFrame>& _pFrame) -> int {
                                        if (bReceiverExit_.load() == true) {
                                                return -1;
                                        }

                                        // store the last picture
                                        if (_pFrame->Stream() == STREAM_VIDEO) {
                                                std::lock_guard<std::mutex> lock(lastVideoLck_);
                                                pLastVideo_ = _pFrame;
                                        }

                                        // copy packets for each input clones
                                        clones_.Foreach([_pFrame](const std::string& _clone, std::shared_ptr<Stream> _pStream) {
                                                        _pStream->SetFrame(std::make_shared<MediaFrame>(*_pFrame));
                                                });

                                        return 0;
                                };

                                // start decoder loop
                                int nStatus = 0;
                                if (_pPacket->Stream() == STREAM_VIDEO) {
                                        nStatus = vDecoder->Decode(_pPacket, decoderHook);
                                } else if (_pPacket->Stream() == STREAM_AUDIO) {
                                        nStatus = aDecoder->Decode(_pPacket, decoderHook);
                                }

                                return nStatus;
                        };

                        // start receiver loop
                        avReceiver->Receive(_url, receiverHook);

                        // prevent receiver reconnecting too fast
                        std::this_thread::sleep_for(std::chrono::milliseconds(500));
                }
        };

        receiver_ = std::thread(recv);
}

void Stream::Stop()
{
        bReceiverExit_.store(true);
        if (receiver_.joinable()) {
                receiver_.join();
        }
}

void Stream::SetFrame(const std::shared_ptr<MediaFrame>& _pFrame)
{
        // format video/audio data
        if (_pFrame->Stream() == STREAM_VIDEO) {
                SetVideo(_pFrame);
        } else if (_pFrame->Stream() == STREAM_AUDIO) {
                SetAudio(_pFrame);
        }
}

void Stream::SetVideo(const std::shared_ptr<MediaFrame>& _pFrame)
{
        // TODO: move rescaler to avfilter or PresetVideo so that text inputs can support width,height options

        // stretch mode
        int nColor = 0x0; // black
        bool bStretch = false; // by default
        if (GetOption(options::stretch)) {
                bStretch = true;
        }
        if (!bStretch) {
                GetOption(options::bgcolor, nColor);
        }

        // convert color space to YUV420p and rescale the image
        int nW, nH;
        bool bNeedRescale = false;
        AVPixelFormat colorspace = VideoRescaler::PIXEL_FMT;
        if (_pFrame->AvFrame()->format == AV_PIX_FMT_ARGB || _pFrame->AvFrame()->format == AV_PIX_FMT_ABGR ||
            _pFrame->AvFrame()->format == AV_PIX_FMT_RGBA || _pFrame->AvFrame()->format == AV_PIX_FMT_BGRA) {
                colorspace = AV_PIX_FMT_YUVA420P; // save alpha channel
        }
        if (pRescaler_ == nullptr) {
                auto pAvf = _pFrame->AvFrame();
                if (GetOption(options::width, nW) == true && nW != pAvf->width) {
                        bNeedRescale = true;
                }
                if (GetOption(options::height, nH) == true && nH != pAvf->height) {
                        bNeedRescale = true;
                }
                if (pAvf->format != colorspace) {
                        bNeedRescale = true;
                }
                if (bNeedRescale) {
                        pRescaler_ = std::make_shared<VideoRescaler>(nW, nH, colorspace, bStretch, nColor);
                }
        } else {
                // if target w or h is changed, reinit the rescaler
                if (GetOption(options::width, nW) == true && nW != pRescaler_->TargetW()) {
                        bNeedRescale = true;
                }
                if (GetOption(options::height, nH) == true && nH != pRescaler_->TargetH()) {
                        bNeedRescale = true;
                }
                if (bNeedRescale) {
                        pRescaler_->Reset(nW, nH, colorspace, bStretch, nColor);
                }
        }

        // rescale the video frame
        auto pFrame = _pFrame;
        if (pRescaler_ != nullptr) {
                pRescaler_->Rescale(_pFrame, pFrame);
        }

        // send to video filter
        PresetVideo(pFrame);

        videoQ_.ForcePush(pFrame);
}

void Stream::SetAudio(const std::shared_ptr<MediaFrame>& _pFrame)
{
        // resample the audio and push data to the buffer

        // lock buffer and queue
        std::lock_guard<std::mutex> lock(sampleBufferLck_);

        size_t nBufSize = sampleBuffer_.size();
        std::vector<uint8_t> buffer;

        // resample to the same audio format
        if (resampler_.Resample(_pFrame, buffer) != 0) {
                return;
        }

        // detect user option for audio input gain
        int nVol;
        if (GetOption(options::volume, nVol) == true) {
                if (nVol >= 0) {
                        sound::Gain(buffer.data(), buffer.size(), nVol);
                }
        }

        // save resampled data in audio buffer
        sampleBuffer_.resize(sampleBuffer_.size() + buffer.size());
        std::copy(buffer.begin(), buffer.end(), &sampleBuffer_[nBufSize]);

        // if the buffer size meets the min requirement of encoding one frame, build a frame and push upon audio queue
        size_t nSizeEachFrame = AudioResampler::FRAME_SIZE * AudioResampler::CHANNELS * av_get_bytes_per_sample(AudioResampler::SAMPLE_FMT);

        while (sampleBuffer_.size() >= nSizeEachFrame) {
                auto pNewFrame = std::make_shared<MediaFrame>(AudioResampler::FRAME_SIZE, AudioResampler::CHANNELS, AudioResampler::SAMPLE_FMT);
                pNewFrame->Stream(_pFrame->Stream());
                pNewFrame->Codec(_pFrame->Codec());
                pNewFrame->AvFrame()->sample_rate = AudioResampler::SAMPLE_RATE;
                std::copy(&sampleBuffer_[0], &sampleBuffer_[nSizeEachFrame], pNewFrame->AvFrame()->data[0]);

                // move rest samples to beginning of the buffer
                std::copy(&sampleBuffer_[nSizeEachFrame], &sampleBuffer_[sampleBuffer_.size()], sampleBuffer_.begin());
                sampleBuffer_.resize(sampleBuffer_.size() - nSizeEachFrame);

                audioQ_.ForcePush(pNewFrame);
        }
}

bool Stream::GetVideo(OUT std::shared_ptr<MediaFrame>& _pFrame)
{
        // make sure preserve at least one video frame
        bool bOk = videoQ_.TryPop(pLastVideo_);
        if (bOk == true){
                _pFrame = pLastVideo_;
        } else {
                if (pLastVideo_ == nullptr) {
                        // before we get the first renderable frame, if background option is set
                        // just show the color block
                        int nColor, nW, nH;
                        if (GetOption(options::bgcolor, nColor) == false ||
                            GetOption(options::width, nW) == false ||
                            GetOption(options::height, nH) == false) {
                                return false;
                        } else {
                                _pFrame = std::make_shared<MediaFrame>(nW, nH, VideoRescaler::PIXEL_FMT, nColor);
                                _pFrame->Stream(STREAM_VIDEO);
                                _pFrame->Codec(CODEC_H264);
                        }
                } else {
                        _pFrame = pLastVideo_;
                }
        }

        ProsetVideo(_pFrame);

        return true;
}

bool Stream::GetAudio(OUT std::shared_ptr<MediaFrame>& _pFrame)
{
        // lock buffer and queue
        std::lock_guard<std::mutex> lock(sampleBufferLck_);
        size_t nSizeEachFrame = AudioResampler::FRAME_SIZE * AudioResampler::CHANNELS * av_get_bytes_per_sample(AudioResampler::SAMPLE_FMT);

        bool bOk = audioQ_.TryPop(_pFrame);
        if (bOk == true) {
                return true;
        } else {
                if (sampleBuffer_.size() > 0 && sampleBuffer_.size() < nSizeEachFrame) {
                        // sample buffer does contain data but less than frame size
                        _pFrame = std::make_shared<MediaFrame>(AudioResampler::FRAME_SIZE, AudioResampler::CHANNELS,
                                                               AudioResampler::SAMPLE_FMT, true);
                        _pFrame->Stream(STREAM_AUDIO);
                        _pFrame->Codec(CODEC_AAC);
                        _pFrame->AvFrame()->sample_rate = AudioResampler::SAMPLE_RATE;

                        // copy existing buffer contents
                        std::copy(&sampleBuffer_[0], &sampleBuffer_[sampleBuffer_.size()], _pFrame->AvFrame()->data[0]);
                        sampleBuffer_.resize(0);
                        return true;
                }
        }

        return false;
}

Stream::~Stream()
{
        Stop();
}

std::shared_ptr<Input> Stream::CreateClone(IN const std::string& _clone, IN const Option& _opt)
{
        auto pStream = std::make_shared<Stream>(_clone);
        pStream->GetOptions(_opt);

        // in the scenario of stream disconnection, original input will keep the last picture
        // if we create clone from the input the last rendered picture will be copied as well
        {
                std::lock_guard<std::mutex> lock(lastVideoLck_);
                if (pLastVideo_ != nullptr) {
                        pStream->SetFrame(std::make_shared<MediaFrame>(*pLastVideo_));
                }
        }

        if (clones_.Insert(_clone, pStream) == true) {
                return std::dynamic_pointer_cast<Input>(pStream);
        }

        Error("[%s] stream clone %s already exists", Name().c_str(), _clone.c_str());
        return nullptr;
}

bool Stream::RemoveClone(IN const std::string& _clone)
{
        return clones_.Erase(_clone);
}

std::shared_ptr<Input> Stream::GetClone(IN const std::string& _clone)
{
        std::shared_ptr<Input> pClone = nullptr;
        clones_.Find(_clone, [&pClone](std::shared_ptr<Stream>& _pStream){
                        pClone = std::dynamic_pointer_cast<Input>(_pStream);
                });

        return pClone;
}

std::shared_ptr<Input> Stream::NextClone()
{
        std::shared_ptr<Input> pNext = nullptr;
        clones_.FindIf([&pNext](const std::string& key, std::shared_ptr<Stream>& _pStream) -> bool {
                        pNext = std::dynamic_pointer_cast<Input>(_pStream);
                        return true;
                });
        return pNext;
}

std::string Stream::String()
{
        std::string out = "<" + Name() + "> stream\n";
        clones_.Foreach([&out, this](const std::string _key, std::shared_ptr<Stream> _pClone){
                        out += "  clone=" + _pClone->Name() + " -> " + _pClone->AttachedMuxer() + "\n";
                });

        return out;
}

//
// Text
//

Text::Text(IN const std::string& _name, IN const std::string& _text)
        :Input(_name),
         text_(_text),
         videoQ_(Text::VIDEO_Q_LEN)
{
        bMakerExit_.store(false);
}

Text::~Text()
{
        Stop();
}

std::string Text::String()
{
        std::string out = "<" + Name() + "> text\n";
        clones_.Foreach([&out, this](const std::string _key, std::shared_ptr<Text> _pClone){
                        out += "  clone=" + _pClone->Name() + " -> " + _pClone->AttachedMuxer() + "\n";
                });

        return out;
}

void Text::Start()
{
        auto thread = [this] {
                auto pRender = std::make_shared<TextRender>(text_);
                while (bMakerExit_.load() == false) {
                        int nOffsetX, nOffsetY, nW, nH, nOpacity, nFtColor, nBgColor, nSpeedx, nSpeedy, nSize;
                        std::string font, text;

                        // give default values if not specified
                        if (GetOption(options::text, text) == false) {
                                text = text_;
                        }
                        if (GetOption(options::width, nW) == false) {
                                nW = 100;
                        }
                        if (GetOption(options::height, nH) == false) {
                                nH = 20;
                        }
                        if (GetOption(options::opacity, nOpacity) == false) {
                                nOpacity = 0x7f;
                        }
                        if (GetOption(options::offsetx, nOffsetX) == false) {
                                nOffsetX = 0;
                        }
                        if (GetOption(options::offsety, nOffsetY) == false) {
                                nOffsetY = 0;
                        }
                        if (GetOption(options::bgcolor, nBgColor) == false) {
                                nBgColor = 0x000000;
                        }
                        if (GetOption(options::ftcolor, nFtColor) == false) {
                                nFtColor = 0xffffff;
                        }
                        if (GetOption(options::font, font) == false) {
                                font = "yahei";
                        }
                        if (GetOption(options::speedx, nSpeedx) == false) {
                                nSpeedx = 0;
                        }
                        if (GetOption(options::speedy, nSpeedy) == false) {
                                nSpeedy = 0;
                        }
                        if (GetOption(options::size, nSize) == false) {
                                nSize = 20;
                        }

                        // generate text frame
                        std::shared_ptr<MediaFrame> pFrame = nullptr;
                        if (pRender->GetPicture(text, font, nSize, nFtColor, nOffsetX, nOffsetY,
                                                nW, nH, nBgColor, nOpacity, nSpeedx, nSpeedy, pFrame) == true) {
                                // send to video filter
                                PresetVideo(pFrame);

                                // push in the queue
                                videoQ_.ForcePush(pFrame);
                        }

                        std::this_thread::sleep_for(std::chrono::milliseconds(40));
                }
        };

        picMaker_ = std::thread(thread);
}

void Text::Stop()
{
        bMakerExit_.store(true);
        if (picMaker_.joinable()) {
                picMaker_.join();
        }
}

bool Text::GetVideo(OUT std::shared_ptr<MediaFrame>& _pFrame)
{
        // make sure preserve at least one video frame
        bool bOk = videoQ_.TryPop(pLastPic_);
        if (bOk == true){
                _pFrame = pLastPic_;
        } else {
                if (pLastPic_ == nullptr) {
                        return false;
                }
                _pFrame = pLastPic_;
        }

        ProsetVideo(_pFrame);

        return true;
}

bool Text::GetAudio(OUT std::shared_ptr<MediaFrame>& _pFrame)
{
        // no audio data
        return false;
}

std::shared_ptr<Input> Text::CreateClone(IN const std::string& _clone, IN const Option& _opt)
{
        // create a clone
        auto pText = std::make_shared<Text>(_clone, text_);
        pText->GetOptions(_opt);

        // add clone text in clone list
        if (clones_.Insert(_clone, pText) == true) {
                pText->Start(); // trigger thread loop to generate text picture continuously
                return std::dynamic_pointer_cast<Input>(pText);
        }

        Error("[%s] text clone %s already exists", Name().c_str(), _clone.c_str());
        return nullptr;
}

bool Text::RemoveClone(IN const std::string& _clone)
{
        return clones_.Erase(_clone);
}

std::shared_ptr<Input> Text::GetClone(IN const std::string& _clone)
{
        std::shared_ptr<Input> pClone = nullptr;
        clones_.Find(_clone, [&pClone](std::shared_ptr<Text>& _pText){
                        pClone = std::dynamic_pointer_cast<Input>(_pText);
                });

        return pClone;
}

std::shared_ptr<Input> Text::NextClone()
{
        std::shared_ptr<Input> pNext = nullptr;
        clones_.FindIf([&pNext](const std::string& key, std::shared_ptr<Text>& _pText) -> bool {
                        pNext = std::dynamic_pointer_cast<Input>(_pText);
                        return true;
                });
        return pNext;
}
