#include "stat.hpp"

using namespace muxer;

Statistic::Statistic()
{
        bThreadExit_.store(false);
}

Statistic::~Statistic()
{
        Stop();
}

void Statistic::Start(IN int _nSeconds)
{
        // prevent from starting in parallel
        if (!startLck_.try_lock()) {
                Warn("stat: already started");
                return;
        }

        // spawn thread to calculate the bandwidth
        stat_ = std::thread([this, _nSeconds](){
                        while (bThreadExit_.load() == false) {
                                std::this_thread::sleep_for(std::chrono::milliseconds(_nSeconds * 1000));
                                SetStat();
                        }
                        startLck_.unlock();
                });
}

void Statistic::Stop()
{
        bThreadExit_.store(true);
        if (stat_.joinable()) {
                stat_.join();
        }
}

void Statistic::OnePacket(IN const std::shared_ptr<MediaPacket> _pPacket)
{
        OneSample(_pPacket->Size());
}

void Statistic::OneSample(IN int _nBytes)
{
        if (bThreadExit_.load() == false) {
                std::lock_guard<std::mutex> lock(statLck_);
                nBytes_ += _nBytes;
                nCounts_++;
        }
}

void Statistic::SetStat()
{
        std::lock_guard<std::mutex> lock(statLck_);

        // save recent stat results
        nRecentBytes_ = nBytes_;
        nRecentCounts_ = nCounts_;
        nBytes_ = 0;
        nCounts_ = 0;
}

void Statistic::GetStat(OUT int& _nBytes, OUT int& _nCounts)
{
        std::lock_guard<std::mutex> lock(statLck_);

        _nBytes = nRecentBytes_;
        _nCounts = nRecentCounts_;
}
