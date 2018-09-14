// Generated by the gRPC C++ plugin.
// If you make any local change, they will be lost.
// source: fast_forward.proto
#ifndef GRPC_fast_5fforward_2eproto__INCLUDED
#define GRPC_fast_5fforward_2eproto__INCLUDED

#include "fast_forward.pb.h"

#include <grpcpp/impl/codegen/async_generic_service.h>
#include <grpcpp/impl/codegen/async_stream.h>
#include <grpcpp/impl/codegen/async_unary_call.h>
#include <grpcpp/impl/codegen/method_handler_impl.h>
#include <grpcpp/impl/codegen/proto_utils.h>
#include <grpcpp/impl/codegen/rpc_method.h>
#include <grpcpp/impl/codegen/service_type.h>
#include <grpcpp/impl/codegen/status.h>
#include <grpcpp/impl/codegen/stub_options.h>
#include <grpcpp/impl/codegen/sync_stream.h>

namespace grpc {
class CompletionQueue;
class Channel;
class ServerCompletionQueue;
class ServerContext;
}  // namespace grpc

namespace fastforward {

class FastForward final {
 public:
  static constexpr char const* service_full_name() {
    return "fastforward.FastForward";
  }
  class StubInterface {
   public:
    virtual ~StubInterface() {}
    std::unique_ptr< ::grpc::ClientReaderInterface< ::fastforward::FastForwardStream>> GetTsStream(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request) {
      return std::unique_ptr< ::grpc::ClientReaderInterface< ::fastforward::FastForwardStream>>(GetTsStreamRaw(context, request));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::fastforward::FastForwardStream>> AsyncGetTsStream(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::fastforward::FastForwardStream>>(AsyncGetTsStreamRaw(context, request, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::fastforward::FastForwardStream>> PrepareAsyncGetTsStream(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReaderInterface< ::fastforward::FastForwardStream>>(PrepareAsyncGetTsStreamRaw(context, request, cq));
    }
  private:
    virtual ::grpc::ClientReaderInterface< ::fastforward::FastForwardStream>* GetTsStreamRaw(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request) = 0;
    virtual ::grpc::ClientAsyncReaderInterface< ::fastforward::FastForwardStream>* AsyncGetTsStreamRaw(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request, ::grpc::CompletionQueue* cq, void* tag) = 0;
    virtual ::grpc::ClientAsyncReaderInterface< ::fastforward::FastForwardStream>* PrepareAsyncGetTsStreamRaw(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request, ::grpc::CompletionQueue* cq) = 0;
  };
  class Stub final : public StubInterface {
   public:
    Stub(const std::shared_ptr< ::grpc::ChannelInterface>& channel);
    std::unique_ptr< ::grpc::ClientReader< ::fastforward::FastForwardStream>> GetTsStream(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request) {
      return std::unique_ptr< ::grpc::ClientReader< ::fastforward::FastForwardStream>>(GetTsStreamRaw(context, request));
    }
    std::unique_ptr< ::grpc::ClientAsyncReader< ::fastforward::FastForwardStream>> AsyncGetTsStream(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request, ::grpc::CompletionQueue* cq, void* tag) {
      return std::unique_ptr< ::grpc::ClientAsyncReader< ::fastforward::FastForwardStream>>(AsyncGetTsStreamRaw(context, request, cq, tag));
    }
    std::unique_ptr< ::grpc::ClientAsyncReader< ::fastforward::FastForwardStream>> PrepareAsyncGetTsStream(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request, ::grpc::CompletionQueue* cq) {
      return std::unique_ptr< ::grpc::ClientAsyncReader< ::fastforward::FastForwardStream>>(PrepareAsyncGetTsStreamRaw(context, request, cq));
    }

   private:
    std::shared_ptr< ::grpc::ChannelInterface> channel_;
    ::grpc::ClientReader< ::fastforward::FastForwardStream>* GetTsStreamRaw(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request) override;
    ::grpc::ClientAsyncReader< ::fastforward::FastForwardStream>* AsyncGetTsStreamRaw(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request, ::grpc::CompletionQueue* cq, void* tag) override;
    ::grpc::ClientAsyncReader< ::fastforward::FastForwardStream>* PrepareAsyncGetTsStreamRaw(::grpc::ClientContext* context, const ::fastforward::FastForwardInfo& request, ::grpc::CompletionQueue* cq) override;
    const ::grpc::internal::RpcMethod rpcmethod_GetTsStream_;
  };
  static std::unique_ptr<Stub> NewStub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options = ::grpc::StubOptions());

  class Service : public ::grpc::Service {
   public:
    Service();
    virtual ~Service();
    virtual ::grpc::Status GetTsStream(::grpc::ServerContext* context, const ::fastforward::FastForwardInfo* request, ::grpc::ServerWriter< ::fastforward::FastForwardStream>* writer);
  };
  template <class BaseClass>
  class WithAsyncMethod_GetTsStream : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service *service) {}
   public:
    WithAsyncMethod_GetTsStream() {
      ::grpc::Service::MarkMethodAsync(0);
    }
    ~WithAsyncMethod_GetTsStream() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetTsStream(::grpc::ServerContext* context, const ::fastforward::FastForwardInfo* request, ::grpc::ServerWriter< ::fastforward::FastForwardStream>* writer) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestGetTsStream(::grpc::ServerContext* context, ::fastforward::FastForwardInfo* request, ::grpc::ServerAsyncWriter< ::fastforward::FastForwardStream>* writer, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncServerStreaming(0, context, request, writer, new_call_cq, notification_cq, tag);
    }
  };
  typedef WithAsyncMethod_GetTsStream<Service > AsyncService;
  template <class BaseClass>
  class WithGenericMethod_GetTsStream : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service *service) {}
   public:
    WithGenericMethod_GetTsStream() {
      ::grpc::Service::MarkMethodGeneric(0);
    }
    ~WithGenericMethod_GetTsStream() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetTsStream(::grpc::ServerContext* context, const ::fastforward::FastForwardInfo* request, ::grpc::ServerWriter< ::fastforward::FastForwardStream>* writer) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
  };
  template <class BaseClass>
  class WithRawMethod_GetTsStream : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service *service) {}
   public:
    WithRawMethod_GetTsStream() {
      ::grpc::Service::MarkMethodRaw(0);
    }
    ~WithRawMethod_GetTsStream() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable synchronous version of this method
    ::grpc::Status GetTsStream(::grpc::ServerContext* context, const ::fastforward::FastForwardInfo* request, ::grpc::ServerWriter< ::fastforward::FastForwardStream>* writer) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    void RequestGetTsStream(::grpc::ServerContext* context, ::grpc::ByteBuffer* request, ::grpc::ServerAsyncWriter< ::grpc::ByteBuffer>* writer, ::grpc::CompletionQueue* new_call_cq, ::grpc::ServerCompletionQueue* notification_cq, void *tag) {
      ::grpc::Service::RequestAsyncServerStreaming(0, context, request, writer, new_call_cq, notification_cq, tag);
    }
  };
  typedef Service StreamedUnaryService;
  template <class BaseClass>
  class WithSplitStreamingMethod_GetTsStream : public BaseClass {
   private:
    void BaseClassMustBeDerivedFromService(const Service *service) {}
   public:
    WithSplitStreamingMethod_GetTsStream() {
      ::grpc::Service::MarkMethodStreamed(0,
        new ::grpc::internal::SplitServerStreamingHandler< ::fastforward::FastForwardInfo, ::fastforward::FastForwardStream>(std::bind(&WithSplitStreamingMethod_GetTsStream<BaseClass>::StreamedGetTsStream, this, std::placeholders::_1, std::placeholders::_2)));
    }
    ~WithSplitStreamingMethod_GetTsStream() override {
      BaseClassMustBeDerivedFromService(this);
    }
    // disable regular version of this method
    ::grpc::Status GetTsStream(::grpc::ServerContext* context, const ::fastforward::FastForwardInfo* request, ::grpc::ServerWriter< ::fastforward::FastForwardStream>* writer) override {
      abort();
      return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
    }
    // replace default version of method with split streamed
    virtual ::grpc::Status StreamedGetTsStream(::grpc::ServerContext* context, ::grpc::ServerSplitStreamer< ::fastforward::FastForwardInfo,::fastforward::FastForwardStream>* server_split_streamer) = 0;
  };
  typedef WithSplitStreamingMethod_GetTsStream<Service > SplitStreamedService;
  typedef WithSplitStreamingMethod_GetTsStream<Service > StreamedService;
};

}  // namespace fastforward


#endif  // GRPC_fast_5fforward_2eproto__INCLUDED
