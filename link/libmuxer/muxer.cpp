#include "muxer.hpp"

using namespace muxer;

//
// OptionMap
//

bool OptionMap::GetOption(IN const std::string& _key, OUT std::string& _value)
{
        std::lock_guard<std::mutex> lock(paramsLck_);

        auto it = params_.find(_key);
        if (it != params_.end()) {
                _value = it->second;
                return true;
        }
        return false;
}

bool OptionMap::GetOption(IN const std::string& _key, OUT int& _value)
{
        std::lock_guard<std::mutex> lock(intparamsLck_);

        auto it = intparams_.find(_key);
        if (it != intparams_.end()) {
                _value = it->second;
                return true;
        }
        return false;
}

bool OptionMap::GetOption(IN const std::string& _key)
{
        std::lock_guard<std::mutex> lock(paramsLck_);

        if (params_.find(_key) != params_.end()) {
                return true;
        }
        return false;
}

void OptionMap::SetOption(IN const std::string& _key, IN const std::string& _val)
{
        std::lock_guard<std::mutex> lock(paramsLck_);

        params_[_key] = _val;
}

void OptionMap::SetOption(IN const std::string& _key, IN int _val)
{
        std::lock_guard<std::mutex> lock(paramsLck_);

        intparams_[_key] = _val;
}

void OptionMap::SetOption(IN const std::string& _flag)
{
        std::lock_guard<std::mutex> lock(paramsLck_);

        params_[_flag] = "";
}

void OptionMap::DelOption(IN const std::string& _key)
{
        std::lock_guard<std::mutex> lock(paramsLck_);

        params_.erase(_key);
        intparams_.erase(_key);
}

void OptionMap::GetOptions(IN const OptionMap& _opts)
{
        params_ = _opts.params_;
        intparams_ = _opts.intparams_;
}

//
// Muxer
//

Muxer::Muxer(IN const std::string& _name, IN int _nWidth, IN int _nHeight)
        :name_(_name),
         videoMuxer_(_nWidth, _nHeight)
{
}

Muxer::~Muxer()
{
        Stop();
}

std::string Muxer::Name() const
{
        return name_;
}

std::string Muxer::String()
{
        std::string out = "[" + Name() + "]\n";
        outputs_.Foreach([&](std::shared_ptr<Output> _pOutput){
                        out += "  output=" + _pOutput->Name() + "\n";
                });

        inputClones_.Foreach([&](std::string _key, std::shared_ptr<Input> _pClone){
                        out += "  input=" + _pClone->Name() + " (" + _key + ")\n";
                });

        return out;
}

void Muxer::GetOutputBps(OUT int& _nOutBps)
{
        _nOutBps = 0;

        outputs_.Foreach([&_nOutBps](std::shared_ptr<Output> _pOutput){
                        int nBytes, nFrames;
                        _pOutput->stat.GetStat(nBytes, nFrames);
                        _nOutBps += nBytes * 8;
                });
}

