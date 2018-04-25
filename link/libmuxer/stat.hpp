#ifndef __STAT_HPP__
#define __STAT_HPP__

#include "common.hpp"

namespace muxer
{
        // forward decls
        class MediaPacket;

        class Statistic
        {
        public:
                Statistic();
                ~Statistic();

                void Start(IN int nSeconds);
                void Stop();

                void OnePacket(IN const std::shared_ptr<MediaPacket> pPacket);
                void OneSample(IN int nBytes);
                void GetStat(OUT int& nBytes, OUT int& nCounts);
        private:
                void SetStat();
        private:
                std::mutex statLck_;
                int nRecentBytes_ = 0, nRecentCounts_ = 0;
                int nBytes_ = 0, nCounts_ = 0;

                std::atomic<bool> bThreadExit_;
                std::mutex startLck_;
                std::thread stat_;
        };
}

#endif
