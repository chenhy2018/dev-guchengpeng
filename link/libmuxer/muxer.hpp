#ifndef __MUXER_HPP__
#define __MUXER_HPP__

#include "common.hpp"
#include "input.hpp"
#include "output.hpp"

namespace muxer
{
        class VideoMuxer
        {
        public:
                VideoMuxer(IN int nW, IN int nH);
                ~VideoMuxer();
                int Mux(IN std::vector<std::shared_ptr<MediaFrame>>& frames, OUT std::shared_ptr<MediaFrame>&);
                void BgColor(int nRGB);
        private:
                void Overlay(IN const std::shared_ptr<MediaFrame>& pFrom, OUT std::shared_ptr<MediaFrame>& pTo);
                int nCanvasW_ = 0, nCanvasH_ = 0;
                int nBackground_ = 0x000000; // black
        };

        class AudioMixer
        {
        public:
                AudioMixer();
                ~AudioMixer();
                int Mix(IN const std::vector<std::shared_ptr<MediaFrame>>& frames, OUT std::shared_ptr<MediaFrame>&);
        private:
                void SimpleMix(IN const std::shared_ptr<MediaFrame>& pFrom, OUT std::shared_ptr<MediaFrame>& pTo);
        };

        class Muxer : public OptionMap
        {
        public:
                Muxer(IN const std::string& name, IN int nWidth, IN int nHeight);
                ~Muxer();
                int Start();
                void Stop();
                std::string Name() const;
                std::string String();
                void GetOutputBps(OUT int& nOutBps);

                int AddRtmpOutput(IN const std::string& output, IN const std::string& url); // TODO delete
                int AddRtmpOutput(IN const std::string& output, IN const std::string& url, IN const Option& opt);
                int AddFileOutput(IN const std::string& output, IN const std::string& path, IN const Option& opt);
                int ModOutputOption(IN const std::string& output, IN const std::string& key, IN const std::string& val = "");
                int ModOutputOption(IN const std::string& output, IN const std::string& key, IN int nVal);
                int DelOutputOption(IN const std::string& output, IN const std::string& key);
                int RemoveOutput(IN const std::string& output);
                int AttachInputClone(IN const std::string& input, IN const std::string& clone,
                                     IN const std::shared_ptr<Input> pClone);
                int DetachInputClone(IN const std::string& input, IN const std::string& clone);
        private:
                std::shared_ptr<Output> FindOutput(IN const std::string& name);
        private:
                const std::string DELIMITER = "@#$%^&*";

                std::string name_;
                std::atomic<bool> bMuxerExit_;

                SharedQueue<std::shared_ptr<Output>> outputs_;
                SharedMap<std::string, std::shared_ptr<Input>> inputClones_;

                std::thread videoMuxerThread_;
                std::thread audioMuxerThread_;

                VideoMuxer videoMuxer_;
                AudioMixer audioMixer_;

                // internal clock to generate pts
                std::mutex clockLck_;
                uint64_t nInitClock_ = 0;
                double dAudioAccPts_ = 0.0;
                uint64_t nVideoAccPts_ = 0;
        };

        class AvMuxer
        {
        public:
                AvMuxer();
                ~AvMuxer();
                void Print();

                // gereral inputs
                int AddInput(IN const std::string& input, IN const std::string& url);
                int AddStream(IN const std::string& input, IN const std::string& url);
                int AddText(IN const std::string& input, IN const std::string& text);
                int RemoveInput(IN const std::string& input);

                // input clones
                int AddInputClone(IN const std::string& input, IN const std::string& clone);
                int AddInputClone(IN const std::string& input, IN const std::string& clone, IN const Option& opt);
                int Attach(IN const std::string& mux, IN const std::string& input, IN const std::string& clone);
                int Detach(IN const std::string& mux, IN const std::string& input, IN const std::string& clone);
                int ModInputCloneOption(IN const std::string& input, IN const std::string& clone,
                                        IN const std::string& key, IN const std::string& val = "");
                int ModInputCloneOption(IN const std::string& input, IN const std::string& clone,
                                        IN const std::string& key, IN int nVal);
                int DelInputCloneOption(IN const std::string& input, IN const std::string& clone,
                                        IN const std::string& key);
                int RemoveInputClone(IN const std::string& input, IN const std::string& clone);

                // general outputs
                int AddOutput(IN const std::string& mux, IN const std::string& output, IN const std::string& url);
                int AddOutput(IN const std::string& mux, IN const std::string& output, IN const std::string& url, IN const Option& opt);
                int AddFileOutput(IN const std::string& mux, IN const std::string& output, IN const std::string& path);
                int AddFileOutput(IN const std::string& mux, IN const std::string& output, IN const std::string& path, IN const Option& opt);
                //int AddRtmpOutput(IN const std::string& mux, IN const std::string& output, IN const std::string& url);
                //int AddRtmpOutput(IN const std::string& mux, IN const std::string& output, IN const std::string& url, IN const Option& opt);
                int ModOutputOption(IN const std::string& mux, IN const std::string& output,
                                    IN const std::string& key, IN const std::string& val = "");
                int ModOutputOption(IN const std::string& mux, IN const std::string& output, IN const std::string& key, IN int nVal);
                int DelOutputOption(IN const std::string& mux, IN const std::string& output, IN const std::string& key);
                int RemoveOutput(IN const std::string& mux, IN const std::string& output);

                // general muxers
                int AddMux(IN const std::string& mux, IN int nWidth, IN int nHeight);
                int RemoveMux(IN const std::string& mux);
                int ModMuxOption(IN const std::string& mux, IN const std::string& key, IN const std::string& val = "");
                int ModMuxOption(IN const std::string& mux, IN const std::string& key, IN int nVal);
                int DelMuxOption(IN const std::string& mux, IN const std::string& key);

                // get statistic information
                void GetBandwidth(OUT int& nInBps, OUT int& nOutBps);
        private:
                std::shared_ptr<Input> FindInput(IN const std::string& input);
                std::shared_ptr<Input> FindInputClone(IN const std::string& input, IN const std::string& clone);
                // without api locks
                int RemoveInputClone_(IN const std::string& input, IN const std::string& clone);
                int Detach_(IN const std::string& mux, IN const std::string& input, IN const std::string& clone);
        private:
                SharedQueue<std::shared_ptr<Input>> inputs_;
                SharedMap<std::string, std::shared_ptr<Muxer>> muxers_;
                std::mutex apiLck_;
        };
}

#endif