int Muxer::Start()
{
        auto muxVideo = [this](){
                std::vector<std::shared_ptr<MediaFrame>> videoFrames;
                std::shared_ptr<MediaFrame> pOutFrame;

                using namespace std::chrono;
                auto start = high_resolution_clock::now();
                uint64_t nIndex = 0;
                int nDuration = 40; // same as encoder time_base (1/25)

                while (bMuxerExit_.load() == false) {
                        inputClones_.CriticalSection(
                                [&videoFrames](std::unordered_map<std::string, std::shared_ptr<Input>>& _map) {
                                        for (auto it = _map.begin(); it != _map.end(); it++) {
                                                // get frames from input clones
                                                auto pClone = it->second;
                                                std::shared_ptr<MediaFrame> pFrame;
                                                if (pClone->GetVideo(pFrame) == true) {
                                                        if (pClone->GetOption(options::hidden) == false) {
                                                                videoFrames.push_back(pFrame);
                                                        }
                                                }
                                        }
                                });

                        // background color
                        int nRGB;
                        if (GetOption(options::bgcolor, nRGB) == true) {
                                videoMuxer_.BgColor(nRGB);
                        }

                        // mux pictures
                        if (videoMuxer_.Mux(videoFrames, pOutFrame) == 0) {
                                pOutFrame->AvFrame()->pts = nVideoAccPts_;
                                nVideoAccPts_ += nDuration;
                                outputs_.Foreach([&](std::shared_ptr<Output>& _pOutput) {
                                                _pOutput->Push(std::make_shared<MediaFrame>(*pOutFrame));
                                        });
                        }
                        videoFrames.clear();

                        // sync by system clock
                        nIndex++;
                        high_resolution_clock::time_point now = high_resolution_clock::now();
                        auto pastTime = duration_cast<microseconds>(now - start).count();
                        int delay = nIndex * nDuration * 1000 - pastTime; // microseconds
                        if (delay > 0) {
                                usleep(delay);
                        }
                }
        };

        auto muxAudio = [this](){
                std::vector<std::shared_ptr<MediaFrame>> audioFrames;
                std::shared_ptr<MediaFrame> pOutFrame;

                using namespace std::chrono;
                auto start = high_resolution_clock::now();
                uint64_t nIndex = 0;
                double dDuration = static_cast<double>(AudioResampler::FRAME_SIZE) * 1000 / AudioResampler::SAMPLE_RATE;

                while (bMuxerExit_.load() == false) {
                        inputClones_.CriticalSection(
                                [&audioFrames](std::unordered_map<std::string, std::shared_ptr<Input>>& _map) {
                                        for (auto it = _map.begin(); it != _map.end(); it++) {
                                                // get frames from input clones
                                                auto pClone = it->second;
                                                std::shared_ptr<MediaFrame> pFrame;
                                                if (pClone->GetAudio(pFrame) == true) {
                                                        if (pClone->GetOption(options::muted) == false) {
                                                                audioFrames.push_back(pFrame);
                                                        }
                                                }
                                        }
                                });

                        if (audioMixer_.Mix(audioFrames, pOutFrame) == 0) {
                                pOutFrame->AvFrame()->pts = static_cast<uint64_t>(dAudioAccPts_);
                                dAudioAccPts_ += dDuration;
                                outputs_.Foreach([&](std::shared_ptr<Output>& _pOutput) {
                                                _pOutput->Push(std::make_shared<MediaFrame>(*pOutFrame));
                                        });
                        }
                        audioFrames.clear();

                        // sync by system clock
                        nIndex++;
                        high_resolution_clock::time_point now = high_resolution_clock::now();
                        auto pastTime = duration_cast<microseconds>(now - start).count();
                        int delay = dDuration * nIndex * 1000 - pastTime; // microseconds
                        if (delay > 0) {
                                usleep(delay);
                        }
                }
        };

        videoMuxerThread_ = std::thread(muxVideo);
        audioMuxerThread_ = std::thread(muxAudio);

        return 0;
}

void Muxer::Stop()
{
        bMuxerExit_.store(true);
        if (videoMuxerThread_.joinable()) {
                videoMuxerThread_.join();
        }
        if (audioMuxerThread_.joinable()) {
                audioMuxerThread_.join();
        }
}

int Muxer::AddRtmpOutput(IN const std::string& _output, IN const std::string& _url)
{
        auto pFound = FindOutput(_output);
        if (pFound != nullptr) {
                Error("[%s] output %s already exisits", Name().c_str(), _output.c_str());
                return -1;
        }

        auto pRtmp = std::make_shared<Rtmp>(_output);
        pRtmp->Start(_url);
        std::shared_ptr<Output> r = std::dynamic_pointer_cast<Output>(pRtmp);
        outputs_.Push(std::move(r));

        return 0;
}

int Muxer::AddRtmpOutput(IN const std::string& _output, IN const std::string& _url, IN const Option& _opt)
{
        auto pFound = FindOutput(_output);
        if (pFound != nullptr) {
                Error("[%s] output %s already exisits", Name().c_str(), _output.c_str());
                return -1;
        }

        auto pRtmp = std::make_shared<Rtmp>(_output);
        pRtmp->GetOptions(_opt);
        pRtmp->Start(_url);
        std::shared_ptr<Output> r = std::dynamic_pointer_cast<Output>(pRtmp);
        outputs_.Push(std::move(r));

        return 0;
}

