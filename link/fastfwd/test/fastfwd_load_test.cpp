#include <iostream>
#include <fstream>
#include <memory>
#include <string>
#include <getopt.h>

#include <grpcpp/grpcpp.h>
#include "fast_forward.grpc.pb.h"
#include "fastfwd.hpp"

using grpc::ClientReader;
using grpc::Channel;
using grpc::ClientContext;
using grpc::Status;
using fastforward::FastForward;
using fastforward::FastForwardInfo;
using fastforward::FastForwardStream;

unsigned int fastfwd::global::nLogLevel = 4;
extern char *optarg;

class FastForwardClient {
public:
        FastForwardClient(std::shared_ptr<Channel> channel) :
                        stub_(FastForward::NewStub(channel)) {
        }

        int ReceiveStream(const std::string& _sUrl, const int _nSpeed, int _nIndex,
                        std::chrono::high_resolution_clock::time_point* _pTResponse) {
                // Data we are sending to the server.
                size_t nReceiveSize = 0;
                FastForwardStream ffs;
                ClientContext context;
                FastForwardInfo request;
                request.set_url(_sUrl);
                request.set_speed(_nSpeed);

                std::unique_ptr<ClientReader<FastForwardStream> > reader(
                                stub_->GetTsStream(&context, request));

                // before read stream, need to wait grpc channel ready
                reader->WaitForInitialMetadata();
                // record time point of grpc server response
                *_pTResponse = std::chrono::high_resolution_clock::now();
                while (reader->Read(&ffs)) {
                        // calculate total size received from grpc server
                        nReceiveSize += ffs.stream().size();
                        Info("client_%05d received %ld KB", _nIndex, nReceiveSize / 1024);
                }
                Status status = reader->Finish();
                if (status.ok()) {
                        return static_cast<int>(nReceiveSize / 1024);
                } else {
                        return -1;
                }
        }

private:
        std::unique_ptr<FastForward::Stub> stub_;
};

int main(int argc, char** argv) {

        std::string szConcurrent = "1";
        std::string szHost = "127.0.0.1:50051";
        std::string szUrl = "test.mp4";
        std::string szSpeed = "x2";

        const char* shortOptions = "c:h:u:x:";
        struct option longOptions[] = {
                {"concurrent", 1, nullptr, 'c'},
                {"host",       1, nullptr, 'h'},
                {"url",        1, nullptr, 'u'},
                {"speed",      1, nullptr, 'x'},
                {nullptr,      0, nullptr,   0}
        };
        int c;
        while ((c = getopt_long(argc, argv, shortOptions, longOptions, nullptr)) != -1) {
                switch (c) {
                case -1:
                case 0: break;
                case 'c': szConcurrent = optarg; break;
                case 'h': szHost = optarg; break;
                case 'u': szUrl = optarg; break;
                case 'x': szSpeed = optarg; break;
                default:
                        std::cout << "usage: " << argv[0] << "\n"
                                  << " --concurrent, -c <concurrent thread> default: 1 \n"
                                  << " --host,       -h <host address> default: 127.0.0.1:50051 \n"
                                  << " --url,        -u <url> default: test.mp4 \n"
                                  << " --speed,      -x <speed> default: x2 \n"
                                  << std::endl;
                        exit(1);
                }
        }

        int nSpeed = fastfwd::x2;
        if (szSpeed.compare("x2") == 0) {
                nSpeed = fastfwd::x2;
        } else if (szSpeed.compare("x4") == 0) {
                nSpeed = fastfwd::x4;
        } else if (szSpeed.compare("x8") == 0) {
                nSpeed = fastfwd::x8;
        } else if (szSpeed.compare("x16") == 0) {
                nSpeed = fastfwd::x16;
        } else if (szSpeed.compare("x32") == 0) {
                nSpeed = fastfwd::x32;
        } else {
                std::cout << "supported speed: x2, x4, x8, x16, x32" << std::endl;
                exit(1);
        }

        int nConcurrent;
        try {
                nConcurrent = std::stoi(szConcurrent);
        } catch(std::invalid_argument& e) {
                std::cout << "Concurrent number not valid" << std::endl;
                exit(1);
        } catch(std::out_of_range& e) {
                std::cout << "Concurrent number out of range" << std::endl;
                exit(1);
        }
        std::vector<std::thread> threads(nConcurrent);

        int nThreadIndex = 0;
        for(std::thread& thread : threads)
        {
                // Capture threadIndex by value, capture other variables by reference
                thread = std::thread([&, nThreadIndex]()
                {
                        using namespace std::chrono;
                        high_resolution_clock::time_point timeResponse;
                        high_resolution_clock::time_point timeStart = high_resolution_clock::now();
                        FastForwardClient ffclient(grpc::CreateChannel(szHost, grpc::InsecureChannelCredentials()));
                        int nReceived = ffclient.ReceiveStream(szUrl, nSpeed, nThreadIndex, &timeResponse);

                        // calculate time duration
                        auto duration = duration_cast<microseconds>(high_resolution_clock::now() - timeStart);
                        double dwDurationTime = double(duration.count()) * microseconds::period::num / microseconds::period::den;
                        duration = duration_cast<microseconds>(timeResponse - timeStart);
                        double dwResponseTime = double(duration.count()) * microseconds::period::num / microseconds::period::den;

                        Info("[* *] Client_%05d received %dKB. grpc server response time: %.3fs, total time: %.3fs.", nThreadIndex, nReceived, dwResponseTime, dwDurationTime);
                        return;
                });

                nThreadIndex++;
        }

        // Wait for all threads to complete
        for(std::thread& thread : threads)
        {
            thread.join();
        }

        Info("All client threads are completed");
        return 0;


}
