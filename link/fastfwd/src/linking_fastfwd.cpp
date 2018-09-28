#include <iostream>
#include <memory>
#include <string>
#include <iostream>
#include <fstream>
#include <getopt.h>

#include <grpcpp/grpcpp.h>

#include "fast_forward.grpc.pb.h"
#include "fastfwd.hpp"

using grpc::Server;
using grpc::ServerBuilder;
using grpc::ServerContext;
using grpc::ServerWriter;
using grpc::Status;
using fastforward::FastForward;
using fastforward::FastForwardInfo;
using fastforward::FastForwardStream;

extern char *optarg;
std::string version = "201809281044";
unsigned int fastfwd::global::nLogLevel = 4;
int fastfwd::nStatPeriod = 10;

// Logic and data behind the server's behavior.
class FastForwardServiceImpl final : public FastForward::Service {
        Status GetTsStream(ServerContext* context, const FastForwardInfo* request,
                        ServerWriter<FastForwardStream>* writer) override {

                FastForwardStream ffs;
                std::string szUrl = request->url();
                int nSpeed = reinterpret_cast<int>(request->speed());
                auto pStat = Singular<fastfwd::Statistic>::Instance();
                auto pPumper = std::make_unique<fastfwd::StreamPumper>(szUrl, nSpeed, 16384);

                std::vector<char> chBuffer;
                while (pPumper->Pump(chBuffer, 16384) == 0) {
                        ffs.set_stream(chBuffer.data(), chBuffer.size());
                        if(!writer->Write(ffs)) {
                                return Status::CANCELLED;
                        }
                        // statistic output bytes
                        pStat->IncOutBytes(chBuffer.size());
                }
                return Status::OK;
        }
};

void RunServer(const std::string port) {
        std::string server_address = "0.0.0.0";
        server_address.append(":").append(port);
        FastForwardServiceImpl service;

        ServerBuilder builder;
        // Listen on the given address without any authentication mechanism.
        builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
        // Register "service" as the instance through which we'll communicate with
        // clients. In this case it corresponds to an *synchronous* service.
        builder.RegisterService(&service);
        // Finally assemble the server.
        std::unique_ptr<Server> server(builder.BuildAndStart());
        Info("Server listening on %s", server_address.c_str());

        // Wait for the server to shutdown. Note that some other thread must be
        // responsible for shutting down the server for this call to ever return.
        server->Wait();
}

int main(int argc, char** argv) {

        std::string sPort = "50051";

        const char* shortOptions = "p:";
        struct option longOptions[] = {
                {"port",       1, nullptr, 'p'},
                {nullptr,      0, nullptr,   0}
        };
        int c;
        while ((c = getopt_long(argc, argv, shortOptions, longOptions, nullptr)) != -1) {
                switch (c) {
                case -1:
                case 0: break;
                case 'p': sPort = optarg; break;
                default:
                        std::cout << "usage: " << argv[0] << "\n"
                                  << " --port, -p <port number> default: 50051 \n"
                                  << std::endl;
                        exit(1);
                }
        }

        // port
        int nPort;
        try {
                nPort = std::stoi(sPort);
        } catch(std::invalid_argument& e) {
                std::cout << "port number not valid" << std::endl;
                exit(1);
        } catch(std::out_of_range& e) {
                std::cout << "port number out of range" << std::endl;
                exit(1);
        }
        Info("linking_fastfwd started, version=%s, arguments: port=%s ", version.c_str(), sPort.c_str());;

        auto pStat = Singular<fastfwd::Statistic>::Instance();
        RunServer(sPort);

        return 0;
}