int Muxer::AddFileOutput(IN const std::string& _output, IN const std::string& _path, IN const Option& _opt)
{
        auto pFound = FindOutput(_output);
        if (pFound != nullptr) {
                Error("[%s] output %s already exisits", Name().c_str(), _output.c_str());
                return -1;
        }

        auto pFile = std::make_shared<File>(_output);
        pFile->GetOptions(_opt);
        pFile->Start(_path);
        std::shared_ptr<Output> r = std::dynamic_pointer_cast<Output>(pFile);
        outputs_.Push(std::move(r));

        return 0;
}

int Muxer::ModOutputOption(IN const std::string& _output, IN const std::string& _key, IN const std::string& _val)
{
        auto r = FindOutput(_output);
        if (r == nullptr) {
                return -1;
        }

        r->SetOption(_key, _val);

        return 0;
}

int Muxer::ModOutputOption(IN const std::string& _output, IN const std::string& _key, IN int _nVal)
{
        auto r = FindOutput(_output);
        if (r == nullptr) {
                return -1;
        }

        r->SetOption(_key, _nVal);

        return 0;
}

int Muxer::DelOutputOption(IN const std::string& _output, IN const std::string& _key)
{
        auto r = FindOutput(_output);
        if (r == nullptr) {
                return -1;
        }

        r->DelOption(_key);

        return 0;
}

int Muxer::RemoveOutput(IN const std::string& _output)
{
        outputs_.CriticalSection([_output](std::deque<std::shared_ptr<Output>>& _queue){
                        for (auto it = _queue.begin(); it != _queue.end(); it++) {
                                if ((*it)->Name().compare(_output) == 0) {
                                        it = _queue.erase(it);
                                        return;
                                }
                        }
                });

        return 0;
}

int Muxer::AttachInputClone(IN const std::string& _input, IN const std::string& _clone,
                            IN const std::shared_ptr<Input> _pClone)
{
        auto key = _input + Muxer::DELIMITER + _clone;
        if (inputClones_.Insert(key, _pClone) != true) {
                Error("[%s] input clone already exists: input: %s, clone: %s",
                      Name().c_str(), _input.c_str(), _clone.c_str());
                return -1;
        }

        return 0;
}

int Muxer::DetachInputClone(IN const std::string& _input, IN const std::string& _clone)
{
        auto key = _input + Muxer::DELIMITER + _clone;
        if (inputClones_.Erase(key) == false) {
                return -1;
        }

        return 0;
}

std::shared_ptr<Output> Muxer::FindOutput(IN const std::string& _name)
{
        std::shared_ptr<Output> p = nullptr;
        outputs_.FindIf([&](std::shared_ptr<Output>& _pOutput) -> bool {
                        if (_pOutput->Name().compare(_name) == 0) {
                                p = _pOutput;
                                return true;
                        }
                        return false;
                });

        return p;
}

//
// AvMuxer
//

AvMuxer::AvMuxer()
{
        av_register_all();
        avformat_network_init();
        avfilter_register_all();
}

AvMuxer::~AvMuxer()
{
        muxers_.Foreach([](const std::string _key, std::shared_ptr<Muxer> _pMuxer){
                        _pMuxer->Stop();
                });
}

int AvMuxer::AddInput(IN const std::string& _input, IN const std::string& _url)
{
        return AddStream(_input, _url);
}

int AvMuxer::AddStream(IN const std::string& _input, IN const std::string& _url)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        // test existence
        auto pInput = FindInput(_input);
        if (pInput != nullptr) {
                Error("add stream: already exists: %s", _input.c_str());
                return -1;
        }

        auto pStream = std::make_shared<Stream>(_input);
        pStream->Start(_url);
        std::shared_ptr<Input> r = std::dynamic_pointer_cast<Input>(pStream);
        inputs_.Push(std::move(r));

        return 0;
}

