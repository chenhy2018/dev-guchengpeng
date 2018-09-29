#include <iostream>
#include <fstream>
#include <memory>
#include <string>

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

class FastForwardClient {
public:
        FastForwardClient(std::shared_ptr<Channel> channel) :
                        stub_(FastForward::NewStub(channel)) {
        }

        int ReceiveStream(const std::string& _url, const int _speed) {
                // Data we are sending to the server.
                FastForwardStream ffs;
                ClientContext context;
                FastForwardInfo request;
                request.set_url(_url);
                request.set_speed(_speed);

                std::unique_ptr<ClientReader<FastForwardStream> > reader(
                                stub_->GetTsStream(&context, request));

                std::ofstream outfile("grpc_test.mp4",std::ofstream::binary);

                // before read stream, need to wait grpc channel ready
                reader->WaitForInitialMetadata();

                while (reader->Read(&ffs)) {

                        Info("%ld", ffs.stream().size());
                        outfile.write((ffs.stream().data()), ffs.stream().size());
                }
                Status status = reader->Finish();
                if (status.ok()) {
                  std::cout << "fast forward grpc succeeded." << std::endl;
                  return 0;
                } else {
                  std::cout << "fast forward grpc failed." << std::endl;
                  return -1;
                }
        }

private:
        std::unique_ptr<FastForward::Stub> stub_;
};

int main(int argc, char** argv) {
        if (argc != 4) {
                std::cout << argv[0] << "<server-address:port> <url> <speed>" << std::endl;
                exit(1);
        }

        std::string szServerAddress = argv[1];
        std::string szUrl = argv[2];
        std::string szSpeed = argv[3];
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

        // (use of InsecureChannelCredentials()).
        FastForwardClient ffclient(grpc::CreateChannel(szServerAddress, grpc::InsecureChannelCredentials()));
        int reply = ffclient.ReceiveStream(szUrl, nSpeed);
        std::cout << "GRPC returned: " << reply << std::endl;

        return reply;
}
