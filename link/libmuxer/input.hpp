#ifndef __INPUT_HPP__
#define __INPUT_HPP__

#include "common.hpp"

namespace muxer
{
        // AvReceiver
        typedef const std::function<int(const std::unique_ptr<MediaPacket>)> PacketHandlerType;
        class AvReceiver
        {
        public:
                AvReceiver();
                ~AvReceiver();
                int Receive(IN const std::string& url, IN PacketHandlerType& callback);
        private:
                struct StreamInfo;
                static int AvInterruptCallback(void* pContext);
                static bool EmulateFramerate(IN int64_t nPts, OUT StreamInfo& stream);
        private:
                std::chrono::high_resolution_clock::time_point start_;
                long nTimeout_ = 10000; // 10 seconds timeout by default

                struct AVFormatContext* pAvContext_ = nullptr;
                struct StreamInfo {
                        struct AVStream* pAvStream;
                        int64_t nFirstPts = -1;
                        std::chrono::high_resolution_clock::time_point start;
                };
                std::vector<StreamInfo> streams_;
        };

        // AvDecoder
        typedef const std::function<int(const std::shared_ptr<MediaFrame>&)> FrameHandlerType;
        class AvDecoder
        {
        public:
                AvDecoder();
                ~AvDecoder();
                int Decode(IN const std::unique_ptr<MediaPacket>& pPacket, IN FrameHandlerType& callback);
        private:
                int Init(IN const std::unique_ptr<MediaPacket>& pPakcet);
        private:
                AVCodecContext* pAvDecoderContext_ = nullptr;
                bool bIsDecoderAvailable_ = false;
        };

        // Input
        class Input : public OptionMap
        {
        public:
                Input(IN const std::string& name);
                virtual ~Input();
                virtual std::string Name() final;
                virtual std::string String() = 0;

                // pop one video/audio
                virtual bool GetVideo(OUT std::shared_ptr<MediaFrame>& pFrame) = 0;
                virtual bool GetAudio(OUT std::shared_ptr<MediaFrame>& pFrame) = 0;

                virtual std::shared_ptr<Input> CreateClone(IN const std::string& clone, IN const Option& opt) = 0;
                virtual bool RemoveClone(IN const std::string& clone) = 0;
                virtual std::shared_ptr<Input> GetClone(IN const std::string& clone) = 0;
                virtual std::shared_ptr<Input> NextClone() = 0;

                // attached muxer
                virtual void AttachMuxer(IN const std::string& mux) final;
                virtual std::string& AttachedMuxer() final;
                virtual void DetachMuxer() final;

                // frame filters
                virtual void PresetVideo(INOUT std::shared_ptr<MediaFrame>& pFrame) final;
                virtual void ProsetVideo(INOUT std::shared_ptr<MediaFrame>& pFrame) final;
        public:
                Statistic stat;
        protected:
                std::string name_;
                std::string attachedMuxer_ = ""; // attached muxer
                AvFilter avFilter;
        };

        // Stream input
        class Stream : public Input
        {
        public:
                Stream(IN const std::string& name);
                ~Stream();
                void Start(IN const std::string& url);
                void Stop();
                virtual std::string String() override;

                // pop one video/audio
                virtual bool GetVideo(OUT std::shared_ptr<MediaFrame>& pFrame) override;
                virtual bool GetAudio(OUT std::shared_ptr<MediaFrame>& pFrame) override;

                virtual std::shared_ptr<Input> CreateClone(IN const std::string& clone, IN const Option& opt) override;
                virtual bool RemoveClone(IN const std::string& clone) override;
                virtual std::shared_ptr<Input> GetClone(IN const std::string& clone) override;
                virtual std::shared_ptr<Input> NextClone() override;
        private:
                // push one video/audio
                void SetAudio(const std::shared_ptr<MediaFrame>& pFrame);
                void SetVideo(const std::shared_ptr<MediaFrame>& pFrame);
                void SetFrame(const std::shared_ptr<MediaFrame>& pFrame);
        private:
                static const size_t AUDIO_Q_LEN = 15;
                static const size_t VIDEO_Q_LEN = 10;

                std::thread receiver_;
                std::atomic<bool> bReceiverExit_;
                SharedQueue<std::shared_ptr<MediaFrame>> videoQ_;
                SharedQueue<std::shared_ptr<MediaFrame>> audioQ_;

                std::shared_ptr<MediaFrame> pLastVideo_ = nullptr;
                std::mutex lastVideoLck_; // Start() vs CreateClone()

                AudioResampler resampler_;
                std::vector<uint8_t> sampleBuffer_;
                std::mutex sampleBufferLck_; // GetAudio() vs SetAudio()

                std::shared_ptr<VideoRescaler> pRescaler_ = nullptr;

                SharedMap<std::string, std::shared_ptr<Stream>> clones_;
        };

        // Text input
        class Text : public Input
        {
        public:
                Text(IN const std::string& name, IN const std::string& text);
                ~Text();
                virtual std::string String();

                virtual bool GetVideo(OUT std::shared_ptr<MediaFrame>& pFrame);
                virtual bool GetAudio(OUT std::shared_ptr<MediaFrame>& pFrame);

                virtual std::shared_ptr<Input> CreateClone(IN const std::string& clone, IN const Option& opt);
                virtual bool RemoveClone(IN const std::string& clone);
                virtual std::shared_ptr<Input> GetClone(IN const std::string& clone);
                virtual std::shared_ptr<Input> NextClone();
        private:
                void Start();
                void Stop();
        private:
                static const size_t VIDEO_Q_LEN = 10;
                const std::string text_;

                std::thread picMaker_;
                std::atomic<bool> bMakerExit_;
                SharedQueue<std::shared_ptr<MediaFrame>> videoQ_;
                std::shared_ptr<MediaFrame> pLastPic_ = nullptr;

                SharedMap<std::string, std::shared_ptr<Text>> clones_;
        };
}

#endif