int AvMuxer::AddText(IN const std::string& _input, IN const std::string& _text)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        // test existence
        auto pInput = FindInput(_input);
        if (pInput != nullptr) {
                Error("add text: already exists: %s", _input.c_str());
                return -1;
        }

        // text input will spawn thread in each clone
        auto pText = std::make_shared<Text>(_input, _text);
        std::shared_ptr<Input> r = std::dynamic_pointer_cast<Input>(pText);
        inputs_.Push(std::move(r));

        return 0;
}

int AvMuxer::RemoveInput(IN const std::string& _input)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        auto pInput = FindInput(_input);
        if (pInput == nullptr) {
                Error("remove input: no such input: %s", _input.c_str());
                return -1;
        }

        // detach and remove all input clones first
        for(;;) {
                auto pClone = pInput->NextClone();
                if (pClone == nullptr) {
                        break;
                }
                RemoveInputClone_(_input, pClone->Name());
        }

        // remove input from input list
        inputs_.CriticalSection([_input](std::deque<std::shared_ptr<Input>>& _queue){
                        for (auto it = _queue.begin(); it != _queue.end(); it++) {
                                if ((*it)->Name().compare(_input) == 0) {
                                        it = _queue.erase(it);
                                        return;
                                }
                        }
                });

        return 0;
}

int AvMuxer::AddInputClone(IN const std::string& _input, IN const std::string& _clone)
{
        return AddInputClone(_input, _clone, Option());
}

int AvMuxer::AddInputClone(IN const std::string& _input, IN const std::string& _clone, IN const Option& _opt)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        auto pInput = FindInput(_input);
        if (pInput == nullptr) {
                Error("no such input: %s", _input.c_str());
                return -1;
        }

        auto pClone = pInput->CreateClone(_clone, _opt);
        if (pClone == nullptr) {
                return -1;
        }

        return 0;
}

int AvMuxer::ModInputCloneOption(IN const std::string& _input, IN const std::string& _clone,
                                 IN const std::string& _key, IN const std::string& _val)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        auto pClone = FindInputClone(_input, _clone);
        if (pClone == nullptr) {
                return -1;
        }
        pClone->SetOption(_key, _val);

        return 0;
}

int AvMuxer::ModInputCloneOption(IN const std::string& _input, IN const std::string& _clone,
                                 IN const std::string& _key, IN int _nVal)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        auto pClone = FindInputClone(_input, _clone);
        if (pClone == nullptr) {
                return -1;
        }
        pClone->SetOption(_key, _nVal);

        return 0;
}

int AvMuxer::DelInputCloneOption(IN const std::string& _input, IN const std::string& _clone, IN const std::string& _key)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        auto pClone = FindInputClone(_input, _clone);
        if (pClone == nullptr) {
                return -1;
        }
        pClone->DelOption(_key);

        return 0;
}

int AvMuxer::RemoveInputClone(IN const std::string& _input, IN const std::string& _clone)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        return RemoveInputClone_(_input, _clone);
}

int AvMuxer::RemoveInputClone_(IN const std::string& _input, IN const std::string& _clone)
{
        auto pInput = FindInput(_input);
        if (pInput == nullptr) {
                Error("remove input clone: could not find input: %s", _input.c_str());
                return -1;
        }

        auto pClone = pInput->GetClone(_clone);
        if (pClone == nullptr) {
                Error("remove input clone: could not find input: %s, clone: %s", _input.c_str(), _clone.c_str());
                return -1;
        }

        // make sure detach from muxer before removing the input clone
        Detach_(pClone->AttachedMuxer(), _input, _clone);
        pInput->RemoveClone(_clone);

        return 0;
}

int AvMuxer::Attach(IN const std::string& _mux, IN const std::string& _input, IN const std::string& _clone)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        auto pClone = FindInputClone(_input, _clone);
        if (pClone == nullptr) {
                Error("could not add input clone, no such clone: %s", _clone.c_str());
                return -1;
        }
        pClone->AttachMuxer(_mux);

        int nRet = -1;
        bool bFound = muxers_.Find(_mux, [&nRet, &pClone, _input, _clone](std::shared_ptr<Muxer>& _pMuxer) {
                        nRet = _pMuxer->AttachInputClone(_input, _clone, pClone);
                });
        if (!bFound) {
                Error("attach: could not find sub muxer: %s", _mux.c_str());
        }

        return nRet;
}

int AvMuxer::Detach(IN const std::string& _mux, IN const std::string& _input, IN const std::string& _clone)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        return Detach_(_mux, _input, _clone);
}

int AvMuxer::Detach_(IN const std::string& _mux, IN const std::string& _input, IN const std::string& _clone)
{
        int nRet = -1;
        bool bFound = muxers_.Find(_mux, [&nRet, _input, _clone](std::shared_ptr<Muxer>& _pMuxer) {
                        nRet = _pMuxer->DetachInputClone(_input, _clone);
                });
        if (!bFound) {
                Error("detach: could not find sub muxer: %s", _mux.c_str());
        }

        auto pClone = FindInputClone(_input, _clone);
        if (pClone == nullptr) {
                Error("could not add input clone, no such clone: %s", _clone.c_str());
                return -1;
        }
        pClone->DetachMuxer();

        return nRet;
}

int AvMuxer::AddOutput(IN const std::string& _mux, IN const std::string& _output, IN const std::string& _url)
{
        return AddOutput(_mux, _output, _url, Option());
}

int AvMuxer::AddOutput(IN const std::string& _mux, IN const std::string& _output,
                       IN const std::string& _url, IN const Option& _opt)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        int nRet = -1;
        bool bFound = muxers_.Find(_mux, [&nRet, _output, _url, &_opt](std::shared_ptr<Muxer>& _pMuxer) {
                        nRet = _pMuxer->AddRtmpOutput(_output, _url, _opt);
                });
        if (!bFound) {
                Error("add output: could not find sub muxer: %s", _mux.c_str());
        }
        return nRet;
}

int AvMuxer::AddFileOutput(IN const std::string& _mux, IN const std::string& _output, IN const std::string& _path)
{
        return AddFileOutput(_mux, _output, _path, Option());
}

int AvMuxer::AddFileOutput(IN const std::string& _mux, IN const std::string& _output, IN const std::string& _path, IN const Option& _opt)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        int nRet = -1;
        bool bFound = muxers_.Find(_mux, [&nRet, _output, _path, &_opt](std::shared_ptr<Muxer>& _pMuxer) {
                        nRet = _pMuxer->AddFileOutput(_output, _path, _opt);
                });
        if (!bFound) {
                Error("add file output: could not find sub muxer: %s", _mux.c_str());
        }
        return nRet;
}

int AvMuxer::ModOutputOption(IN const std::string& _mux, IN const std::string& _output,
                             IN const std::string& _key, IN const std::string& _val)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        int nRet = -1;
        bool bFound = muxers_.Find(_mux, [&nRet, _output, _key, _val](std::shared_ptr<Muxer>& _pMuxer) {
                        nRet = _pMuxer->ModOutputOption(_output, _key, _val);
                });
        if (!bFound) {
                Error("modify output: could not find sub muxer: %s", _mux.c_str());
        }

        return nRet;
}

int AvMuxer::ModOutputOption(IN const std::string& _mux, IN const std::string& _output,
                             IN const std::string& _key, IN int _nVal)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        int nRet = -1;
        bool bFound = muxers_.Find(_mux, [&nRet, _output, _key, _nVal](std::shared_ptr<Muxer>& _pMuxer) {
                        nRet = _pMuxer->ModOutputOption(_output, _key, _nVal);
                });
        if (!bFound) {
                Error("modify output: could not find sub muxer: %s", _mux.c_str());
        }

        return nRet;
}

int AvMuxer::DelOutputOption(IN const std::string& _mux, IN const std::string& _output, IN const std::string& _key)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        int nRet = -1;
        bool bFound = muxers_.Find(_mux, [&nRet, _output, _key](std::shared_ptr<Muxer>& _pMuxer) {
                        nRet = _pMuxer->DelOutputOption(_output, _key);
                });
        if (!bFound) {
                Error("delete output option: could not find sub muxer: %s", _mux.c_str());
        }

        return nRet;
}

int AvMuxer::RemoveOutput(IN const std::string& _mux, IN const std::string& _output)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        int nRet = -1;
        bool bFound = muxers_.Find(_mux, [&nRet, _output](std::shared_ptr<Muxer>& _pMuxer) {
                        nRet = _pMuxer->RemoveOutput(_output);
                });
        if (!bFound) {
                Error("remove output: could not find sub muxer: %s", _mux.c_str());
        }
        return nRet;
}

int AvMuxer::AddMux(IN const std::string& _mux, IN int _nWidth, IN int _nHeight)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        auto pMuxer = std::make_shared<Muxer>(_mux, _nWidth, _nHeight);
        pMuxer->Start();
        if (muxers_.Insert(_mux, pMuxer) == false) {
                Error("could not use this muxer name: %s", _mux.c_str());
                return -1;
        }

        return 0;
}

int AvMuxer::RemoveMux(IN const std::string& _mux)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        return muxers_.Erase(_mux);
}

int AvMuxer::ModMuxOption(IN const std::string& _mux, IN const std::string& _key, IN const std::string& _val)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        bool bFound = muxers_.Find(_mux, [_key, _val](std::shared_ptr<Muxer>& _pMuxer) {
                        _pMuxer->SetOption(_key, _val);
                });
        if (!bFound) {
                Error("mod option, muxer not found: %s", _mux.c_str());
                return -1;
        }

        return 0;
}

int AvMuxer::ModMuxOption(IN const std::string& _mux, IN const std::string& _key, IN int _nVal)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        bool bFound = muxers_.Find(_mux, [_key, _nVal](std::shared_ptr<Muxer>& _pMuxer) {
                        _pMuxer->SetOption(_key, _nVal);
                });
        if (!bFound) {
                Error("mod option, muxer not found: %s", _mux.c_str());
                return -1;
        }

        return 0;
}

int AvMuxer::DelMuxOption(IN const std::string& _mux, IN const std::string& _key)
{
        std::lock_guard<std::mutex> lock(apiLck_);

        bool bFound = muxers_.Find(_mux, [_key](std::shared_ptr<Muxer>& _pMuxer) {
                        _pMuxer->DelOption(_key);
                });
        if (!bFound) {
                Error("mod option, muxer not found: %s", _mux.c_str());
                return -1;
        }

        return 0;
}

std::shared_ptr<Input> AvMuxer::FindInputClone(IN const std::string& _input, IN const std::string& _clone)
{
        auto pInput = FindInput(_input);
        if (pInput == nullptr) {
                Error("could not find input: %s", _input.c_str());
                return std::shared_ptr<Input>();
        }

        return pInput->GetClone(_clone);
}

std::shared_ptr<Input> AvMuxer::FindInput(IN const std::string& _input)
{
        std::shared_ptr<Input> p = nullptr;
        inputs_.FindIf([&](std::shared_ptr<Input>& _pInput) -> bool {
                        if (_pInput->Name().compare(_input) == 0) {
                                p = _pInput;
                                return true;
                        }
                        return false;
                });

        return p;
}

void AvMuxer::Print()
{
        std::lock_guard<std::mutex> lock(apiLck_);

        std::string out;
        out += "print detailed inforamtion of muxers and inputs:\n";
        muxers_.Foreach([&](std::string _mux, std::shared_ptr<Muxer> _pMuxer){
                        out += _pMuxer->String();
                });
        out += "\n";
        inputs_.Foreach([&](std::shared_ptr<Input> _pInput) {
                        out += _pInput->String();
                });
        Info("%s", out.c_str());
}

void AvMuxer::GetBandwidth(OUT int& _nInBps, OUT int& _nOutBps)
{
        _nInBps = 0;
        _nOutBps = 0;

        std::lock_guard<std::mutex> lock(apiLck_);

        // calculate input bandwidth
        inputs_.Foreach([&_nInBps](std::shared_ptr<Input> _pInput) {
                        int nBytes, nFrames;
                        _pInput->stat.GetStat(nBytes, nFrames);
                        _nInBps += nBytes * 8;
                });

        muxers_.Foreach([&_nOutBps](std::string _mux, std::shared_ptr<Muxer> _pMuxer) {
                        int nBps;
                        _pMuxer->GetOutputBps(nBps);
                        _nOutBps += nBps;
                });
}

//
// VideoMuxer
//

VideoMuxer::VideoMuxer(IN int _nW, IN int _nH)
{
        nCanvasW_ = _nW;
        nCanvasH_ = _nH;
}

VideoMuxer::~VideoMuxer()
{
}

void VideoMuxer::BgColor(int _nRGB)
{
        nBackground_ = _nRGB;
}

int VideoMuxer::Mux(IN std::vector<std::shared_ptr<MediaFrame>>& _frames, OUT std::shared_ptr<MediaFrame>& _pOut)
{
        // by default the canvas is pure black, or customized background color
        auto pMuxed = std::make_shared<MediaFrame>(nCanvasW_, nCanvasH_, VideoRescaler::PIXEL_FMT, nBackground_);
        pMuxed->Stream(STREAM_VIDEO);
        pMuxed->Codec(CODEC_H264);

        // sort by Z coordinate
        std::sort(_frames.begin(), _frames.end(),
                  [](const std::shared_ptr<MediaFrame>& i, const std::shared_ptr<MediaFrame>& j) {
                          return i->Z() < j->Z();
                });

        // mux pictures
        for (auto& pFrame : _frames) {
                if (pFrame == nullptr) {
                        Warn("internal: got 1 null frame, something was wrong");
                        continue;
                }
                Overlay(pFrame, pMuxed);
        }

        _pOut = pMuxed;

        return 0;
}

void VideoMuxer::Overlay(IN const std::shared_ptr<MediaFrame>& _pFrom, OUT std::shared_ptr<MediaFrame>& _pTo)
{
        merge::Overlay(_pFrom, _pTo);
}

//
// AudioMixer
//

AudioMixer::AudioMixer()
{
}

AudioMixer::~AudioMixer()
{
}

int AudioMixer::Mix(IN const std::vector<std::shared_ptr<MediaFrame>>& _frames, OUT std::shared_ptr<MediaFrame>& _pOut)
{
        // get a silent audio frame
        auto pMuted = std::make_shared<MediaFrame>(AudioResampler::FRAME_SIZE, AudioResampler::CHANNELS,
                                                   AudioResampler::SAMPLE_FMT, true);
        pMuted->Stream(STREAM_AUDIO);
        pMuted->Codec(CODEC_AAC);
        pMuted->AvFrame()->sample_rate = AudioResampler::SAMPLE_RATE;

        // mixer works here
        for (auto& pFrame : _frames) {
                if (pFrame == nullptr) {
                        Warn("internal: got 1 null frame, something was wrong");
                        continue;
                }
                SimpleMix(pFrame, pMuted);
        }

        _pOut = pMuted;

        return 0;
}

void AudioMixer::SimpleMix(IN const std::shared_ptr<MediaFrame>& _pFrom, OUT std::shared_ptr<MediaFrame>& _pTo)
{
        AVFrame* pF = _pFrom->AvFrame();
        AVFrame* pT = _pTo->AvFrame();

        int16_t* pF16 = (int16_t*)pF->data[0];
        int16_t* pT16 = (int16_t*)pT->data[0];

        for (int i = 0; i < pF->linesize[0] && i < pT->linesize[0]; i += 2) {
                int32_t nMixed = *pF16 + *pT16;
                if (nMixed > 32767) {
                        nMixed = 32767;
                } else if (nMixed < -32768) {
                        nMixed = -32768;
                }
                *pT16 = static_cast<int16_t>(nMixed);

                pF16++;
                pT16++;
        }
}
